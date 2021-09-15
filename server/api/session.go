package api

import (
	"bytes"
	"fmt"
	"github.com/ergoapi/errors"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"next-terminal/pkg/constant"
	"next-terminal/pkg/global"
	"next-terminal/server/model"
	"next-terminal/server/utils"

	"github.com/pkg/sftp"
)

var (
	MIMEOctetStream = "application/octet-stream"
)

func SessionPagingEndpoint(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("pageIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	status := c.Query("status")
	userId := c.Query("userId")
	clientIp := c.Query("clientIp")
	assetId := c.Query("assetId")
	protocol := c.Query("protocol")

	items, total, err := sessionRepository.Find(pageIndex, pageSize, status, userId, clientIp, assetId, protocol)

	if err != nil {
		errors.Dangerous(err)
		return
	}

	for i := 0; i < len(items); i++ {
		if status == constant.Disconnected && len(items[i].Recording) > 0 {

			var recording string
			if items[i].Mode == constant.Naive {
				recording = items[i].Recording
			} else {
				recording = items[i].Recording + "/recording"
			}

			if utils.FileExists(recording) {
				items[i].Recording = "1"
			} else {
				items[i].Recording = "0"
			}
		} else {
			items[i].Recording = "0"
		}
	}

	Success(c, H{
		"total": total,
		"items": items,
	})
}

func SessionDeleteEndpoint(c *gin.Context) {
	sessionIds := c.Param("id")
	split := strings.Split(sessionIds, ",")
	err := sessionRepository.DeleteByIds(split)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, nil)
}

func SessionConnectEndpoint(c *gin.Context) {
	sessionId := c.Param("id")

	session := model.Session{}
	session.ID = sessionId
	session.Status = constant.Connected
	session.ConnectedTime = utils.NowJsonTime()

	if err := sessionRepository.UpdateById(&session, sessionId); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, nil)
}

func SessionDisconnectEndpoint(c *gin.Context) {
	sessionIds := c.Param("id")

	split := strings.Split(sessionIds, ",")
	for i := range split {
		CloseSessionById(split[i], ForcedDisconnect, "管理员强制关闭了此会话")
	}
	Success(c, nil)
}

var mutex sync.Mutex

func CloseSessionById(sessionId string, code int, reason string) {
	mutex.Lock()
	defer mutex.Unlock()
	observable, _ := global.Store.Get(sessionId)
	if observable != nil {
		zlog.Debug("会话%v创建者退出，原因：%v", sessionId, reason)
		observable.Subject.Close(code, reason)

		for i := 0; i < len(observable.Observers); i++ {
			observable.Observers[i].Close(code, reason)
			zlog.Debug("强制踢出会话%v的观察者", sessionId)
		}
	}
	global.Store.Del(sessionId)

	s, err := sessionRepository.FindById(sessionId)
	if err != nil {
		return
	}

	if s.Status == constant.Disconnected {
		return
	}

	if s.Status == constant.Connecting {
		// 会话还未建立成功，无需保留数据
		_ = sessionRepository.DeleteById(sessionId)
		return
	}

	session := model.Session{}
	session.ID = sessionId
	session.Status = constant.Disconnected
	session.DisconnectedTime = utils.NowJsonTime()
	session.Code = code
	session.Message = reason
	session.Password = "-"
	session.PrivateKey = "-"
	session.Passphrase = "-"

	_ = sessionRepository.UpdateById(&session, sessionId)
}

func SessionResizeEndpoint(c *gin.Context) {
	width := c.Query("width")
	height := c.Query("height")
	sessionId := c.Param("id")

	if len(width) == 0 || len(height) == 0 {
		panic("参数异常")
	}

	intWidth, _ := strconv.Atoi(width)

	intHeight, _ := strconv.Atoi(height)

	if err := sessionRepository.UpdateWindowSizeById(intWidth, intHeight, sessionId); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, "")
}

