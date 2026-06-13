package models

import "gorm.io/datatypes"

type GrammarAnalysis struct {
	BaseModel
	SessionID    string         `gorm:"type:uuid;not null"`
	MessageID    string         `gorm:"type:uuid;not null;uniqueIndex"`
	OriginalText string         `gorm:"type:text;not null"`
	Result       datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
}

func (GrammarAnalysis) TableName() string { return "grammar_analyses" }
