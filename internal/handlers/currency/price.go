package currency

import (
	"affarm/internal/models"
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
// @Router /price/get [post]
func (h *CurrencyHandler) GetPriceAtTime(w http.ResponseWriter, r *http.Request) {
	// Парсинг запроса
	var req GetPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Ошибочное тело запроса: %v", err)
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Валидация
	if err := h.validate.Struct(req); err != nil {
		log.Printf("Ошибка валидации: %v", err)
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Поиск ближайшей цены
	var price models.Price
	query := h.db.Where("symbol = ?", req.Symbol)

	// Сначала пробуем найти точное совпадение
	exactQuery := query.Where("timestamp = ?", req.Timestamp)
	if err := exactQuery.First(&price).Error; err == nil {
		response := PriceResponse{
			Symbol: req.Symbol,
			Price:  price.Price,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Если точного совпадения нет, ищем ближайшее
	var beforePrice, afterPrice models.Price

	// Ближайшее до запрошенного времени
	query.Where("timestamp <= ?", req.Timestamp).
		Order("timestamp DESC").
		First(&beforePrice)

	// Ближайшее после запрошенного времени
	query.Where("timestamp >= ?", req.Timestamp).
		Order("timestamp ASC").
		First(&afterPrice)

	// Выбираем ближайшее по времени значение
	var result models.Price
	switch {
	case beforePrice.ID == 0 && afterPrice.ID == 0:
		http.Error(w, `{"error": "No price data available"}`, http.StatusNotFound)
		return
	case beforePrice.ID == 0:
		result = afterPrice
	case afterPrice.ID == 0:
		result = beforePrice
	default:
		beforeDiff := req.Timestamp.Sub(beforePrice.Timestamp)
		afterDiff := afterPrice.Timestamp.Sub(req.Timestamp)
		if beforeDiff < afterDiff {
			result = beforePrice
		} else {
			result = afterPrice
		}
	}

	response := PriceResponse{
		Symbol: req.Symbol,
		Price:  result.Price,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
