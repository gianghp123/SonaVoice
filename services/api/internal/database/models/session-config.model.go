package models

import (
	"gorm.io/datatypes"
)

type SessionConfig struct {
	BaseModel
	Config datatypes.JSON `gorm:"type:jsonb;not null"`
}

func (SessionConfig) TableName() string {
	return "session_config"
}
