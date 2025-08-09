package models

import (
	"gorm.io/gorm"
)

// Currency represents a currency model
// @Description Currency entity
// @Success 200 {object} CurrencyResponse
type Currency struct {
	gorm.Model `swaggerignore:"true"`
	Symbol     string         `gorm:"uniqueIndex;size:10"`
	DeletedAt  gorm.DeletedAt `swaggerignore:"true"`
	Prices     []Price        `swaggerignore:"true"` // Связь один-ко-многим
}
