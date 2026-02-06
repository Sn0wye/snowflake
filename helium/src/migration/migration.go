package migration

import (
	"snowflake/pkg/logger"
	"snowflake/src/models"

	"gorm.io/gorm"
)

type Migrate struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewMigrate(db *gorm.DB, log *logger.Logger) *Migrate {
	return &Migrate{
		db:  db,
		log: log,
	}
}
func (m *Migrate) Run() {
	err := m.db.AutoMigrate(models.RetrieveAll()...)
	if err != nil {
		return
	}

	m.log.Info("Migration ended")
}

func (m *Migrate) DropAll() {
	err := m.db.Migrator().DropTable(models.RetrieveAll()...)
	if err != nil {
		return
	}
}
