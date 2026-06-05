package models

import "gorm.io/datatypes"

type UserProfile struct {
	BaseModel
	UserID       string           `gorm:"type:varchar(255);uniqueIndex;not null"`
	DisplayName  string           `gorm:"type:varchar(255);not null"`
	EnglishLevel string           `gorm:"type:varchar(50);not null"`
	Preferences  datatypes.JSON   `gorm:"type:jsonb"`
}

func (UserProfile) TableName() string { return "user_profiles" }