func SessionCreateEndpoint(c *gin.Context) {
	assetId := c.Query("assetId")
	mode := c.Query("mode")

	if mode == constant.Naive {
		mode = constant.Naive
	} else {
		mode = constant.Guacd
	}

	user, _ := GetCurrentAccount(c)

	if constant.TypeUser == user.Type {
		// 检测是否有访问权限
		assetIds, err := resourceSharerRepository.FindAssetIdsByUserId(user.ID)
		if err != nil {
			errors.Dangerous(err)
			return
		}

		if !utils.Contains(assetIds, assetId) {
			errors.Dangerous("您没有权限访问此资产")
			return
		}
	}

	asset, err := assetRepository.FindById(assetId)
	if err != nil {
		errors.Dangerous("您没有权限访问此资产")
		return
	}

	session := &model.Session{
		ID:         utils.UUID(),
		AssetId:    asset.ID,
		Username:   asset.Username,
		Password:   asset.Password,
		PrivateKey: asset.PrivateKey,
		Passphrase: asset.Passphrase,
		Protocol:   asset.Protocol,
		IP:         asset.IP,
		Port:       asset.Port,
		Status:     constant.NoConnect,
		Creator:    user.ID,
		ClientIP:   c.ClientIP(),
		Mode:       mode,
	}

	if asset.AccountType == "credential" {
		credential, err := credentialRepository.FindById(asset.CredentialId)
		if err != nil {
			errors.Dangerous(err)
			return
		}

		if credential.Type == constant.Custom {
			session.Username = credential.Username
			session.Password = credential.Password
		} else {
			session.Username = credential.Username
			session.PrivateKey = credential.PrivateKey
			session.Passphrase = credential.Passphrase
		}
	}

	if err := sessionRepository.Create(session); err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, map[string]interface{}{"id": session.ID})
}

func SessionUploadEndpoint(c *gin.Context) {
	sessionId := c.Param("id")
	session, err := sessionRepository.FindById(sessionId)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		errors.Dangerous(err)
		return
	}

	filename := file.Filename
	src, err := file.Open()
	if err != nil {
		errors.Dangerous(err)
		return
	}

	remoteDir := c.Query("dir")
	remoteFile := path.Join(remoteDir, filename)

	if "ssh" == session.Protocol {
		tun, ok := global.Store.Get(sessionId)
		if !ok {
			errors.Dangerous("获取sftp客户端失败")
			return
		}

		dstFile, err := tun.Subject.NextTerminal.SftpClient.Create(remoteFile)
		if err != nil {
			errors.Dangerous(err)
			return
		}
		defer dstFile.Close()

		buf := make([]byte, 1024)
		for {
			n, err := src.Read(buf)
			if err != nil {
				if err != io.EOF {
					zlog.Warn("文件上传错误 %v", err)
				} else {
					break
				}
			}
			_, _ = dstFile.Write(buf[:n])
		}
		Success(c, nil)
		return
	} else if "rdp" == session.Protocol {

		if strings.Contains(remoteFile, "../") {
			SafetyRuleTrigger(c)
			Fail(c, -1, ":) 您的IP已被记录，请去向管理员自首。")
			return
		}

		drivePath, err := configsRepository.GetDrivePath()
		if err != nil {
			errors.Dangerous(err)
			return
		}

		// Destination
		dst, err := os.Create(path.Join(drivePath, remoteFile))
		if err != nil {
			errors.Dangerous(err)
			return
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			errors.Dangerous(err)
			return
		}
		Success(c, nil)
		return
	}

	errors.Dangerous(err)
}

func SessionDownloadEndpoint(c *gin.Context) {
	sessionId := c.Param("id")
	session, err := sessionRepository.FindById(sessionId)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	//remoteDir := c.Query("dir")
	remoteFile := c.Query("file")
	// 获取带后缀的文件名称
	filenameWithSuffix := path.Base(remoteFile)
	if "ssh" == session.Protocol {
		tun, ok := global.Store.Get(sessionId)
		if !ok {
			errors.Dangerous("获取sftp客户端失败")
			return
		}

		dstFile, err := tun.Subject.NextTerminal.SftpClient.Open(remoteFile)
		if err != nil {
			errors.Dangerous(err)
			return
		}

		defer dstFile.Close()
		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filenameWithSuffix))

		var buff bytes.Buffer
		if _, err := dstFile.WriteTo(&buff); err != nil {
			errors.Dangerous(err)
			return
		}
		c.Data(http.StatusOK, MIMEOctetStream, buff.Bytes())
	} else if "rdp" == session.Protocol {
		if strings.Contains(remoteFile, "../") {
			SafetyRuleTrigger(c)
			Fail(c, -1, ":) 您的IP已被记录，请去向管理员自首。")
			return
		}
		drivePath, err := configsRepository.GetDrivePath()
		if err != nil {
			errors.Dangerous(err)
			return
		}
		c.FileAttachment(path.Join(drivePath, remoteFile), filenameWithSuffix)
		return
	}

	errors.Dangerous(err)
}

