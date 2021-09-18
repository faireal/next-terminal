package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	"next-terminal/constants"
	"next-terminal/models"
	"path"
	"strconv"

	"github.com/gorilla/websocket"
	"next-terminal/pkg/global"
	"next-terminal/pkg/guacd"
)

const (
	TunnelClosed     int = -1
	Normal           int = 0
	NotFoundSession  int = 800
	NewTunnelError   int = 801
	ForcedDisconnect int = 802
)

func TunEndpoint(c *gin.Context) {

	ws, err := UpGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Error("升级为WebSocket协议失败：%v", err.Error())
		errors.Dangerous(err)
		return
	}

	width := c.Query("width")
	height := c.Query("height")
	dpi := c.Query("dpi")
	sessionId := c.Query("sessionId")
	connectionId := c.Query("connectionId")

	intWidth, _ := strconv.Atoi(width)
	intHeight, _ := strconv.Atoi(height)

	configuration := guacd.NewConfiguration()

	propertyMap := configsRepository.FindAllMap()

	var session models.Session

	if len(connectionId) > 0 {
		session, err = sessionRepository.FindByConnectionId(connectionId)
		if err != nil {
			zlog.Warn("会话不存在")
			errors.Dangerous(err)
			return
		}
		if session.Status != constants.Connected {
			zlog.Warn("会话未在线")
			errors.Dangerous("会话未在线")
			return
		}
		configuration.ConnectionID = connectionId
		sessionId = session.ID
		configuration.SetParameter("width", strconv.Itoa(session.Width))
		configuration.SetParameter("height", strconv.Itoa(session.Height))
		configuration.SetParameter("dpi", "96")
	} else {
		configuration.SetParameter("width", width)
		configuration.SetParameter("height", height)
		configuration.SetParameter("dpi", dpi)
		session, err = sessionRepository.FindByIdAndDecrypt(sessionId)
		if err != nil {
			CloseSessionById(sessionId, NotFoundSession, "会话不存在")
			errors.Dangerous("会话不存在")
			return
		}

		if propertyMap[guacd.EnableRecording] == "true" {
			configuration.SetParameter(guacd.RecordingPath, path.Join(propertyMap[guacd.RecordingPath], sessionId))
			configuration.SetParameter(guacd.CreateRecordingPath, propertyMap[guacd.CreateRecordingPath])
		} else {
			configuration.SetParameter(guacd.RecordingPath, "")
		}

		configuration.Protocol = session.Protocol
		switch configuration.Protocol {
		case "rdp":
			configuration.SetParameter("username", session.Username)
			configuration.SetParameter("password", session.Password)

			configuration.SetParameter("security", "any")
			configuration.SetParameter("ignore-cert", "true")
			configuration.SetParameter("create-drive-path", "true")
			configuration.SetParameter("resize-method", "reconnect")
			configuration.SetParameter(guacd.EnableDrive, propertyMap[guacd.EnableDrive])
			configuration.SetParameter(guacd.DriveName, propertyMap[guacd.DriveName])
			configuration.SetParameter(guacd.DrivePath, propertyMap[guacd.DrivePath])
			configuration.SetParameter(guacd.EnableWallpaper, propertyMap[guacd.EnableWallpaper])
			configuration.SetParameter(guacd.EnableTheming, propertyMap[guacd.EnableTheming])
			configuration.SetParameter(guacd.EnableFontSmoothing, propertyMap[guacd.EnableFontSmoothing])
			configuration.SetParameter(guacd.EnableFullWindowDrag, propertyMap[guacd.EnableFullWindowDrag])
			configuration.SetParameter(guacd.EnableDesktopComposition, propertyMap[guacd.EnableDesktopComposition])
			configuration.SetParameter(guacd.EnableMenuAnimations, propertyMap[guacd.EnableMenuAnimations])
			configuration.SetParameter(guacd.DisableBitmapCaching, propertyMap[guacd.DisableBitmapCaching])
			configuration.SetParameter(guacd.DisableOffscreenCaching, propertyMap[guacd.DisableOffscreenCaching])
			configuration.SetParameter(guacd.DisableGlyphCaching, propertyMap[guacd.DisableGlyphCaching])
		case "ssh":
			if len(session.PrivateKey) > 0 && session.PrivateKey != "-" {
				configuration.SetParameter("username", session.Username)
				configuration.SetParameter("private-key", session.PrivateKey)
				configuration.SetParameter("passphrase", session.Passphrase)
			} else {
				configuration.SetParameter("username", session.Username)
				configuration.SetParameter("password", session.Password)
			}

			configuration.SetParameter(guacd.FontSize, propertyMap[guacd.FontSize])
			configuration.SetParameter(guacd.FontName, propertyMap[guacd.FontName])
			configuration.SetParameter(guacd.ColorScheme, propertyMap[guacd.ColorScheme])
			configuration.SetParameter(guacd.Backspace, propertyMap[guacd.Backspace])
			configuration.SetParameter(guacd.TerminalType, propertyMap[guacd.TerminalType])
		case "vnc":
			configuration.SetParameter("username", session.Username)
			configuration.SetParameter("password", session.Password)
		case "telnet":
			configuration.SetParameter("username", session.Username)
			configuration.SetParameter("password", session.Password)

			configuration.SetParameter(guacd.FontSize, propertyMap[guacd.FontSize])
			configuration.SetParameter(guacd.FontName, propertyMap[guacd.FontName])
			configuration.SetParameter(guacd.ColorScheme, propertyMap[guacd.ColorScheme])
			configuration.SetParameter(guacd.Backspace, propertyMap[guacd.Backspace])
			configuration.SetParameter(guacd.TerminalType, propertyMap[guacd.TerminalType])
		case "kubernetes":

			configuration.SetParameter(guacd.FontSize, propertyMap[guacd.FontSize])
			configuration.SetParameter(guacd.FontName, propertyMap[guacd.FontName])
			configuration.SetParameter(guacd.ColorScheme, propertyMap[guacd.ColorScheme])
			configuration.SetParameter(guacd.Backspace, propertyMap[guacd.Backspace])
			configuration.SetParameter(guacd.TerminalType, propertyMap[guacd.TerminalType])
		default:
			zlog.Error("UnSupport Protocol: %v", configuration.Protocol)
			Fail(c, 400, "不支持的协议")
			return
		}

		configuration.SetParameter("hostname", session.IP)
		configuration.SetParameter("port", strconv.Itoa(session.Port))

		// 加载资产配置的属性，优先级比全局配置的高，因此最后加载，覆盖掉全局配置
		attributes, _ := assetRepository.FindAttrById(session.AssetId)
		if len(attributes) > 0 {
			for i := range attributes {
				attribute := attributes[i]
				configuration.SetParameter(attribute.Name, attribute.Value)
			}
		}
	}
	for name := range configuration.Parameters {
		// 替换数据库空格字符串占位符为真正的空格
		if configuration.Parameters[name] == "-" {
			configuration.Parameters[name] = ""
		}
	}

	addr := propertyMap[guacd.Host] + ":" + propertyMap[guacd.Port]

	tunnel, err := guacd.NewTunnel(addr, configuration)
	if err != nil {
		if connectionId == "" {
			CloseSessionById(sessionId, NewTunnelError, err.Error())
		}
		zlog.Error("建立连接失败: %v", err.Error())
		errors.Dangerous(err)
		return
	}

	tun := global.Tun{
		Protocol:  session.Protocol,
		Mode:      session.Mode,
		WebSocket: ws,
		Tunnel:    tunnel,
	}

	if len(session.ConnectionId) == 0 {

		var observers []global.Tun
		observable := global.Observable{
			Subject:   &tun,
			Observers: observers,
		}

		global.Store.Set(sessionId, &observable)

		sess := models.Session{
			ConnectionId: tunnel.UUID,
			Width:        intWidth,
			Height:       intHeight,
			Status:       constants.Connecting,
			Recording:    configuration.GetParameter(guacd.RecordingPath),
		}
		// 创建新会话
		zlog.Debug("创建新会话 %v", sess.ConnectionId)
		if err := sessionRepository.UpdateById(&sess, sessionId); err != nil {
			errors.Dangerous(err)
			return
		}
	} else {
		// 监控会话
		observable, ok := global.Store.Get(sessionId)
		if ok {
			observers := append(observable.Observers, tun)
			observable.Observers = observers
			global.Store.Set(sessionId, observable)
			zlog.Debug("加入会话%v,当前观察者数量为：%v", session.ConnectionId, len(observers))
		}
	}

	go func() {
		for {
			instruction, err := tunnel.Read()
			if err != nil {
				if connectionId == "" {
					CloseSessionById(sessionId, TunnelClosed, "远程连接关闭")
				}
				break
			}
			if len(instruction) == 0 {
				continue
			}
			err = ws.WriteMessage(websocket.TextMessage, instruction)
			if err != nil {
				if connectionId == "" {
					CloseSessionById(sessionId, Normal, "正常退出")
				}
				break
			}
		}
	}()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if connectionId == "" {
				CloseSessionById(sessionId, Normal, "正常退出")
			}
			break
		}
		_, err = tunnel.WriteAndFlush(message)
		if err != nil {
			if connectionId == "" {
				CloseSessionById(sessionId, TunnelClosed, "远程连接关闭")
			}
			break
		}
	}
	errors.Dangerous(err)
}
