package api

import (
	"context"
	"fmt"
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"gopkg.in/guregu/null.v3"
	"gorm.io/gorm"
	"net/http"
	"next-terminal/constants"
	"next-terminal/models"
	"next-terminal/pkg/ldap"
	"next-terminal/pkg/utils"
	"strings"
	"time"

	"next-terminal/pkg/totp"

	ntcache "next-terminal/pkg/cache"
	GitHubAPI "github.com/google/go-github/github"
	GitHubOauth2 "golang.org/x/oauth2/github"

)

const (
	RememberEffectiveTime    = time.Hour * time.Duration(24*14)
	NotRememberEffectiveTime = time.Hour * time.Duration(2)
)

type LoginAccount struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
	TOTP     string `json:"totp"`
}

type ConfirmTOTP struct {
	Secret string `json:"secret"`
	TOTP   string `json:"totp"`
}

type ChangePassword struct {
	NewPassword string `json:"newPassword"`
	OldPassword string `json:"oldPassword"`
}

type Authorization struct {
	Token    string
	Remember bool
	User     models.User
}

// loginfail 默认密码校验次数
func loginfail() int {
	defaultnum := viper.GetInt("core.login.trynum")
	if defaultnum > 0 {
		return defaultnum
	}
	return 3
}

func LoginEndpoint(c *gin.Context) {
	var loginAccount LoginAccount
	exgin.Bind(c, &loginAccount)
	user, err := userRepository.FindByUsername(loginAccount.Username)

	// 存储登录失败次数信息
	loginFailCountKey := loginAccount.Username
	v, ok := ntcache.MemCache.Get(loginFailCountKey)
	if !ok {
		v = 1
	}
	count := v.(int)
	if count >= loginfail() {
		Fail(c, -1, "登录失败次数过多，请稍后再试")
		return
	}

	if err != nil {
		count++
		ntcache.MemCache.Set(loginFailCountKey, count, time.Minute*time.Duration(5))
		FailWithData(c, -1, "您输入的账号或密码不正确", count)
		return
	}

	if err := utils.Encoder.Match([]byte(user.Password), []byte(loginAccount.Password)); err != nil {
		count++
		ntcache.MemCache.Set(loginFailCountKey, count, time.Minute*time.Duration(5))
		FailWithData(c, -1, "您输入的账号或密码不正确", count)
		return
	}

	if user.TOTPSecret != "" && user.TOTPSecret != "-" {
		Fail(c, 0, "")
		return
	}

	if user.Baned {
		Fail(c, -1, "当前账号已被禁用")
		return
	}

	token, err := LoginSuccess(c, loginAccount, user, "pass")
	if err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, token)
}

func LdapLoginEndpoint(c *gin.Context) {
	var loginAccount LoginAccount
	exgin.Bind(c, &loginAccount)

	// 存储登录失败次数信息
	loginFailCountKey := loginAccount.Username
	v, ok := ntcache.MemCache.Get(loginFailCountKey)
	if !ok {
		v = 1
	}
	count := v.(int)
	if count >= loginfail() {
		Fail(c, -1, "登录失败次数过多，请稍后再试")
		return
	}

	ldapuser, err := ldap.LdapLogin(loginAccount.Username, loginAccount.Password)
	if err != nil {
		count++
		ntcache.MemCache.Set(loginFailCountKey, count, time.Minute*time.Duration(5))
		FailWithData(c, -1, fmt.Sprintf("您输入的账号或密码不正确: %v", err), count)
		return
	}

	u, err := userRepository.UserGet("username = ?", ldapuser.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		errors.Dangerous(err)
		return
	}

	if err == gorm.ErrRecordNotFound {
		// 创建用户
		u = &models.User{
			ID:         utils.UUID(),
			Username:   ldapuser.Username,
			Nickname:   ldapuser.Nickname,
			Role:       constants.RoleDefault,
			Department: ldapuser.Department,
			Mode:       "ldap",
			Mail:       ldapuser.Mail,
		}
		userRepository.Create(u)
	} else {
		u.Department = ldapuser.Department
		u.Nickname = ldapuser.Nickname
		u.Mail = ldapuser.Mail
		u.Mode = "ldap"
	}

	if u.TOTPSecret != "" && u.TOTPSecret != "-" {
		Fail(c, 0, "")
		return
	}

	if u.Baned {
		Fail(c, -1, "当前账号已被禁用")
		return
	}

	token, err := LoginSuccess(c, loginAccount, u, "ldap")
	if err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, token)
}

