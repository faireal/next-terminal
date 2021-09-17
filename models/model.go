// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package models

import (
	"github.com/ergoapi/glog"
	"github.com/ergoapi/util/ztime"
	"github.com/ergoapi/zlog"
	"github.com/spf13/viper"
	"gopkg.in/guregu/null.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/prometheus"
	"time"
)

var Migrates []interface{}

type Model struct {
	CreatedAt null.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt null.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt null.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func Migrate(obj interface{}) {
	Migrates = append(Migrates, obj)
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

	dbcfg, _ := db.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	dbcfg.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	dbcfg.SetMaxOpenConns(100)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	dbcfg.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(Migrates...); err != nil {
		zlog.Panic("初始化数据库表结构异常: %v", err)
	}
	zlog.Info("create db engine success...")
	return db
}
