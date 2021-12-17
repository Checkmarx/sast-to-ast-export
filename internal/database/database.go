package database

import (
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	return gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
}