func LoginSuccess(c *gin.Context, loginAccount LoginAccount, user *models.User, logintype string) (token string, err error) {
	token = strings.Join([]string{utils.UUID(), utils.UUID(), utils.UUID(), utils.UUID()}, "")

	authorization := Authorization{
		Token:    token,
		Remember: loginAccount.Remember,
		User:     *user,
	}

	cacheKey := BuildCacheKeyByToken(token)

	if authorization.Remember {
		// 记住登录有效期两周
		ntcache.MemCache.Set(cacheKey, authorization, RememberEffectiveTime)
	} else {
		ntcache.MemCache.Set(cacheKey, authorization, NotRememberEffectiveTime)
	}

	// 保存登录日志
	loginLog := models.LoginLog{
		ID:              token,
		UserId:          user.ID,
		ClientIP:        c.ClientIP(),
		ClientUserAgent: c.Request.UserAgent(),
		LoginTime:       null.TimeFrom(time.Now()),
		Remember:        authorization.Remember,
	}

	if loginLogRepository.Create(&loginLog) != nil {
		return "", err
	}

	// 修改登录状态
	var u models.User
	if logintype == "ldap" {
		u.Department = user.Department
	}
	u.Mail = user.Mail
	u.Nickname = user.Nickname
	u.Mode = user.Mode
	u.Online = true
	u.ID = user.ID
	err = userRepository.Update(u)

	return token, err
}

func BuildCacheKeyByToken(token string) string {
	cacheKey := strings.Join([]string{constants.Token, token}, ":")
	return cacheKey
}

func GetTokenFormCacheKey(cacheKey string) string {
	token := strings.Split(cacheKey, ":")[1]
	return token
}

func loginWithTotpEndpoint(c *gin.Context) {
	var loginAccount LoginAccount
	exgin.Bind(c, &loginAccount)

	// 存储登录失败次数信息
	loginFailCountKey := loginAccount.Username
	v, ok := ntcache.MemCache.Get(loginFailCountKey)
	if !ok {
		v = 1
	}
	count := v.(int)
	if count >= 5 {
		Fail(c, -1, "登录失败次数过多，请稍后再试")
		return
	}

	user, err := userRepository.FindByUsername(loginAccount.Username)
	if err != nil {
		count++
		ntcache.MemCache.Set(loginFailCountKey, count, time.Minute*time.Duration(5))
		FailWithData(c, -1, "您输入的账号或密码不正确", count)
		return
	}

	if err := utils.Encoder.Match([]byte(user.Password), []byte(loginAccount.Password)); err != nil {
		count++
		ntcache.MemCache.Set(loginFailCountKey, count, time.Minute*time.Duration(5))
		FailWithData(c, -1, "您输入的账号或密码不正确", count)
		return
	}

	if !totp.Validate(loginAccount.TOTP, user.TOTPSecret) {
		count++
		ntcache.MemCache.Set(loginFailCountKey, count, time.Minute*time.Duration(5))
		FailWithData(c, -1, "您输入双因素认证授权码不正确", count)
		return
	}

	token, err := LoginSuccess(c, loginAccount, user, "pass")
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, token)
}

func getCommonOauth2Config(c *gin.Context, oauth2type, clientid, clientsecret string) *oauth2.Config {
	if oauth2type == constants.ConfigTypeGitee {
		return &oauth2.Config{
			ClientID:    clientid,
			ClientSecret: clientsecret,
			Scopes:       []string{},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://gitee.com/oauth/authorize",
				TokenURL: "https://gitee.com/oauth/token",
			},
			RedirectURL: getRedirectURL(c),
		}
	} else {
		return &oauth2.Config{
			ClientID:     clientid,
			ClientSecret: clientsecret,
			Scopes:       []string{},
			Endpoint:     GitHubOauth2.Endpoint,
		}
	}
}

func getRedirectURL(c *gin.Context) string {
	schame := "http://"
	if strings.HasPrefix(c.Request.Referer(), "https://") {
		schame = "https://"
	}
	return schame + c.Request.Host + "/oauth2/callback"
}

func Oauth2login(c *gin.Context)  {
	state := utils.RandStringBytesMaskImprSrcUnsafe(6)
	ntcache.MemCache.Set(fmt.Sprintf("%s%s", constants.CacheKeyOauth2State, c.ClientIP()), state, 0)
	authtype := viper.GetString("core.login.oauth2.type")
	authid := viper.GetString("core.login.oauth2.id")
	authsecret := viper.GetString("core.login.oauth2.secret")

	url := getCommonOauth2Config(c, authtype, authid, authsecret).AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusFound, url)
}

