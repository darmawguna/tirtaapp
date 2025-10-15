package models

import "time"

type DrugSchedule struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null"`
	User         User      `gorm:"foreignKey:UserID"`
	DrugName     string    `gorm:"type:varchar(255);not null"`
	Dose         string    `gorm:"type:varchar(100);not null"`
	ScheduleDate time.Time `gorm:"type:date;not null"`
	At06         bool      `gorm:"not null;default:false"`
	At12         bool      `gorm:"not null;default:false"`
	At18         bool      `gorm:"not null;default:false"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}