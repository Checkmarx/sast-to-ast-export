package store

import (
	"github.com/checkmarxDev/ast-sast-export/internal/database"
	"gorm.io/gorm"
)

type NodeResultsStore interface {
	GetByResultPathAndNode(resultID, pathID string, nodeID int) (*database.NodeResult, error)
}

type NodeResults struct {
	db *gorm.DB
}

func NewNodeResults(db *gorm.DB) *NodeResults {
	return &NodeResults{db: db}
}

func (e *NodeResults) GetByResultPathAndNode(resultID, pathID string, nodeID int) (*database.NodeResult, error) {
	m := database.NodeResult{}
	tx := e.db.Model(&m).Table("NodeResults").Where("[ResultId] = ? AND [Path_Id] = ? AND [Node_Id] = ?", resultID, pathID, nodeID).First(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &m, nil
}
