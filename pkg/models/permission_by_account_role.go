package models

import "gorm.io/gorm"

type PermissionByAccountRole struct {
	gorm.Model
	PermissionID  uint        `gorm:"not null"`
	AccountRoleID uint        `gorm:"not null"`
	Description   string      `gorm:"not null;type:TEXT"`
	Permission    Permission  `gorm:"foreignkey:PermissionID"`
	AccountRole   AccountRole `gorm:"foreignkey:AccountRoleID"`
}
