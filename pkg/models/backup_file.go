package models

import "gorm.io/gorm"

type BackupFile struct {
	gorm.Model
	CreatedByAccountID uint    `gorm:"not null"`
	FileName           string  `gorm:"not null; type:TEXT"`
	Location           string  `gorm:"not null; type:TEXT"`
	CreatedByAccount   Account `gorm:"foreignkey:CreatedByAccountID"`
}
