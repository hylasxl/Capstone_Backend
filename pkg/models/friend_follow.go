package models

import "gorm.io/gorm"

type FriendFollow struct {
	gorm.Model
	FromAccountID uint `gorm:"not null"`
	ToAccountID   uint `gorm:"not null"`
	IsFollowed    bool `gorm:"default:true"`
}
