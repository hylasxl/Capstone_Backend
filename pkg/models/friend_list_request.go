package models

import "gorm.io/gorm"

type FriendListRequest struct {
	gorm.Model
	SenderAccountID   uint         `gorm:"not null"`
	ReceiverAccountID uint         `gorm:"not null"`
	RequestStatus     NormalStatus `gorm:"type:ENUM('approved','rejected','pending');default:'pending';not null"`
	IsRecalled        bool         `gorm:"default:false"`
	SenderAccount     Account      `gorm:"foreignkey:SenderAccountID"`
	ReceiverAccount   Account      `gorm:"foreignkey:ReceiverAccountID"`
}
