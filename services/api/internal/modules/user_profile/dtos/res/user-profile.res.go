package res

import (
	"time"

	"gorm.io/datatypes"
)

type UserProfileRes struct {
	ID           string         `json:"id"`
	UserID       string         `json:"user_id"`
	DisplayName  string         `json:"display_name"`
	EnglishLevel string         `json:"english_level"`
	Preferences  datatypes.JSON `json:"preferences"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
