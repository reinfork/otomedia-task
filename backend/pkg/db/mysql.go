package db

import (
	"github.com/ucup/taskmanager/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectMySQL opens a GORM connection using a go-sql-driver DSN, e.g.
// taskuser:taskpass@tcp(localhost:3306)/taskdb?parseTime=true&charset=utf8mb4&loc=Local
func ConnectMySQL(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
}

// AutoMigrate is used for local/dev convenience.
// Production-like environments should apply backend/migrations/*.sql instead.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.Task{}); err != nil {
		return err
	}
	// Active titles must be unique (duplicate -> 1062 -> HTTP 409).
	return db.Exec(`CREATE UNIQUE INDEX ux_tasks_title_active ON tasks (title)`).Error
}
