package models

import "gorm.io/gorm"

type DataFieldPrivacy struct {
	gorm.Model
	AccountID      uint          `gorm:"not null"`
	DataFieldIndex uint          `gorm:"not null"`
	PrivacyStatus  PrivacyStatus `gorm:"type:ENUM('public','private','friend_only');default:'public';not null"`
	Account        Account       `gorm:"foreignkey:AccountID"`
}
