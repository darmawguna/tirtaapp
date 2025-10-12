package models

import "time"

type Education struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	Url       string    `gorm:"not null"`
	Thumbnail string    `gorm:"not null"`
	CreatedBy uint      `gorm:"not null"`
	User      User      `gorm:"foreignKey:CreatedBy"`
	CreatedAt time.Time
	UpdatedAt time.Time
}