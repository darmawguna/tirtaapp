package models // atau model

import "time"

type MedicationRefillSchedule struct {
	ID               uint      `gorm:"primaryKey"`
	UserID           uint      `gorm:"not null"`
	User             User      `gorm:"foreignKey:UserID"`
	RefillDate       time.Time `gorm:"type:date;not null"`         // Tanggal pengambilan obat
	IsActive         bool      `gorm:"not null;default:true"`
	NotificationSent bool      `gorm:"not null;default:false"` // Flag notifikasi H-1
	CreatedAt        time.Time
	UpdatedAt        time.Time
}