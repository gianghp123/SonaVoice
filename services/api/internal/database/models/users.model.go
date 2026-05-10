package models

type User struct {
	BaseModel
	Username string `gorm:"type:varchar(255);uniqueIndex"`
	Email    string `gorm:"type:varchar(255);uniqueIndex"`
	Role     string `gorm:"type:varchar(255);not null"`
}

func (User) TableName() string { return "users" }