type File struct {
	Name    string         `json:"name"`
	Path    string         `json:"path"`
	IsDir   bool           `json:"isDir"`
	Mode    string         `json:"mode"`
	IsLink  bool           `json:"isLink"`
	ModTime utils.JsonTime `json:"modTime"`
	Size    int64          `json:"size"`
}

func SessionLsEndpoint(c *gin.Context) {
	sessionId := c.Param("id")
	session, err := sessionRepository.FindByIdAndDecrypt(sessionId)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	remoteDir := c.Query("dir")
	if "ssh" == session.Protocol {
		tun, ok := global.Store.Get(sessionId)
		if !ok {
			errors.Dangerous("获取sftp客户端失败")
			return
		}

		if tun.Subject.NextTerminal == nil {
			nextTerminal, err := CreateNextTerminalBySession(session)
			if err != nil {
				errors.Dangerous(err)
				return
			}
			tun.Subject.NextTerminal = nextTerminal
		}

		if tun.Subject.NextTerminal.SftpClient == nil {
			sftpClient, err := sftp.NewClient(tun.Subject.NextTerminal.SshClient)
			if err != nil {
				zlog.Error("创建sftp客户端失败：%v", err.Error())
				errors.Dangerous(err)
				return
			}
			tun.Subject.NextTerminal.SftpClient = sftpClient
		}

		fileInfos, err := tun.Subject.NextTerminal.SftpClient.ReadDir(remoteDir)
		if err != nil {
			errors.Dangerous(err)
			return
		}

		var files = make([]File, 0)
		for i := range fileInfos {

			// 忽略因此文件
			if strings.HasPrefix(fileInfos[i].Name(), ".") {
				continue
			}

			file := File{
				Name:    fileInfos[i].Name(),
				Path:    path.Join(remoteDir, fileInfos[i].Name()),
				IsDir:   fileInfos[i].IsDir(),
				Mode:    fileInfos[i].Mode().String(),
				IsLink:  fileInfos[i].Mode()&os.ModeSymlink == os.ModeSymlink,
				ModTime: utils.NewJsonTime(fileInfos[i].ModTime()),
				Size:    fileInfos[i].Size(),
			}

			files = append(files, file)
		}

		Success(c, files)
		return
	} else if "rdp" == session.Protocol {
		if strings.Contains(remoteDir, "../") {
			SafetyRuleTrigger(c)
			Fail(c, -1, ":) 您的IP已被记录，请去向管理员自首。")
			return
		}
		drivePath, err := configsRepository.GetDrivePath()
		if err != nil {
			errors.Dangerous(err)
			return
		}
		fileInfos, err := ioutil.ReadDir(path.Join(drivePath, remoteDir))
		if err != nil {
			errors.Dangerous(err)
			return
		}

		var files = make([]File, 0)
		for i := range fileInfos {
			file := File{
				Name:    fileInfos[i].Name(),
				Path:    path.Join(remoteDir, fileInfos[i].Name()),
				IsDir:   fileInfos[i].IsDir(),
				Mode:    fileInfos[i].Mode().String(),
				IsLink:  fileInfos[i].Mode()&os.ModeSymlink == os.ModeSymlink,
				ModTime: utils.NewJsonTime(fileInfos[i].ModTime()),
				Size:    fileInfos[i].Size(),
			}

			files = append(files, file)
		}

		Success(c, files)
		return
	}
	errors.Dangerous("当前协议不支持此操作")
}

func SafetyRuleTrigger(c *gin.Context) {
	zlog.Warn("IP %v 尝试进行攻击，请ban掉此IP", c.ClientIP())
	security := model.AccessSecurity{
		ID:     utils.UUID(),
		Source: "安全规则触发",
		IP:     c.ClientIP(),
		Rule:   constant.AccessRuleReject,
	}

	_ = accessSecurityRepository.Create(&security)
}

