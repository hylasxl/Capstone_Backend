package models

import (
	"gorm.io/gorm"
	"time"
)

type AccountInfo struct {
	gorm.Model
	AccountID       uint            `gorm:"not null"`
	AvatarID        uint            `gorm:"not null"`
	FirstName       string          `gorm:"type:varchar(25)"`
	LastName        string          `gorm:"type:varchar(25)"`
	DateOfBirth     time.Time       `gorm:"type:DATE"`
	Gender          Gender          `gorm:"type:ENUM('male', 'female', 'other');not null"`
	MaritalStatus   MaritalStatus   `gorm:"type:ENUM('single', 'in_a_relationship', 'engaged', 'married', 'in_a_civil_union', 'in_a_domestic_partnership', 'in_an_open_relationship', 'it_complicated', 'separated', 'divorced', 'widowed');not null; default:'single'"`
	PhoneNumber     string          `gorm:"unique; default: null"`
	Email           string          `gorm:"not null; unique"`
	NameDisplayType NameDisplayType `gorm:"type:ENUM('first_name_first','last_name_first');not null;default:'first_name_first'"`
	Account         Account         `gorm:"foreignkey:AccountID"`
	Avatar          AccountAvatar   `gorm:"foreignkey:AvatarID"`
}
