package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"default:now()"`
	UpdatedAt time.Time `gorm:"default:now()"`
	DeletedAt gorm.DeletedAt
}
