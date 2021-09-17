package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	model2 "next-terminal/models"
	"next-terminal/pkg/utils"
	"strings"
	"time"

	"next-terminal/pkg/totp"

	ntcache "next-terminal/pkg/cache"
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
	User     model2.User
}

//
//type UserServer struct {
//	repository.UserRepository
//}

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
	if count >= 5 {
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

	token, err := LoginSuccess(c, loginAccount, user)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, token)
}

func LoginSuccess(c *gin.Context, loginAccount LoginAccount, user model2.User) (token string, err error) {
	token = strings.Join([]string{utils.UUID(), utils.UUID(), utils.UUID(), utils.UUID()}, "")

	authorization := Authorization{
		Token:    token,
		Remember: loginAccount.Remember,
		User:     user,
	}

	cacheKey := BuildCacheKeyByToken(token)

	if authorization.Remember {
		// 记住登录有效期两周
		ntcache.MemCache.Set(cacheKey, authorization, RememberEffectiveTime)
	} else {
		ntcache.MemCache.Set(cacheKey, authorization, NotRememberEffectiveTime)
	}

	// 保存登录日志
	loginLog := model2.LoginLog{
		ID:              token,
		UserId:          user.ID,
		ClientIP:        c.ClientIP(),
		ClientUserAgent: c.Request.UserAgent(),
		LoginTime:       utils.NowJsonTime(),
		Remember:        authorization.Remember,
	}

	if loginLogRepository.Create(&loginLog) != nil {
		return "", err
	}

	// 修改登录状态
	err = userRepository.Update(&model2.User{Online: true, ID: user.ID})

	return token, err
}

func BuildCacheKeyByToken(token string) string {
	cacheKey := strings.Join([]string{Token, token}, ":")
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

	token, err := LoginSuccess(c, loginAccount, user)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, token)
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

	u := &model2.User{
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
	u := &model2.User{
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
	u := &model2.User{
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
		Type:       user.Type,
		EnableTotp: user.TOTPSecret != "" && user.TOTPSecret != "-",
	}
	Success(c, info)
}
