package models

import (
	"time"

	"gorm.io/datatypes"
)

type HemodialysisComplaint struct {
	ID         uint           `gorm:"primaryKey "`
	UserID     uint           `gorm:"not null;index"`
	Phase      string         `gorm:"type:varchar(20);not null;index"`
	OtherText  *string        `gorm:"type:text"`
	User       User           `gorm:"foreignKey:UserID" json:"-"`
	Complaints datatypes.JSON `gorm:"not null"`
	Message    string         `gorm:"type:text;not null"`
	CreatedAt  time.Time
}