func SessionMkDirEndpoint(c *gin.Context) {
	sessionId := c.Param("id")
	session, err := sessionRepository.FindById(sessionId)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	remoteDir := c.Query("dir")
	if "ssh" == session.Protocol {
		tun, ok := global.Store.Get(sessionId)
		if !ok {
			errors.Dangerous("获取sftp客户端失败")
			return
		}
		if err := tun.Subject.NextTerminal.SftpClient.Mkdir(remoteDir); err != nil {
			errors.Dangerous(err)
			return
		}
		Success(c, nil)
		return
	} else if "rdp" == session.Protocol {
		if strings.Contains(remoteDir, "../") {
			SafetyRuleTrigger(c)
			Fail(c, -1, ":) 您的IP已被记录，请去向管理员自首。")
			return
		}
		drivePath, err := configsRepository.GetDrivePath()
		if err != nil {
			errors.Dangerous(err)
			return
		}

		if err := os.MkdirAll(path.Join(drivePath, remoteDir), os.ModePerm); err != nil {
			errors.Dangerous(err)
			return
		}
		Success(c, nil)
		return
	}

	errors.Dangerous("当前协议不支持此操作")
}

func SessionRmEndpoint(c *gin.Context) {
	sessionId := c.Param("id")
	session, err := sessionRepository.FindById(sessionId)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	key := c.Query("key")
	if "ssh" == session.Protocol {
		tun, ok := global.Store.Get(sessionId)
		if !ok {
			errors.Dangerous("获取sftp客户端失败")
			return
		}

		sftpClient := tun.Subject.NextTerminal.SftpClient

		stat, err := sftpClient.Stat(key)
		if err != nil {
			errors.Dangerous(err)
			return
		}

		if stat.IsDir() {
			fileInfos, err := sftpClient.ReadDir(key)
			if err != nil {
				errors.Dangerous(err)
				return
			}

			for i := range fileInfos {
				if err := sftpClient.Remove(path.Join(key, fileInfos[i].Name())); err != nil {
					errors.Dangerous(err)
					return
				}
			}

			if err := sftpClient.RemoveDirectory(key); err != nil {
				errors.Dangerous(err)
				return
			}
		} else {
			if err := sftpClient.Remove(key); err != nil {
				errors.Dangerous(err)
				return
			}
		}

		Success(c, nil)
		return
	} else if "rdp" == session.Protocol {
		if strings.Contains(key, "../") {
			SafetyRuleTrigger(c)
			Fail(c, -1, ":) 您的IP已被记录，请去向管理员自首。")
			return
		}
		drivePath, err := configsRepository.GetDrivePath()
		if err != nil {
			errors.Dangerous(err)
			return
		}

		if err := os.RemoveAll(path.Join(drivePath, key)); err != nil {
			errors.Dangerous(err)
			return
		}

		Success(c, nil)
		return
	}

	errors.Dangerous("当前协议不支持此操作")
}

func SessionRenameEndpoint(c *gin.Context) {
	sessionId := c.Param("id")
	session, err := sessionRepository.FindById(sessionId)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	oldName := c.Query("oldName")
	newName := c.Query("newName")
	if "ssh" == session.Protocol {
		tun, ok := global.Store.Get(sessionId)
		if !ok {
			errors.Dangerous("获取sftp客户端失败")
			return
		}

		sftpClient := tun.Subject.NextTerminal.SftpClient

		if err := sftpClient.Rename(oldName, newName); err != nil {
			errors.Dangerous(err)
			return
		}

		Success(c, nil)
		return
	} else if "rdp" == session.Protocol {
		if strings.Contains(oldName, "../") {
			SafetyRuleTrigger(c)
			Fail(c, -1, ":) 您的IP已被记录，请去向管理员自首。")
			return
		}
		drivePath, err := configsRepository.GetDrivePath()
		if err != nil {
			errors.Dangerous(err)
			return
		}

		if err := os.Rename(path.Join(drivePath, oldName), path.Join(drivePath, newName)); err != nil {
			errors.Dangerous(err)
			return
		}

		Success(c, nil)
		return
	}
	errors.Dangerous("当前协议不支持此操作")
}

func SessionRecordingEndpoint(c *gin.Context) {
	sessionId := c.Param("id")
	session, err := sessionRepository.FindById(sessionId)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	var recording string
	if session.Mode == constant.Naive {
		recording = session.Recording
	} else {
		recording = session.Recording + "/recording"
	}

	zlog.Debug("读取录屏文件：%v,是否存在: %v, 是否为文件: %v", recording, utils.FileExists(recording), utils.IsFile(recording))
	c.File(recording)
}
