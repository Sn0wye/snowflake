package db

import (
	"log"

	"snowflake/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func GetDB() *gorm.DB {
	conf := config.GetConfig()
	driver := conf.GetString("db.driver")
	conn := conf.GetString("db.connectionString")

	var db *gorm.DB
	var err error

	switch driver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(conn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Info),
		})
	default: // sqlite (default)
		db, err = gorm.Open(sqlite.Open("db.sqlite"), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Info),
		})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}
