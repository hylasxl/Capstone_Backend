package models

import "gorm.io/gorm"

type GBannedWord struct {
	gorm.Model
	CreatedByAccountID uint    `gorm:"not null"`
	Content            string  `gorm:"type:TEXT;not null"`
	IsDeleted          bool    `gorm:"default:false"`
	CreatedByAccount   Account `gorm:"foreignkey:CreatedByAccountID"`
}
