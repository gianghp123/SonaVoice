package models

import "gorm.io/datatypes"

type GrammarAnalysis struct {
	BaseModel
	SessionID        string         `gorm:"type:uuid;not null"`
	MessageID        string         `gorm:"type:uuid;not null;uniqueIndex"`
	OriginalText     string         `gorm:"type:text;not null"`
	CorrectedText    string         `gorm:"type:text"`
	Explanation      string         `gorm:"type:text"`
	HasCorrection    bool           `gorm:"type:boolean;not null;default:true"`
	Severity         string         `gorm:"type:varchar(255);not null;default:'low'"`
	PracticeSentence string         `gorm:"type:text"`
	PracticeFocus    string         `gorm:"type:text"`
	PracticeReason   string         `gorm:"type:text"`
	Metadata         datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
}

func (GrammarAnalysis) TableName() string { return "grammar_analyses" }
