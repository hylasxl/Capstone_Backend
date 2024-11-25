package models

import "gorm.io/gorm"

type DataFieldIndex struct {
	gorm.Model
	DataFieldName string `gorm:"not null"`
}
