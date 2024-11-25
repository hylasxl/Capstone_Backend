package models

import (
	"gorm.io/gorm"
	"time"
)

type Post struct {
	gorm.Model
	AccountID             uint            `gorm:"not null"`
	Content               string          `gorm:"type:TEXT"`
	IsPublishedLater      bool            `gorm:"default:false"`
	PublishLaterTimestamp time.Time       `gorm:"default:null"`
	Latitude              float64         `gorm:"type:decimal(9,6)"`
	Longitude             float64         `gorm:"type:decimal(9,6)"`
	IsShared              bool            `gorm:"default:false"`
	OriginalPostID        uint            `gorm:"not null"`
	IsSelfDeleted         bool            `gorm:"default:false"`
	IsDeletedByAdmin      bool            `gorm:"default:false"`
	IsHidden              bool            `gorm:"default:false"`
	IsContentEdited       bool            `gorm:"default:false"`
	PrivacyStatus         PrivacyStatus   `gorm:"type:ENUM('public','private','friend_only');default:'public';not null"`
	OriginalPost          *Post           `gorm:"foreignkey:OriginalPostID"`
	Account               Account         `gorm:"foreignkey:AccountID"`
	PostReaction          []PostReaction  `gorm:"foreignkey:PostID"`
	PostTags              []PostTagFriend `gorm:"foreignkey:PostID"`
}
