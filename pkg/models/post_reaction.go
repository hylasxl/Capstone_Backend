package models

import "gorm.io/gorm"

type PostReaction struct {
	gorm.Model
	PostID       uint         `gorm:"not null"`
	AccountID    uint         `gorm:"not null"`
	IsRecalled   bool         `gorm:"default:false"`
	ReactionType ReactionType `gorm:"type:ENUM('like','dislike','love','hate','cry');default:null"`
	Post         Post         `gorm:"foreignkey:PostID"`
	Account      Account      `gorm:"foreignkey:AccountID"`
}
