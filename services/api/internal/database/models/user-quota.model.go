package models

import "time"

type UserQuota struct {
	BaseModel
	UserID    string    `gorm:"type:varchar(255);not null;uniqueIndex:uq_user_quotas_user_key_date"`
	QuotaKey  string    `gorm:"type:varchar(255);not null;uniqueIndex:uq_user_quotas_user_key_date"`
	QuotaDate time.Time `gorm:"type:date;not null;uniqueIndex:uq_user_quotas_user_key_date"`
	Remaining int64     `gorm:"type:bigint;not null"`
}

func (UserQuota) TableName() string { return "user_quotas" }
