package models

import "gorm.io/gorm"

type Role string

const (
	User  Role = "user"
	Admin Role = "admin"
)

type AccountRole struct {
	gorm.Model
	Role        Role                      `gorm:"type: ENUM('user','admin');default:'user';not null"`
	Description string                    `gorm:"type: TEXT"`
	Permissions []PermissionByAccountRole `gorm:"foreignKey:AccountRoleID"`
}
