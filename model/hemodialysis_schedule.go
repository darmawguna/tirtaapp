package models

import "time"

type HemodialysisSchedule struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null"`
	User         User      `gorm:"foreignKey:UserID"`
	ScheduleDate time.Time `gorm:"type:date;not null"`
	Notes        string    `gorm:"type:text"`
	IsActive     bool      `gorm:"not null;default:true"`
	NotificationSent bool      `gorm:"not null;default:false"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}