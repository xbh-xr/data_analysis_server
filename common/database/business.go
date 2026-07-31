package database

import (
	"errors"
	"time"

	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	toolsDB "github.com/go-admin-team/go-admin-core/tools/database"
	. "github.com/go-admin-team/go-admin-core/tools/gorm/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"go-admin/config"
)

const BusinessDBKey = "business"

var businessDB *gorm.DB

// SetupBusiness 初始化业务库（仅查询，不初始化 Casbin）
func SetupBusiness() {
	c := config.ExtConfig.Business
	if !c.Enabled {
		log.Info("business database is disabled")
		return
	}
	if c.Source == "" {
		log.Warn("business database enabled but source is empty, skip")
		return
	}
	if c.Driver == "" {
		c.Driver = "mysql"
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 10
	}
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = 100
	}

	open, ok := opens[c.Driver]
	if !ok {
		log.Fatal(pkg.Red("business database unsupported driver: " + c.Driver))
	}

	log.Infof("%s => %s", BusinessDBKey, pkg.Green(c.Source))
	resolverConfig := toolsDB.NewConfigure(
		c.Source,
		c.MaxIdleConns,
		c.MaxOpenConns,
		c.ConnMaxIdleTime,
		c.ConnMaxLifeTime,
		nil,
	)
	db, err := resolverConfig.Init(&gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: New(
			logger.Config{
				SlowThreshold: time.Second,
				Colorful:      true,
				LogLevel: logger.LogLevel(
					log.DefaultLogger.Options().Level.LevelForGorm()),
			},
		),
	}, open)
	if err != nil {
		log.Fatal(pkg.Red("business "+c.Driver+" connect error :"), err)
	}

	businessDB = db
	log.Info(pkg.Green("business " + c.Driver + " connect success !"))
}

// GetBusinessDB 获取业务库连接
func GetBusinessDB() (*gorm.DB, error) {
	if businessDB == nil {
		return nil, errors.New("business database is not initialized")
	}
	return businessDB, nil
}

// MustGetBusinessDB 获取业务库连接，未初始化时直接 panic
func MustGetBusinessDB() *gorm.DB {
	db, err := GetBusinessDB()
	if err != nil {
		panic(err)
	}
	return db
}
