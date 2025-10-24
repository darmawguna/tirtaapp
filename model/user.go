package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	Email     string    `gorm:"size:255;not null;unique"`
	Password  string    `gorm:"size:255;not null"`
	ProfilePicture string    `gorm:"size:255;default:null"`
	PhoneNumber string  `gorm:"size:30; not null"`
	Role      string    `gorm:"size:50;not null;default:'user' "`
	Timezone  string    `gorm:"size:100;not null;default:'Asia/Makassar'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}