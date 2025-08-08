package models

import (
	"gorm.io/gorm"
)

// Модель валюты
type Currency struct {
	gorm.Model
	Symbol    string `gorm:"uniqueIndex;size:10"`
	DeletedAt gorm.DeletedAt
	Prices    []Price // Связь один-ко-многим
}
