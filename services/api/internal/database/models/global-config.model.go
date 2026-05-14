package models

import (
	"gorm.io/datatypes"
)

type GlobalConfig struct {
	BaseModel
	Config datatypes.JSON `gorm:"type:jsonb;not null"`
}

func (GlobalConfig) TableName() string {
	return "global_config"
}
