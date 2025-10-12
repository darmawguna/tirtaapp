package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	Email     string    `gorm:"size:255;not null;unique"`
	Password  string    `gorm:"size:255;not null"`
	Role      string    `gorm:"size:50;not null;default:'user' "`
	CreatedAt time.Time
	UpdatedAt time.Time
}