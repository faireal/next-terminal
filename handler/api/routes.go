package api

import (
	"crypto/md5"
	"fmt"
	"github.com/ergoapi/exgin"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"next-terminal/constants"
	"next-terminal/models"
	"next-terminal/pkg/utils"
	"next-terminal/repository"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
	"next-terminal/pkg/service"
)

var (
	userRepository           *repository.UserRepository
	userGroupRepository      *repository.UserGroupRepository
	resourceSharerRepository *repository.ResourceSharerRepository
	assetRepository          *repository.AssetRepository
	clusterRepository        *repository.ClusterRepository

	credentialRepository     *repository.CredentialRepository
	configsRepository        *repository.ConfigsRepository
	commandRepository        *repository.CommandRepository
	sessionRepository        *repository.SessionRepository
	numRepository            *repository.NumRepository
	accessSecurityRepository *repository.AccessSecurityRepository
	jobRepository            *repository.JobRepository
	jobLogRepository         *repository.JobLogRepository
	logsRepository           *repository.LogsRepository

	jobService        *service.JobService
	configsService    *service.ConfigsService
	userService       *service.UserService
	sessionService    *service.SessionService
	mailService       *service.MailService
	numService        *service.NumService
	assetService      *service.AssetService
	credentialService *service.CredentialService
)

