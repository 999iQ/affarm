package currency

import (
	"affarm/internal/models"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"log"
	"net/http"
)

// AddCurrencyRequest - структура запроса
type AddCurrencyRequest struct {
	Symbol string `json:"symbol" validate:"required,uppercase,max=10"`
}

// AddCurrency godoc
// @Summary Добавить новую криптовалюту
// @Description Добавляет новую криптовалюту в систему отслеживания
// @Tags currencies
// @Accept json
// @Produce json
// @Param request body AddCurrencyRequest true "Данные валюты"
// @Success 201 {object} models.Currency
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /currency/add [post]
func (h *CurrencyHandler) AddCurrency(w http.ResponseWriter, r *http.Request) {
	// Парсинг запроса
	var req AddCurrencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ошибочное тело запроса", err)
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Валидация
	if err := h.validate.Struct(req); err != nil {
		log.Printf("ошибка валидации", err)
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Поиск существующей валюты с тем же символом, включая удаленные
	var existingCurrency models.Currency
	result := h.db.Unscoped().Where("symbol = ?", req.Symbol).First(&existingCurrency)

	if result.Error == nil { // Валюта найдена (даже если удалена)
		if existingCurrency.DeletedAt != (gorm.DeletedAt{}) { // Валюта была удалена (имеет ненулевое значение DeletedAt)
			// Восстанавливаем валюту, делая DeletedAt невалидным
			existingCurrency.DeletedAt = gorm.DeletedAt{Valid: false}
			if err := h.db.Save(&existingCurrency).Error; err != nil {
				log.Printf("ошибка восстановления валюты из бд", err)
				http.Error(w, `{"error": "Failed to restore currency"}`, http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK) // Используем 200 OK, так как мы обновили существующую запись
			json.NewEncoder(w).Encode(existingCurrency)
			log.Printf("Валюта для отслеживания восстановлена", req.Symbol)
			return
		} else {
			// Валюта уже существует и не удалена
			http.Error(w, `{"error": "Currency already exists"}`, http.StatusConflict)
			return
		}
	} else if errors.Is(result.Error, gorm.ErrRecordNotFound) { // Валюта не найдена (раньше не существовала)
		// Создание новой валюты
		newCurrency := models.Currency{
			Symbol: req.Symbol,
		}

		// Сохранение в БД
		if err := h.db.Create(&newCurrency).Error; err != nil {
			log.Printf("ошибка сохранения в бд", err)
			http.Error(w, `{"error": "Failed to save currency"}`, http.StatusInternalServerError)
			return
		}

		// Ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newCurrency)
		log.Printf("Новая валюта для отслеживания добавлена", req.Symbol)
	} else {
		// Другая ошибка при запросе
		log.Printf("ошибка при запросе в бд", result.Error)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
}
