package api

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"next-terminal/constants"
	"next-terminal/models"
	"next-terminal/pkg/utils"
	"strconv"
	"strings"
)

func AssetCreateEndpoint(c *gin.Context) {
	m := map[string]interface{}{}
	exgin.Bind(c, &m)

	data, _ := json.Marshal(m)
	var item models.Asset
	if err := json.Unmarshal(data, &item); err != nil {
		errors.Dangerous(err)
		return
	}

	account, _ := GetCurrentAccount(c)
	item.Owner = account.ID
	item.ID = utils.UUID()
	item.Created = utils.NowJsonTime()

	if err := assetRepository.InitAsset(&item, m); err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, item)
}

func AssetImportEndpoint(c *gin.Context) {
	account, _ := GetCurrentAccount(c)

	file, err := c.FormFile("file")
	if err != nil {
		errors.Dangerous(err)
		return
	}

	src, err := file.Open()
	if err != nil {
		errors.Dangerous(err)
		return
	}

	defer src.Close()
	reader := csv.NewReader(bufio.NewReader(src))
	records, err := reader.ReadAll()
	if err != nil {
		errors.Dangerous(err)
		return
	}

	total := len(records)
	if total == 0 {
		errors.Dangerous("csv数据为空")
		return
	}

	var successCount = 0
	var errorCount = 0
	m := map[string]interface{}{}

	for i := 0; i < total; i++ {
		record := records[i]
		if len(record) >= 9 {
			port, _ := strconv.Atoi(record[3])
			asset := models.Asset{
				ID:          utils.UUID(),
				Name:        record[0],
				Protocol:    record[1],
				IP:          record[2],
				Port:        port,
				AccountType: constants.Custom,
				Username:    record[4],
				Password:    record[5],
				PrivateKey:  record[6],
				Passphrase:  record[7],
				Description: record[8],
				Created:     utils.NowJsonTime(),
				Owner:       account.ID,
			}

			if len(record) >= 10 {
				tags := strings.ReplaceAll(record[9], "|", ",")
				asset.Tags = tags
			}

			err := assetRepository.Create(&asset)
			if err != nil {
				errorCount++
				m[strconv.Itoa(i)] = err.Error()
			} else {
				successCount++
				// 创建后自动检测资产是否存活
				go func() {
					active := utils.Tcping(asset.IP, asset.Port)
					_ = assetRepository.UpdateActiveById(active, asset.ID)
				}()
			}
		}
	}

	Success(c, map[string]interface{}{
		"successCount": successCount,
		"errorCount":   errorCount,
		"data":         m,
	})
}

func AssetPagingEndpoint(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("pageIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	name := c.Query("name")
	protocol := c.Query("protocol")
	tags := c.Query("tags")
	owner := c.Query("owner")
	sharer := c.Query("sharer")
	userGroupId := c.Query("userGroupId")
	ip := c.Query("ip")

	order := c.Query("order")
	field := c.Query("field")

	account, _ := GetCurrentAccount(c)
	items, total, err := assetRepository.Find(pageIndex, pageSize, name, protocol, tags, account, owner, sharer, userGroupId, ip, order, field)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, H{
		"total": total,
		"items": items,
	})
}

func AssetAllEndpoint(c *gin.Context) {
	protocol := c.Query("protocol")
	account, _ := GetCurrentAccount(c)
	items, _ := assetRepository.FindByProtocolAndUser(protocol, account)
	Success(c, items)
}

func AssetUpdateEndpoint(c *gin.Context) {
	id := c.Param("id")
	if err := PreCheckAssetPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	m := map[string]interface{}{}
	exgin.Bind(c, &m)

	data, _ := json.Marshal(m)
	var item models.Asset
	if err := json.Unmarshal(data, &item); err != nil {
		errors.Dangerous(err)
		return
	}

	switch item.AccountType {
	case "credential":
		item.Username = "-"
		item.Password = "-"
		item.PrivateKey = "-"
		item.Passphrase = "-"
	case "private-key":
		item.Password = "-"
		item.CredentialId = "-"
		if len(item.Username) == 0 {
			item.Username = "-"
		}
		if len(item.Passphrase) == 0 {
			item.Passphrase = "-"
		}
	case "custom":
		item.PrivateKey = "-"
		item.Passphrase = "-"
		item.CredentialId = "-"
	}

	if len(item.Tags) == 0 {
		item.Tags = "-"
	}

	if item.Description == "" {
		item.Description = "-"
	}

	if err := assetRepository.Encrypt(&item, utils.Encryption()); err != nil {
		errors.Dangerous(err)
		return
	}
	if err := assetRepository.UpdateById(&item, id); err != nil {
		errors.Dangerous(err)
		return
	}
	if err := assetRepository.UpdateAttributes(id, item.Protocol, m); err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, nil)
}

func AssetDeleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	split := strings.Split(id, ",")
	for i := range split {
		if err := PreCheckAssetPermission(c, split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
		if err := assetRepository.DeleteById(split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
		// 删除资产与用户的关系
		if err := resourceSharerRepository.DeleteResourceSharerByResourceId(split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
	}

	Success(c, nil)
}

func AssetGetEndpoint(c *gin.Context) {
	id := c.Param("id")
	if err := PreCheckAssetPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	var item models.Asset
	var err error
	if item, err = assetRepository.FindByIdAndDecrypt(id); err != nil {
		errors.Dangerous(err)
		return
	}
	attributeMap, err := assetRepository.FindAssetAttrMapByAssetId(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	itemMap := utils.StructToMap(item)
	for key := range attributeMap {
		itemMap[key] = attributeMap[key]
	}

	Success(c, itemMap)
}

func AssetTcpingEndpoint(c *gin.Context) {
	id := c.Param("id")

	var item models.Asset
	var err error
	if item, err = assetRepository.FindById(id); err != nil {
		errors.Dangerous(err)
		return
	}

	active := utils.Tcping(item.IP, item.Port)

	if item.Active != active {
		if err := assetRepository.UpdateActiveById(active, item.ID); err != nil {
			errors.Dangerous(err)
			return
		}
	}

	Success(c, active)
}

func AssetTagsEndpoint(c *gin.Context) {
	var items []string
	var err error
	if items, err = assetRepository.FindTags(); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, items)
}

func AssetChangeOwnerEndpoint(c *gin.Context) {
	id := c.Param("id")

	if err := PreCheckAssetPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	owner := c.Query("owner")
	if err := assetRepository.UpdateById(&models.Asset{Owner: owner}, id); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, "")
}

func PreCheckAssetPermission(c *gin.Context, id string) error {
	item, err := assetRepository.FindById(id)
	if err != nil {
		return err
	}

	if !HasPermission(c, item.Owner) {
		return fmt.Errorf("permission denied")
	}
	return nil
}