func SetupRoutes(db *gorm.DB) *gin.Engine {

	InitRepository(db)
	InitService()

	LoadJobs()

	if err := InitDBData(); err != nil {
		zlog.Panic("初始化数据异常: %v", err)
	}

	if err := ReloadData(); err != nil {
		return nil
	}

	e := exgin.Init(true)
	e.Use(ExLog())
	e.Use(ExCors())
	e.Use(ErrorHandler())
	e.Use(TcpWall())
	e.Use(Auth())

	e.StaticFile("/", "web/build/index.html")
	e.StaticFile("/asciinema.html", "web/build/asciinema.html")
	e.StaticFile("/asciinema-player.js", "web/build/asciinema-player.js")
	e.StaticFile("/asciinema-player.css", "web/build/asciinema-player.css")
	e.StaticFile("/logo.svg", "web/build/logo.svg")
	e.StaticFile("/favicon.ico", "web/build/favicon.ico")
	e.Static("/static", "web/build/static")

	e.POST("/login", LoginEndpoint)
	e.POST("/ldaplogin", LdapLoginEndpoint)
	e.POST("/loginWithTotp", loginWithTotpEndpoint)
	if viper.GetBool("core.login.oauth2") {
		e.GET("/oauth2/login", Oauth2login)
		e.GET("/oauth2/callback", Oauth2Callback)
	}
	e.GET("/version", RVersion)
	e.GET("/metrics", gin.WrapH(promhttp.Handler()))

	e.GET("/tunnel", TunEndpoint)
	e.GET("/ssh", SSHEndpoint)

	e.POST("/logout", LogoutEndpoint)
	e.GET("/showcfg", ShowCfg)
	e.POST("/apis/change-password", ChangePasswordEndpoint)
	e.GET("/apis/reload-totp", ReloadTOTPEndpoint)
	e.POST("/apis/reset-totp", ResetTOTPEndpoint)
	e.POST("/apis/confirm-totp", ConfirmTOTPEndpoint)
	e.GET("/apis/info", InfoEndpoint)

	users := e.Group("/apis/users")
	{
		users.POST("", UserCreateEndpoint, Admin())
		users.GET("/paging", UserPagingEndpoint)
		users.PUT("/:id", UserUpdateEndpoint, Admin())
		users.DELETE("/:id", UserDeleteEndpoint, Admin())
		users.GET("/:id", UserGetEndpoint, Admin())
		users.POST("/:id/change-password", UserChangePasswordEndpoint, Admin())
		users.POST("/:id/reset-totp", UserResetTotpEndpoint, Admin())
	}

	userGroups := e.Group("/apis/user-groups", Admin())
	{
		userGroups.POST("", UserGroupCreateEndpoint)
		userGroups.GET("/paging", UserGroupPagingEndpoint)
		userGroups.PUT("/:id", UserGroupUpdateEndpoint)
		userGroups.DELETE("/:id", UserGroupDeleteEndpoint)
		userGroups.GET("/:id", UserGroupGetEndpoint)
		//userGroups.POST("/:id/members", UserGroupAddMembersEndpoint)
		//userGroups.DELETE("/:id/members/:memberId", UserGroupDelMembersEndpoint)
	}

	assets := e.Group("/apis/assets")
	{
		assets.GET("", AssetAllEndpoint)
		assets.POST("", AssetCreateEndpoint)
		assets.POST("/import", AssetImportEndpoint, Admin())
		assets.GET("/paging", AssetPagingEndpoint)
		assets.POST("/:id/tcping", AssetTcpingEndpoint)
		assets.PUT("/:id", AssetUpdateEndpoint)
		assets.DELETE("/:id", AssetDeleteEndpoint)
		assets.GET("/:id", AssetGetEndpoint)
		assets.POST("/:id/change-owner", AssetChangeOwnerEndpoint, Admin())
	}

	clusters := e.Group("/apis/clusters")
	{
		clusters.GET("", ClusterGetAll)
		clusters.GET("/paging", ClusterPagingEndpoint)
	}

	e.GET("/apis/tags", AssetTagsEndpoint)

	k8s := e.Group("/apis/k8s")
	{
		k8s.GET("/paging", CommandPagingEndpoint)
		k8s.POST("", CommandCreateEndpoint)
		k8s.PUT("/:id", CommandUpdateEndpoint)
		k8s.DELETE("/:id", CommandDeleteEndpoint)
		k8s.GET("/:id", CommandGetEndpoint)
		k8s.POST("/:id/change-owner", CommandChangeOwnerEndpoint, Admin())
	}

	credentials := e.Group("/apis/credentials")
	{
		credentials.GET("", CredentialAllEndpoint)
		credentials.GET("/paging", CredentialPagingEndpoint)
		credentials.POST("", CredentialCreateEndpoint)
		credentials.PUT("/:id", CredentialUpdateEndpoint)
		credentials.DELETE("/:id", CredentialDeleteEndpoint)
		credentials.GET("/:id", CredentialGetEndpoint)
		credentials.POST("/:id/change-owner", CredentialChangeOwnerEndpoint, Admin())
	}

	sessions := e.Group("/apis/sessions")
	{
		sessions.POST("", SessionCreateEndpoint)
		sessions.GET("/paging", SessionPagingEndpoint, Admin())
		sessions.POST("/:id/connect", SessionConnectEndpoint)
		sessions.POST("/:id/disconnect", SessionDisconnectEndpoint, Admin())
		sessions.POST("/:id/resize", SessionResizeEndpoint)
		sessions.GET("/:id/ls", SessionLsEndpoint)
		sessions.GET("/:id/download", SessionDownloadEndpoint)
		sessions.POST("/:id/upload", SessionUploadEndpoint)
		sessions.POST("/:id/mkdir", SessionMkDirEndpoint)
		sessions.POST("/:id/rm", SessionRmEndpoint)
		sessions.POST("/:id/rename", SessionRenameEndpoint)
		sessions.DELETE("/:id", SessionDeleteEndpoint, Admin())
		sessions.GET("/:id/recording", SessionRecordingEndpoint)
	}

	resourceSharers := e.Group("/apis/resource-sharers")
	{
		resourceSharers.GET("/sharers", RSGetSharersEndPoint)
		resourceSharers.POST("/overwrite-sharers", RSOverwriteSharersEndPoint)
		resourceSharers.POST("/remove-resources", ResourceRemoveByUserIdAssignEndPoint, Admin())
		resourceSharers.POST("/add-resources", ResourceAddByUserIdAssignEndPoint, Admin())
	}

	loginLogs := e.Group("/apis/login-logs", Admin())
	{
		loginLogs.GET("/paging", LoginLogPagingEndpoint)
		loginLogs.DELETE("/:id", LoginLogDeleteEndpoint)
	}

	e.GET("/apis/properties", PropertyGetEndpoint, Admin())
	e.PUT("/apis/properties", PropertyUpdateEndpoint, Admin())

	e.GET("/apis/overview/counter", OverviewCounterEndPoint)
	e.GET("/apis/overview/sessions", OverviewSessionPoint)

	jobs := e.Group("/apis/jobs", Admin())
	{
		jobs.POST("", JobCreateEndpoint)
		jobs.GET("/paging", JobPagingEndpoint)
		jobs.PUT("/:id", JobUpdateEndpoint)
		jobs.POST("/:id/change-status", JobChangeStatusEndpoint)
		jobs.POST("/:id/exec", JobExecEndpoint)
		jobs.DELETE("/:id", JobDeleteEndpoint)
		jobs.GET("/:id", JobGetEndpoint)
		jobs.GET("/:id/logs", JobGetLogsEndpoint)
		jobs.DELETE("/:id/logs", JobDeleteLogsEndpoint)
	}

	securities := e.Group("/apis/securities", Admin())
	{
		securities.POST("", SecurityCreateEndpoint)
		securities.GET("/paging", SecurityPagingEndpoint)
		securities.PUT("/:id", SecurityUpdateEndpoint)
		securities.DELETE("/:id", SecurityDeleteEndpoint)
		securities.GET("/:id", SecurityGetEndpoint)
	}

	e.NoMethod(func(c *gin.Context) {
		exgin.GinsAbortWithCode(c, 200, fmt.Sprintf("not found: %v", c.Request.Method))
	})
	e.NoRoute(func(c *gin.Context) {
		exgin.GinsAbortWithCode(c, 200, fmt.Sprintf("not found: %v", c.Request.URL.Path))
	})
	return e
}

func ReloadData() error {
	if err := ReloadAccessSecurity(); err != nil {
		return err
	}

	if err := ReloadToken(); err != nil {
		return err
	}
	return nil
}

