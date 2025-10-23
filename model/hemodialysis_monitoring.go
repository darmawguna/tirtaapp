package models

import "time"

type HemodialysisMonitoring struct {
	ID                     uint      `gorm:"primaryKey"`
	UserID                 uint      `gorm:"not null"`
	User                   User      `gorm:"foreignKey:UserID"`
	HemodialysisScheduleID uint      `gorm:"not null;uniqueIndex"` // Link to the schedule
	HemodialysisSchedule   HemodialysisSchedule `gorm:"foreignKey:HemodialysisScheduleID"`
	BPBefore               string    `gorm:"type:varchar(20)"` // Store as string like "120/80"
	BPAfter                string    `gorm:"type:varchar(20)"`
	WeightBefore           float64   `gorm:"type:decimal(5,2)"`
	WeightAfter            float64   `gorm:"type:decimal(5,2)"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
}