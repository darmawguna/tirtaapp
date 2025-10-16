package models

import "time"

type Device struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"not null"`
	User       User      `gorm:"foreignKey:UserID"`
	FCMToken   string    `gorm:"type:text;not null;unique"`
	DeviceType string    `gorm:"type:varchar(50)"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}