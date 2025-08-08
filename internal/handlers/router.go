package handlers

import (
	"affarm/internal/handlers/currency"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func NewRouter(db *gorm.DB) *http.ServeMux {
	mux := http.NewServeMux()

	currencyHandler := currency.NewCurrencyHandler(db)

	// Регистрация маршрутов API v1
	mux.HandleFunc("POST /api/v1/currency/add", currencyHandler.AddCurrency)
	mux.HandleFunc("POST /api/v1/currency/remove", currencyHandler.RemoveCurrency)
	mux.HandleFunc("GET /api/v1/currency/price", currencyHandler.GetPriceAtTime)

	log.Print("POST /api/v1/currency/add")
	log.Print("POST /api/v1/currency/remove")
	log.Print("GET /api/v1/currency/{symbol}")

	// Статические файлы (опционально)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	return mux
}
