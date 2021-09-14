package config

//var GlobalCfg *Config
//
//type Config struct {
//	Debug              bool
//	Demo               bool
//	DB                 string
//	Server             *Server
//	Mysql              *Mysql
//	Sqlite             *Sqlite
//	ResetPassword      string
//	ResetTotp          string
//	EncryptionKey      string
//	EncryptionPassword []byte
//	NewEncryptionKey   string
//}
//
//type Mysql struct {
//	Hostname string
//	Port     int
//	Username string
//	Password string
//	Database string
//}
//
//type Sqlite struct {
//	File string
//}
//
//type Server struct {
//	Addr string
//	Cert string
//	Key  string
//}
//
//func SetupConfig() *Config {
//
//	viper.SetConfigName("config")
//	viper.SetConfigType("yml")
//	viper.AddConfigPath("/etc/next-terminal/")
//	viper.AddConfigPath("$HOME/.next-terminal")
//	viper.AddConfigPath(".")
//	viper.AutomaticEnv()
//	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
//
//	_ = viper.ReadInConfig()
//
//	var config = &Config{
//		DB: viper.GetString("db"),
//		Mysql: &Mysql{
//			Hostname: viper.GetString("mysql.hostname"),
//			Port:     viper.GetInt("mysql.port"),
//			Username: viper.GetString("mysql.username"),
//			Password: viper.GetString("mysql.password"),
//			Database: viper.GetString("mysql.database"),
//		},
//		Sqlite: &Sqlite{
//			File: viper.GetString("sqlite.file"),
//		},
//		Server: &Server{
//			Addr: viper.GetString("server.addr"),
//			Cert: viper.GetString("server.cert"),
//			Key:  viper.GetString("server.key"),
//		},
//		ResetPassword:    viper.GetString("reset-password"),
//		ResetTotp:        viper.GetString("reset-totp"),
//		Debug:            viper.GetBool("debug"),
//		Demo:             viper.GetBool("demo"),
//		EncryptionKey:    viper.GetString("encryption-key"),
//		NewEncryptionKey: viper.GetString("new-encryption-key"),
//	}
//	GlobalCfg = config
//	return config
//}
//
//func init() {
//	GlobalCfg = SetupConfig()
//}
