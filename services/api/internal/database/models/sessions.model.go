package models

import (
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
)

type Session struct {
	BaseModel
	UserID          string              `gorm:"type:varchar(255);not null"`
	SpeechSessionID string              `gorm:"type:varchar(255);"`
	MaxDuration     int64               `gorm:"type:bigint;not null;default:0"`
	ActualUsage     int64               `gorm:"type:bigint;not null;default:0"`
	QuotaDate       *time.Time          `gorm:"type:date;"`
	StartedAt       time.Time
	EndedAt         time.Time
	Messages        []Message           `gorm:"foreignKey:SessionID"`
	Status          enums.SessionStatus `gorm:"type:varchar(255);not null"`
}

func (Session) TableName() string { return "sessions" }
