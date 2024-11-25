package models

import "gorm.io/gorm"

type BackupHistory struct {
	gorm.Model
	BackupFileID           uint         `gorm:"not null"`
	BackupStatus           NormalStatus `gorm:"type:ENUM('approved','rejected','pending');default:'approved';not null"`
	ImplementedByAccountID uint         `gorm:"not null"`
	BackupFile             BackupFile   `gorm:"foreignkey:BackupFileID"`
	ImplementedByAccount   Account      `gorm:"foreignkey:ImplementedByAccountID"`
}
