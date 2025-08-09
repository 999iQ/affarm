package models

import (
	"gorm.io/gorm"
	"time"
)

// Price - Модель цены с временной меткой для валюты
type Price struct {
	gorm.Model `swaggerignore:"true"`
	Price      float64 `gorm:"type:decimal(20,8)"`
	Timestamp  time.Time
	// FK
	CurrencyID uint     // Внешний ключ (обязательное поле)
	Currency   Currency `gorm:"foreignKey:CurrencyID"` // Явное указание связи
}
