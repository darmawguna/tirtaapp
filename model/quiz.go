package models

import "time"

type Quiz struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	Url       string    `gorm:"not null"`
	CreatedBy uint      `gorm:"not null"`
	User      User      `gorm:"foreignKey:CreatedBy"`
	CreatedAt time.Time
	UpdatedAt time.Time
}