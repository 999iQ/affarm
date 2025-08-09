package currency

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// GetPriceRequest - структура запроса
type GetPriceRequest struct {
	Symbol    string    `json:"symbol" validate:"required,uppercase,max=10"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
}

// PriceResponse - структура ответа
type PriceResponse struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

// GetPriceAtTime godoc
// @Summary Получить цену на момент времени
// @Description Возвращает цену для указанной валютной пары на заданный момент времени. Если точное значение отсутствует, возвращает ближайшее доступное.
// @Tags prices
// @Accept json
// @Produce json
// @Param request body GetPriceRequest true "Параметры запроса"
// @Success 200 {object} PriceResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /price/get [get]
func (h *CurrencyHandler) GetPriceAtTime(w http.ResponseWriter, r *http.Request) {
	var req GetPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Получаем соединение с БД
	db, err := h.db.DB()
	if err != nil {
		log.Printf("Failed to get DB connection: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// 1. Проверяем существование валюты
	var currencyID uint
	err = db.QueryRow("SELECT id FROM currencies WHERE symbol = $1 AND deleted_at IS NULL", req.Symbol).Scan(&currencyID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "Currency not found"}`, http.StatusNotFound)
		} else {
			log.Printf("Currency lookup error: %v", err)
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	utcTime := req.Timestamp.UTC()

	// 2. Пытаемся найти точное совпадение (±1 секунда)
	var exactPrice float64
	var exactTimestamp time.Time
	err = db.QueryRow(`
        SELECT price, timestamp 
        FROM prices 
        WHERE currency_id = $1 
        AND timestamp BETWEEN $2 AND $3
        LIMIT 1`,
		currencyID,
		utcTime.Add(-time.Second),
		utcTime.Add(time.Second),
	).Scan(&exactPrice, &exactTimestamp)

	if err == nil {
		jsonResponse(w, PriceResponse{Symbol: req.Symbol, Price: exactPrice})
		return
	} else if err != sql.ErrNoRows {
		log.Printf("Exact price query error: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// 3. Ищем ближайшие цены
	var beforePrice, afterPrice float64
	var beforeTimestamp, afterTimestamp time.Time

	// Ближайшая цена до
	err = db.QueryRow(`
        SELECT price, timestamp 
        FROM prices 
        WHERE currency_id = $1 
        AND timestamp <= $2
        ORDER BY timestamp DESC
        LIMIT 1`,
		currencyID,
		utcTime,
	).Scan(&beforePrice, &beforeTimestamp)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Before price query error: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Ближайшая цена после
	err = db.QueryRow(`
        SELECT price, timestamp 
        FROM prices 
        WHERE currency_id = $1 
        AND timestamp >= $2
        ORDER BY timestamp ASC
        LIMIT 1`,
		currencyID,
		utcTime,
	).Scan(&afterPrice, &afterTimestamp)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("After price query error: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// 4. Выбираем результат
	switch {
	case beforeTimestamp.IsZero() && afterTimestamp.IsZero():
		http.Error(w, `{"error": "No price data available for `+req.Symbol+`"}`, http.StatusNotFound)
	case beforeTimestamp.IsZero():
		jsonResponse(w, PriceResponse{Symbol: req.Symbol, Price: afterPrice})
	case afterTimestamp.IsZero():
		jsonResponse(w, PriceResponse{Symbol: req.Symbol, Price: beforePrice})
	default:
		beforeDiff := utcTime.Sub(beforeTimestamp)
		afterDiff := afterTimestamp.Sub(utcTime)
		if beforeDiff < afterDiff {
			jsonResponse(w, PriceResponse{Symbol: req.Symbol, Price: beforePrice})
		} else {
			jsonResponse(w, PriceResponse{Symbol: req.Symbol, Price: afterPrice})
		}
	}
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
