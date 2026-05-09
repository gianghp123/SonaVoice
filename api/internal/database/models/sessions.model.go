package models

import "time"

type Session struct {
	BaseModel
	UserID    string `gorm:"type:varchar(255);not null"`
	StartedAt time.Time
	EndedAt   time.Time
	Messages  []Message `gorm:"foreignKey:SessionID"`
}

func (Session) TableName() string { return "sessions" }
