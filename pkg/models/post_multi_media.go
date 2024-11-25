package models

import "gorm.io/gorm"

type PostMultiMedia struct {
	gorm.Model
	PostID           uint                     `gorm:"not null"`
	URL              string                   `gorm:"type:TEXT"`
	Content          string                   `gorm:"type:TEXT"`
	IsSelfDeleted    bool                     `gorm:"default:false"`
	IsDeletedByAdmin bool                     `gorm:"default:false"`
	MediaType        MediaType                `gorm:"type:ENUM('picture','video');default:'picture'"`
	UploadStatus     UploadStatus             `gorm:"type:ENUM('uploaded','failed');not null"`
	Post             Post                     `gorm:"foreignkey:PostID"`
	Reactions        []PostMultiMediaReaction `gorm:"foreignkey:PostMediaID"`
}
