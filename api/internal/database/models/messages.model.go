package models

import "github.com/gianghp123/SonaVoice/api/internal/core/enums"

type Message struct {
	BaseModel
	SessionID      string            `gorm:"type:uuid;not null"`
	Role           enums.MessageRole `gorm:"type:varchar(255);not null"`
	Transcript     string            `gorm:"type:text"`
	WasInterrupted bool              `gorm:"type:boolean;default:false"`
}

func (Message) TableName() string { return "messages" }
