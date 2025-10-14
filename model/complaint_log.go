package models

import (
	"time"

	"gorm.io/datatypes"
)

type ComplaintLog struct {
	ID         uint           `gorm:"primaryKey"`
	UserID     uint           `gorm:"not null"`
	User       User           `gorm:"foreignKey:UserID" json:"-"`
	Complaints datatypes.JSON `gorm:"not null"` // Menyimpan array keluhan
	Message    string         `gorm:"type:text;not null"`
	CreatedAt  time.Time
}