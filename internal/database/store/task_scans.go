package store

import (
	"github.com/checkmarxDev/ast-sast-export/internal/database"
	"gorm.io/gorm"
)

type TaskScansStore interface {
	GetByID(scanID string) (*database.TaskScan, error)
}

type TaskScans struct {
	db *gorm.DB
}

func NewTaskScans(db *gorm.DB) (*TaskScans, error) {
	return &TaskScans{db: db}, nil
}

func (e *NodeResults) GetByID(scanID string) (*database.TaskScan, error) {
	m := database.TaskScan{}
	tx := e.db.Model(&m).Table("TaskScans").Where("[Id] = ?", scanID).First(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &m, nil
}
