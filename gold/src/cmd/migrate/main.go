package main

import (
	"log"
	"os"

	"github.com/Sn0wye/snowflake/gold/pkg/config"
	"github.com/Sn0wye/snowflake/gold/src/db"
	"github.com/Sn0wye/snowflake/gold/src/migration"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down>")
	}

	config.GetConfig()

	gormDB := db.GetDB()
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	runner := migration.NewRunner(sqlDB)

	switch os.Args[1] {
	case "up":
		if err := runner.Up(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	case "down":
		if err := runner.Down(); err != nil {
			log.Fatalf("Migration rollback failed: %v", err)
		}
	default:
		log.Fatalf("Unknown migrate direction: %s (use 'up' or 'down')", os.Args[1])
	}
}
