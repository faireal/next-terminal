package api

import (
	"fmt"
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/ergoapi/util/color"
	"github.com/ergoapi/util/file"
	"github.com/ergoapi/util/ztime"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	"net"
	"net/http/httputil"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"next-terminal/pkg/constant"
	"next-terminal/pkg/global"
	"next-terminal/server/utils"
)

// ExLog exlog middleware
func ExLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		if len(query) == 0 {
			query = " - "
		}
		if latency > time.Second*1 {
			zlog.Warn("[msg] api %v query %v", path, latency)
		}
		if len(c.Errors) > 0 || c.Writer.Status() >= 500 {
			msg := fmt.Sprintf("requestid %v => %v | %v | %v | %v | %v | %v <= err: %v", exgin.GetRID(c), color.SRed("%v", c.Writer.Status()), c.ClientIP(), c.Request.Method, path, query, latency, c.Errors.String())
			zlog.Warn(msg)
			go file.Writefile(fmt.Sprintf("/tmp/%v.errreq.txt", ztime.GetToday()), msg)
		} else {
			zlog.Info("requestid %v => %v | %v | %v | %v | %v | %v ", exgin.GetRID(c), color.SGreen("%v", c.Writer.Status()), c.ClientIP(), c.Request.Method, path, query, latency)
		}
	}
}

// ErrorHandler recovery
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if res, ok := err.(errors.ErgoError); ok {
					Fail(c, 400, res.Message)
					return
				}
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					zlog.Error("Recovery from brokenPipe ---> path: %v, err: %v, request: %v", c.Request.URL.Path, err, string(httpRequest))
					Fail(c, 10500, "请求broken")
				} else {
					zlog.Error("Recovery from panic ---> err: %v, request: %v, stack: %v", err, string(httpRequest), string(debug.Stack()))
					Fail(c, 10500, "请求panic")
				}
				return
			}
		}()
		c.Next()
	}
}

func TcpWall() gin.HandlerFunc {

	return func(c *gin.Context) {

		if global.Securities == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		for i := 0; i < len(global.Securities); i++ {
			security := global.Securities[i]

			if strings.Contains(security.IP, "/") {
				// CIDR
				_, ipNet, err := net.ParseCIDR(security.IP)
				if err != nil {
					continue
				}
				if !ipNet.Contains(net.ParseIP(ip)) {
					continue
				}
			} else if strings.Contains(security.IP, "-") {
				// 范围段
				split := strings.Split(security.IP, "-")
				if len(split) < 2 {
					continue
				}
				start := split[0]
				end := split[1]
				intReqIP := utils.IpToInt(ip)
				if intReqIP < utils.IpToInt(start) || intReqIP > utils.IpToInt(end) {
					continue
				}
			} else {
				// IP
				if security.IP != ip {
					continue
				}
			}

			if security.Rule == constant.AccessRuleAllow {
				c.Next()
				return
			}
			if security.Rule == constant.AccessRuleReject {
				if c.Request.Header.Get("X-Requested-With") != "" || c.Request.Header.Get(Token) != "" {
					Fail(c, 0, "您的访问请求被拒绝 :(")
					return
				} else {
					c.String(666, "您的访问请求被拒绝 :(")
					return
				}
			}
		}

		c.Next()
	}
}

func Auth() gin.HandlerFunc {

	startWithUrls := []string{"/login", "/static", "/favicon.ico", "/logo.svg", "/asciinema"}

	download := regexp.MustCompile(`^/sessions/\w{8}(-\w{4}){3}-\w{12}/download`)
	recording := regexp.MustCompile(`^/sessions/\w{8}(-\w{4}){3}-\w{12}/recording`)

	return func(c *gin.Context) {

		uri := c.Request.RequestURI
		if uri == "/" || strings.HasPrefix(uri, "/#") {
			c.Next()
			return
		}
		// 路由拦截 - 登录身份、资源权限判断等
		for i := range startWithUrls {
			if strings.HasPrefix(uri, startWithUrls[i]) {
				c.Next()
				return
			}
		}

		if download.FindString(uri) != "" {
			c.Next()
			return
		}

		if recording.FindString(uri) != "" {
			c.Next()
			return
		}

		token := GetToken(c)
		cacheKey := BuildCacheKeyByToken(token)
		authorization, found := global.Cache.Get(cacheKey)
		if !found {
			Fail(c, 401, "您的登录信息已失效，请重新登录后再试。")
			return
		}

		if authorization.(Authorization).Remember {
			// 记住登录有效期两周
			global.Cache.Set(cacheKey, authorization, time.Hour*time.Duration(24*14))
		} else {
			global.Cache.Set(cacheKey, authorization, time.Hour*time.Duration(2))
		}

		c.Next()
	}
}

func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {

		account, found := GetCurrentAccount(c)
		if !found {
			Fail(c, 401, "您的登录信息已失效，请重新登录后再试。")
			return
		}

		if account.Type != constant.TypeAdmin {
			Fail(c, 403, "permission denied")
			return
		}
		c.Next()
	}
}
