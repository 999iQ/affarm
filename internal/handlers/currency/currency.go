package currency

import (
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// CurrencyHandler - обработчик HTTP-запросов для работы с валютами
type CurrencyHandler struct {
	db       *gorm.DB
	validate *validator.Validate
}

// NewCurrencyHandler - конструктор обработчика
func NewCurrencyHandler(db *gorm.DB) *CurrencyHandler {
	return &CurrencyHandler{db: db,
		validate: validator.New()}
}
