package models

import "gorm.io/gorm"

type OTPRetakePassword struct {
	gorm.Model
	AccountID uint      `gorm:"not null"`
	OTP       string    `gorm:"not null"`
	Status    OTPStatus `gorm:"type:ENUM('valid','expired','invalid');default:'valid';not null"`
	Account   Account   `gorm:"foreignkey:AccountID"`
}
