package models

import "time"

type FluidBalanceLog struct {
	ID             uint      `gorm:"primaryKey"`
	UserID         uint      `gorm:"not null;uniqueIndex:idx_user_date"` // Index unik user+tanggal
	User           User      `gorm:"foreignKey:UserID"`
	LogDate        time.Time `gorm:"type:date;not null;uniqueIndex:idx_user_date"` // Index unik user+tanggal
	IntakeCC       int       `gorm:"not null"`
	OutputCC       int       `gorm:"not null"`
	BalanceCC      int       `gorm:"not null"`
	WarningMessage string    `gorm:"type:text"` // Bisa null jika tidak ada warning
	CreatedAt      time.Time
	UpdatedAt      time.Time
}