package api

import (
	"encoding/base64"
	"fmt"
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/ergoapi/util/zos"
	"github.com/gin-gonic/gin"
	"next-terminal/constants"
	"next-terminal/models"
	"next-terminal/pkg/utils"
	"strconv"
	"strings"
)

func CredentialAllEndpoint(c *gin.Context) {
	account, _ := GetCurrentAccount(c)
	items, _ := credentialRepository.FindByUser(account)
	exgin.GinsData(c, items, nil)
}
func CredentialCreateEndpoint(c *gin.Context) {
	var item models.Credential
	exgin.Bind(c, &item)

	account, _ := GetCurrentAccount(c)
	item.Owner = account.ID
	item.ID = zos.GenUUID()

	switch item.Type {
	case constants.CredentialCustom:
		item.PrivateKey = "-"
		item.Passphrase = "-"
		if item.Username == "" {
			item.Username = "-"
		}
		if item.Password == "" {
			item.Password = "-"
		}
	case constants.CredentialPrivateKey:
		item.Password = "-"
		if item.Username == "" {
			item.Username = "-"
		}
		if item.PrivateKey == "" {
			item.PrivateKey = "-"
		}
		if item.Passphrase == "" {
			item.Passphrase = "-"
		}
	case constants.CredentialAccessKey:
		item.PrivateKey = "-"
		item.Passphrase = "-"
		item.Password = "-"
		item.Username = "-"
	default:
		exgin.GinsData(c, nil, fmt.Errorf("类型错误，暂不支持: %v", item.Type))
		return
	}

	item.Encrypted = true
	if err := credentialRepository.Create(&item); err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, item, nil)
}

func CredentialPagingEndpoint(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("pageIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	name := c.Query("name")

	order := c.Query("order")
	field := c.Query("field")

	account, _ := GetCurrentAccount(c)
	items, total, err := credentialRepository.Find(pageIndex, pageSize, name, order, field, account)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, H{
		"total": total,
		"items": items,
	}, nil)
}

func CredentialUpdateEndpoint(c *gin.Context) {
	id := c.Param("id")

	if err := PreCheckCredentialPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	var item models.Credential
	exgin.Bind(c, &item)

	switch item.Type {
	case constants.CredentialCustom:
		item.PrivateKey = "-"
		item.Passphrase = "-"
		if item.Username == "" {
			item.Username = "-"
		}
		if item.Password == "" {
			item.Password = "-"
		}
		if item.Password != "-" {
			encryptedCBC, err := utils.AesEncryptCBC([]byte(item.Password), utils.Encryption())
			if err != nil {
				errors.Dangerous(err)
				return
			}
			item.Password = base64.StdEncoding.EncodeToString(encryptedCBC)
		}
	case constants.CredentialPrivateKey:
		item.Password = "-"
		if item.Username == "" {
			item.Username = "-"
		}
		if item.PrivateKey == "" {
			item.PrivateKey = "-"
		}
		if item.PrivateKey != "-" {
			encryptedCBC, err := utils.AesEncryptCBC([]byte(item.PrivateKey), utils.Encryption())
			if err != nil {
				errors.Dangerous(err)
				return
			}
			item.PrivateKey = base64.StdEncoding.EncodeToString(encryptedCBC)
		}
		if item.Passphrase == "" {
			item.Passphrase = "-"
		}
		if item.Passphrase != "-" {
			encryptedCBC, err := utils.AesEncryptCBC([]byte(item.Passphrase), utils.Encryption())
			if err != nil {
				errors.Dangerous(err)
				return
			}
			item.Passphrase = base64.StdEncoding.EncodeToString(encryptedCBC)
		}
	case constants.CredentialAccessKey:
		item.PrivateKey = "-"
		item.Passphrase = "-"
		item.Password = "-"
		item.Username = "-"
	default:
		exgin.GinsData(c, nil, fmt.Errorf("类型错误，暂不支持: %v", item.Type))
		return
	}
	item.Encrypted = true

	if err := credentialRepository.UpdateById(&item, id); err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, nil, nil)
}

func CredentialDeleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	split := strings.Split(id, ",")
	for i := range split {
		if err := PreCheckCredentialPermission(c, split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
		if err := credentialRepository.DeleteById(split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
		// 删除资产与用户的关系
		if err := resourceSharerRepository.DeleteResourceSharerByResourceId(split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
	}

	exgin.GinsData(c, nil, nil)
}

func CredentialGetEndpoint(c *gin.Context) {
	id := c.Param("id")
	if err := PreCheckCredentialPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	item, err := credentialRepository.FindByIdAndDecrypt(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	if !HasPermission(c, item.Owner) {
		errors.Dangerous("permission denied")
		return
	}

	exgin.GinsData(c, item, nil)
}

func CredentialChangeOwnerEndpoint(c *gin.Context) {
	id := c.Param("id")

	if err := PreCheckCredentialPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	owner := c.Query("owner")
	if err := credentialRepository.UpdateById(&models.Credential{Owner: owner}, id); err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, "", nil)
}

func PreCheckCredentialPermission(c *gin.Context, id string) error {
	item, err := credentialRepository.FindById(id)
	if err != nil {
		return err
	}

	if !HasPermission(c, item.Owner) {
		return fmt.Errorf("permission denied")
	}
	return nil
}
