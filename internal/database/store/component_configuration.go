package store

import (
	"github.com/checkmarxDev/ast-sast-export/internal/database"
	"gorm.io/gorm"
)

type CxComponentConfigurationStore interface {
	GetByKey(key string) (*database.ComponentConfiguration, error)
}

type ComponentConfiguration struct {
	db *gorm.DB
}

func NewComponentConfigurationStore(db *gorm.DB) *ComponentConfiguration {
	return &ComponentConfiguration{db: db}
}

func (e *ComponentConfiguration) GetByKey(key string) (*database.ComponentConfiguration, error) {
	m := database.ComponentConfiguration{}
	tx := e.db.Model(&m).Table("CxComponentConfiguration").Where("[Key] = ?", key).First(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &m, nil
}
