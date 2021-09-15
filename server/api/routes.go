package api

import (
	"crypto/md5"
	"fmt"
	"github.com/ergoapi/exgin"
	"github.com/ergoapi/glog"
	"github.com/ergoapi/util/ztime"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/prometheus"
	"os"
	"strings"
	"time"

	"next-terminal/pkg/service"
	"next-terminal/server/model"
	"next-terminal/server/repository"
	"next-terminal/server/utils"

	"github.com/patrickmn/go-cache"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const Token = "X-Auth-Token"

var (
	userRepository           *repository.UserRepository
	userGroupRepository      *repository.UserGroupRepository
	resourceSharerRepository *repository.ResourceSharerRepository
	assetRepository          *repository.AssetRepository
	credentialRepository     *repository.CredentialRepository
	configsRepository        *repository.ConfigsRepository
	commandRepository        *repository.CommandRepository
	sessionRepository        *repository.SessionRepository
	numRepository            *repository.NumRepository
	accessSecurityRepository *repository.AccessSecurityRepository
	jobRepository            *repository.JobRepository
	jobLogRepository         *repository.JobLogRepository
	loginLogRepository       *repository.LoginLogRepository

	jobService        *service.JobService
	propertyService   *service.ConfigsService
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

	if err := InitDBData(); err != nil {
		zlog.Error("初始化数据异常")
		os.Exit(0)
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
	e.POST("/loginWithTotp", loginWithTotpEndpoint)

	e.GET("/tunnel", TunEndpoint)
	e.GET("/ssh", SSHEndpoint)

	e.POST("/logout", LogoutEndpoint)
	e.POST("/change-password", ChangePasswordEndpoint)
	e.GET("/reload-totp", ReloadTOTPEndpoint)
	e.POST("/reset-totp", ResetTOTPEndpoint)
	e.POST("/confirm-totp", ConfirmTOTPEndpoint)
	e.GET("/info", InfoEndpoint)

	users := e.Group("/users")
	{
		users.POST("", UserCreateEndpoint, Admin())
		users.GET("/paging", UserPagingEndpoint)
		users.PUT("/:id", UserUpdateEndpoint, Admin())
		users.DELETE("/:id", UserDeleteEndpoint, Admin())
		users.GET("/:id", UserGetEndpoint, Admin())
		users.POST("/:id/change-password", UserChangePasswordEndpoint, Admin())
		users.POST("/:id/reset-totp", UserResetTotpEndpoint, Admin())
	}

	userGroups := e.Group("/user-groups", Admin())
	{
		userGroups.POST("", UserGroupCreateEndpoint)
		userGroups.GET("/paging", UserGroupPagingEndpoint)
		userGroups.PUT("/:id", UserGroupUpdateEndpoint)
		userGroups.DELETE("/:id", UserGroupDeleteEndpoint)
		userGroups.GET("/:id", UserGroupGetEndpoint)
		//userGroups.POST("/:id/members", UserGroupAddMembersEndpoint)
		//userGroups.DELETE("/:id/members/:memberId", UserGroupDelMembersEndpoint)
	}

	assets := e.Group("/assets")
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

	e.GET("/tags", AssetTagsEndpoint)

	commands := e.Group("/commands")
	{
		commands.GET("/paging", CommandPagingEndpoint)
		commands.POST("", CommandCreateEndpoint)
		commands.PUT("/:id", CommandUpdateEndpoint)
		commands.DELETE("/:id", CommandDeleteEndpoint)
		commands.GET("/:id", CommandGetEndpoint)
		commands.POST("/:id/change-owner", CommandChangeOwnerEndpoint, Admin())
	}

	credentials := e.Group("/credentials")
	{
		credentials.GET("", CredentialAllEndpoint)
		credentials.GET("/paging", CredentialPagingEndpoint)
		credentials.POST("", CredentialCreateEndpoint)
		credentials.PUT("/:id", CredentialUpdateEndpoint)
		credentials.DELETE("/:id", CredentialDeleteEndpoint)
		credentials.GET("/:id", CredentialGetEndpoint)
		credentials.POST("/:id/change-owner", CredentialChangeOwnerEndpoint, Admin())
	}

	sessions := e.Group("/sessions")
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

	resourceSharers := e.Group("/resource-sharers")
	{
		resourceSharers.GET("/sharers", RSGetSharersEndPoint)
		resourceSharers.POST("/overwrite-sharers", RSOverwriteSharersEndPoint)
		resourceSharers.POST("/remove-resources", ResourceRemoveByUserIdAssignEndPoint, Admin())
		resourceSharers.POST("/add-resources", ResourceAddByUserIdAssignEndPoint, Admin())
	}

	loginLogs := e.Group("login-logs", Admin())
	{
		loginLogs.GET("/paging", LoginLogPagingEndpoint)
		loginLogs.DELETE("/:id", LoginLogDeleteEndpoint)
	}

	e.GET("/properties", PropertyGetEndpoint, Admin())
	e.PUT("/properties", PropertyUpdateEndpoint, Admin())

	e.GET("/overview/counter", OverviewCounterEndPoint)
	e.GET("/overview/sessions", OverviewSessionPoint)

	jobs := e.Group("/jobs", Admin())
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

	securities := e.Group("/securities", Admin())
	{
		securities.POST("", SecurityCreateEndpoint)
		securities.GET("/paging", SecurityPagingEndpoint)
		securities.PUT("/:id", SecurityUpdateEndpoint)
		securities.DELETE("/:id", SecurityDeleteEndpoint)
		securities.GET("/:id", SecurityGetEndpoint)
	}

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
	credentialRepository = repository.NewCredentialRepository(db)
	configsRepository = repository.NewConfigsRepository(db)
	commandRepository = repository.NewCommandRepository(db)
	sessionRepository = repository.NewSessionRepository(db)
	numRepository = repository.NewNumRepository(db)
	accessSecurityRepository = repository.NewAccessSecurityRepository(db)
	jobRepository = repository.NewJobRepository(db)
	jobLogRepository = repository.NewJobLogRepository(db)
	loginLogRepository = repository.NewLoginLogRepository(db)
}

func InitService() {
	jobService = service.NewJobService(jobRepository, jobLogRepository, assetRepository, credentialRepository)
	propertyService = service.NewConfigsService(configsRepository)
	userService = service.NewUserService(userRepository, loginLogRepository)
	sessionService = service.NewSessionService(sessionRepository)
	mailService = service.NewMailService(configsRepository)
	numService = service.NewNumService(numRepository)
	assetService = service.NewAssetService(assetRepository, userRepository)
	credentialService = service.NewCredentialService(credentialRepository)
}

func InitDBData() (err error) {
	if err := propertyService.InitConfigs(); err != nil {
		return err
	}
	if err := numService.InitNums(); err != nil {
		return err
	}
	if err := userService.InitUser(); err != nil {
		return err
	}
	if err := jobService.InitJob(); err != nil {
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
	if viper.GetBool("mode.demo") {
		return assetService.InitDemoVM()
	}
	return nil
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
	u := &model.User{
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
	u := &model.User{
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
		if strings.HasPrefix(key, Token) {
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

func SetupDB() *gorm.DB {

	zlog.Debug("当前数据库模式为：%v\n", viper.GetString("db.type"))
	var err error
	var db *gorm.DB
	if viper.GetString("db.type") == "mysql" {
		dsn := viper.GetString("db.dsn")
		dblog := glog.New(zlog.Zlog, viper.GetBool("mode.debug"))

		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger:         dblog,
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
		})
	}

	if err != nil {
		zlog.Panic("连接数据库异常: %v", err)
	}

	if viper.GetBool("db.metrics.enable") {
		dbname := viper.GetString("db.metrics.name")
		if len(dbname) == 0 {
			dbname = "example" + ztime.GetToday()
		}
		db.Use(prometheus.New(prometheus.Config{
			DBName: dbname,
			//RefreshInterval:  0,
			//PushAddr:         "",
			//StartServer:      false,
			//HTTPServerPort:   0,
			//MetricsCollector: nil,
		}))
	}

	if err := db.AutoMigrate(&model.User{}, &model.Asset{}, &model.AssetAttribute{}, &model.Session{}, &model.Command{},
		&model.Credential{}, &model.Configs{}, &model.ResourceSharer{}, &model.UserGroup{}, &model.UserGroupMember{},
		&model.LoginLog{}, &model.Num{}, &model.Job{}, &model.JobLog{}, &model.AccessSecurity{}); err != nil {
		zlog.Panic("初始化数据库表结构异常")
	}
	return db
}
