package currency

import (
	"affarm/internal/models"
	"encoding/json"
	"log"
	"net/http"
)

// RemoveCurrencyRequest - структура запроса для удаления
type RemoveCurrencyRequest struct {
	ID     *uint   `json:"id,omitempty"`
	Symbol *string `json:"symbol,omitempty" validate:"omitempty,uppercase,max=10"`
}

// RemoveCurrency godoc
// @Summary Удалить криптовалюту
// @Description Удаляет криптовалюту из системы по ID или Symbol
// @Tags currencies
// @Accept json
// @Produce json
// @Param request body RemoveCurrencyRequest true "Параметры удаления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /currency/remove [post]
func (h *CurrencyHandler) RemoveCurrency(w http.ResponseWriter, r *http.Request) {
	// Парсинг запроса
	var req RemoveCurrencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Ошибочное тело запроса: %v", err)
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Валидация
	if req.ID == nil && req.Symbol == nil {
		http.Error(w, `{"error": "Must provide either ID or Symbol"}`, http.StatusBadRequest)
		return
	}

	if req.Symbol != nil {
		if err := h.validate.Var(*req.Symbol, "uppercase,max=10"); err != nil {
			log.Printf("Ошибка валидации: %v", err)
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
	}

	// Удаление из БД
	var currency models.Currency
	query := h.db.Model(&models.Currency{})

	if req.ID != nil {
		query = query.Where("id = ?", *req.ID)
	} else {
		query = query.Where("symbol = ?", *req.Symbol)
	}

	result := query.Delete(&currency)

	if result.Error != nil {
		log.Printf("Ошибка удаления: %v", result.Error)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, `{"error": "Currency not found"}`, http.StatusNotFound)
		return
	}

	// Ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Currency successfully deleted",
	})
	log.Printf("Валюта удалена: ID=%v, Symbol=%v", req.ID, req.Symbol)
}
