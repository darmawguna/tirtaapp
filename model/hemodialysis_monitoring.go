package models

import "time"

type HemodialysisMonitoring struct {
	ID             uint      `gorm:"primaryKey"`
	UserID         uint      `gorm:"not null;uniqueIndex:idx_user_monitoring_date"` // Bagian dari unique index
	User           User      `gorm:"foreignKey:UserID"`
	MonitoringDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_user_monitoring_date"` // Bagian dari unique index
	BPBefore       string    `gorm:"type:varchar(20)"`
	BPAfter        string    `gorm:"type:varchar(20)"`
	WeightBefore   float64   `gorm:"type:decimal(5,2)"`
	WeightAfter    float64   `gorm:"type:decimal(5,2)"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}