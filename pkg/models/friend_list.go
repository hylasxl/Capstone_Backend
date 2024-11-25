package models

import "gorm.io/gorm"

type FriendList struct {
	gorm.Model
	FirstAccountID  uint    `gorm:"not null"`
	SecondAccountID uint    `gorm:"not null"`
	FirstAccount    Account `gorm:"foreignkey:FirstAccountID"`
	SecondAccount   Account `gorm:"foreignkey:SecondAccountID"`
	IsValid         bool    `gorm:"default:true"`
}
