package models

import "gorm.io/gorm"

type PostReported struct {
	gorm.Model
	PostID              uint          `gorm:"not null"`
	ReportedByAccountID uint          `gorm:"not null"`
	Reason              string        `gorm:"type:text;not null"`
	ReportResolve       ReportResolve `gorm:"type:ENUM('report_pending','report_skipped','delete_post');default:'report_pending'"`
	ResolvedByAccountID uint          `gorm:"not null"`
	Post                Post          `gorm:"foreignKey:PostID"`
	ReportedByAccount   Account       `gorm:"foreignKey:ReportedByAccountID"`
	ResolvedByAccount   Account       `gorm:"foreignKey:ResolvedByAccountID"`
}
