package models

import "gorm.io/gorm"

type Permission struct {
	gorm.Model
	PermissionURL      string  `gorm:"type:TEXT;not null"`
	CreatedByAccountID uint    `gorm:"not null"`
	Description        string  `gorm:"type:TEXT;not null"`
	Account            Account `gorm:"foreignkey:CreatedByAccountID"`
}
