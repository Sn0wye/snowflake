package db

import (
	"log"
	"sync"

	"github.com/getsnowflake/snowflake/helium/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var (
	instance *gorm.DB
	once     sync.Once
)

func GetDB() *gorm.DB {
	once.Do(func() {
		conf := config.GetConfig()
		driver := conf.GetString("db.driver")
		conn := conf.GetString("db.connectionString")

		var err error

		switch driver {
		case "postgres":
			instance, err = gorm.Open(postgres.Open(conn), &gorm.Config{
				Logger: gormLogger.Default.LogMode(gormLogger.Info),
			})
		default: // sqlite (default)
			instance, err = gorm.Open(sqlite.Open("db.sqlite"), &gorm.Config{
				Logger: gormLogger.Default.LogMode(gormLogger.Info),
			})
		}

		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
	})

	return instance
}