func Oauth2Callback(c *gin.Context)  {
	var err error
	authtype := viper.GetString("core.login.oauth2.type")
	authid := viper.GetString("core.login.oauth2.id")
	authsecret := viper.GetString("core.login.oauth2.secret")
	// 验证登录跳转时的 State
	state, ok := ntcache.MemCache.Get(fmt.Sprintf("%s%s", constants.CacheKeyOauth2State, c.ClientIP()))
	if !ok || state.(string) != c.Query("state") {
		err = fmt.Errorf("非法的登录方式")
	}
	oauth2Config := getCommonOauth2Config(c, authtype, authid, authsecret)
	ctx := context.Background()
	var otk *oauth2.Token
	if err == nil {
		otk, err = oauth2Config.Exchange(ctx, c.Query("code"))
	}
	var client *GitHubAPI.Client
	if err == nil {
		oc := oauth2Config.Client(ctx, otk)
		if authtype == constants.ConfigTypeGitee {
			client, err = GitHubAPI.NewEnterpriseClient("https://gitee.com/api/v5/", "https://gitee.com/api/v5/", oc)
		} else {
			client = GitHubAPI.NewClient(oc)
		}
	}
	var gu *GitHubAPI.User
	if err == nil {
		gu, _, err = client.Users.Get(ctx, "")
	}
	if err != nil {
		Fail(c, -1, fmt.Sprintf("登录失败，请联系管理员, 错误信息: %v", err))
		return
	}
	u, err := userRepository.UserGet("username = ?", *gu.Name)
	if err != nil && err != gorm.ErrRecordNotFound {
		errors.Dangerous(err)
		return
	}

	if err == gorm.ErrRecordNotFound {
		// 创建用户
		u = &models.User{
			ID:         utils.UUID(),
			Username:   *gu.Login,
			Nickname:   *gu.Name,
			Role:       constants.RoleDefault,
			Mode:       authtype,
			Mail:       *gu.Email,
		}
		userRepository.Create(u)
	} else {
		u.Nickname = *gu.Name
		u.Mail = *gu.Email
		u.Mode = authtype
	}

	if u.TOTPSecret != "" && u.TOTPSecret != "-" {
		Fail(c, 0, "")
		return
	}

	if u.Baned {
		Fail(c, -1, "当前账号已被禁用")
		return
	}

	token, err := LoginSuccess(c, LoginAccount{Remember: true}, u, authtype)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	//Success(c, token)
	c.SetCookie("X-Auth-Token", token, 60*60*24, "", "", false, false)
	c.Status(http.StatusOK)
	c.Writer.WriteString("<script>window.location.href='/#/'</script>")
}

func LogoutEndpoint(c *gin.Context) {
	token := GetToken(c)
	cacheKey := BuildCacheKeyByToken(token)
	ntcache.MemCache.Delete(cacheKey)
	err := userService.Logout(token)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, nil)
}

func ConfirmTOTPEndpoint(c *gin.Context) {
	if viper.GetBool("mode.demo") {
		Fail(c, 0, "演示模式禁止开启两步验证")
		return
	}
	account, _ := GetCurrentAccount(c)

	var confirmTOTP ConfirmTOTP
	exgin.Bind(c, &confirmTOTP)

	if !totp.Validate(confirmTOTP.TOTP, confirmTOTP.Secret) {
		Fail(c, -1, "TOTP 验证失败，请重试")
		return
	}

	u := models.User{
		TOTPSecret: confirmTOTP.Secret,
		ID:         account.ID,
	}

	if err := userRepository.Update(u); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, nil)
}

func ReloadTOTPEndpoint(c *gin.Context) {
	account, _ := GetCurrentAccount(c)

	key, err := totp.NewTOTP(totp.GenerateOpts{
		Issuer:      c.Request.Host,
		AccountName: account.Username,
	})
	if err != nil {
		Fail(c, -1, err.Error())
		return
	}

	qrcode, err := key.Image(200, 200)
	if err != nil {
		Fail(c, -1, err.Error())
		return
	}

	qrEncode, err := utils.ImageToBase64Encode(qrcode)
	if err != nil {
		Fail(c, -1, err.Error())
		return
	}

	Success(c, map[string]string{
		"qr":     qrEncode,
		"secret": key.Secret(),
	})
}

func ResetTOTPEndpoint(c *gin.Context) {
	account, _ := GetCurrentAccount(c)
	u := models.User{
		TOTPSecret: "-",
		ID:         account.ID,
	}
	if err := userRepository.Update(u); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, "")
}

func ChangePasswordEndpoint(c *gin.Context) {
	if viper.GetBool("mode.demo") {
		Fail(c, 0, "演示模式禁止修改密码")
		return
	}
	account, _ := GetCurrentAccount(c)

	var changePassword ChangePassword
	exgin.Bind(c, &changePassword)

	if err := utils.Encoder.Match([]byte(account.Password), []byte(changePassword.OldPassword)); err != nil {
		Fail(c, -1, "您输入的原密码不正确")
		return
	}

	passwd, err := utils.Encoder.Encode([]byte(changePassword.NewPassword))
	if err != nil {
		errors.Dangerous(err)
		return
	}
	u := models.User{
		Password: string(passwd),
		ID:       account.ID,
	}

	if err := userRepository.Update(u); err != nil {
		errors.Dangerous(err)
		return
	}

	LogoutEndpoint(c)
}

type AccountInfo struct {
	Id         string `json:"id"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Type       string `json:"type"`
	EnableTotp bool   `json:"enableTotp"`
	Mode       string `json:"mode"`
}

func InfoEndpoint(c *gin.Context) {
	account, _ := GetCurrentAccount(c)

	user, err := userRepository.FindById(account.ID)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	info := AccountInfo{
		Id:         user.ID,
		Username:   user.Username,
		Nickname:   user.Nickname,
		Type:       user.Role,
		EnableTotp: user.TOTPSecret != "" && user.TOTPSecret != "-",
		Mode:       user.Mode,
	}
	Success(c, info)
}