func InitRepository(db *gorm.DB) {
	userRepository = repository.NewUserRepository(db)
	userGroupRepository = repository.NewUserGroupRepository(db)
	resourceSharerRepository = repository.NewResourceSharerRepository(db)
	assetRepository = repository.NewAssetRepository(db)
	clusterRepository = repository.NewClusterRepository(db)
	credentialRepository = repository.NewCredentialRepository(db)
	configsRepository = repository.NewConfigsRepository(db)
	commandRepository = repository.NewCommandRepository(db)
	sessionRepository = repository.NewSessionRepository(db)
	numRepository = repository.NewNumRepository(db)
	accessSecurityRepository = repository.NewAccessSecurityRepository(db)
	jobRepository = repository.NewJobRepository(db)
	jobLogRepository = repository.NewJobLogRepository(db)
	logsRepository = repository.NewLogsRepository(db)
}

func InitService() {
	jobService = service.NewJobService(jobRepository, jobLogRepository, assetRepository, credentialRepository)
	configsService = service.NewConfigsService(configsRepository)
	userService = service.NewUserService(userRepository, logsRepository)
	sessionService = service.NewSessionService(sessionRepository)
	mailService = service.NewMailService(configsRepository)
	numService = service.NewNumService(numRepository)
	assetService = service.NewAssetService(assetRepository, userRepository)
	credentialService = service.NewCredentialService(credentialRepository)
}

func LoadJobs() {
	if err := jobService.LoadJobs(); err != nil {
		zlog.Error("load job err: %v", err)
		return
	}
	zlog.Info("load all job done.")
}

func InitDBData() (err error) {
	if configsService.Init() {
		zlog.Info("initialized")
		return nil
	}
	if err := configsService.InitConfigs(); err != nil {
		return err
	}
	if err := numService.InitNums(); err != nil {
		return err
	}
	if err := userService.InitUser(); err != nil {
		return err
	}
	if err := userService.FixUserOnlineState(); err != nil {
		return err
	}
	if err := sessionService.FixSessionState(); err != nil {
		return err
	}
	if err := sessionService.EmptyPassword(); err != nil {
		return err
	}
	if err := credentialService.Encrypt(); err != nil {
		return err
	}
	if err := assetService.Encrypt(); err != nil {
		return err
	}
	if viper.GetBool("demo") {
		if err := assetService.InitDemoVM(); err != nil {
			return err
		}
	}
	return configsService.InitDone()
}

func ResetPassword(username string) error {
	user, err := userRepository.FindByUsername(username)
	if err != nil {
		return err
	}
	password := "next-terminal"
	passwd, err := utils.Encoder.Encode([]byte(password))
	if err != nil {
		return err
	}
	u := models.User{
		Password: string(passwd),
		ID:       user.ID,
	}
	if err := userRepository.Update(u); err != nil {
		return err
	}
	zlog.Debug("用户「%v」密码初始化为: %v", user.Username, password)
	return nil
}

func ResetTotp(username string) error {
	user, err := userRepository.FindByUsername(username)
	if err != nil {
		return err
	}
	u := models.User{
		TOTPSecret: "-",
		ID:         user.ID,
	}
	if err := userRepository.Update(u); err != nil {
		return err
	}
	zlog.Debug("用户「%v」已重置TOTP", user.Username)
	return nil
}

func ChangeEncryptionKey(oldEncryptionKey, newEncryptionKey string) error {

	oldPassword := []byte(fmt.Sprintf("%x", md5.Sum([]byte(oldEncryptionKey))))
	newPassword := []byte(fmt.Sprintf("%x", md5.Sum([]byte(newEncryptionKey))))

	credentials, err := credentialRepository.FindAll()
	if err != nil {
		return err
	}
	for i := range credentials {
		credential := credentials[i]
		if err := credentialRepository.Decrypt(&credential, oldPassword); err != nil {
			return err
		}
		if err := credentialRepository.Encrypt(&credential, newPassword); err != nil {
			return err
		}
		if err := credentialRepository.UpdateById(&credential, credential.ID); err != nil {
			return err
		}
	}
	assets, err := assetRepository.FindAll()
	if err != nil {
		return err
	}
	for i := range assets {
		asset := assets[i]
		if err := assetRepository.Decrypt(&asset, oldPassword); err != nil {
			return err
		}
		if err := assetRepository.Encrypt(&asset, newPassword); err != nil {
			return err
		}
		if err := assetRepository.UpdateById(&asset, asset.ID); err != nil {
			return err
		}
	}
	zlog.Info("encryption key has being changed.")
	return nil
}

func SetupCache() *cache.Cache {
	// 配置缓存器
	mCache := cache.New(5*time.Minute, 10*time.Minute)
	mCache.OnEvicted(func(key string, value interface{}) {
		if strings.HasPrefix(key, constants.Token) {
			token := GetTokenFormCacheKey(key)
			zlog.Debug("用户Token「%v」过期", token)
			err := userService.Logout(token)
			if err != nil {
				zlog.Error("退出登录失败 %v", err)
			}
		}
	})
	return mCache
}
