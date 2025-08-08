package services

import (
	"affarm/config"
	"affarm/internal/models"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"time"
)

type PriceUpdater struct {
	db           *gorm.DB
	interval     time.Duration
	binanceURL   string
	convertation string
	stopChannel  chan bool
}

func NewPriceUpdater(db *gorm.DB, cfg *config.BinanceConfig) *PriceUpdater {
	if db == nil {
		log.Panic("ошибка, подключение к базе не существует")
	}
	if cfg == nil {
		log.Panic("ошибка, конфиг отсутствует")
	}

	return &PriceUpdater{
		db:           db,
		interval:     time.Duration(cfg.TimeoutSec) * time.Second,
		binanceURL:   cfg.APIURL + "/api/v3/ticker/price?symbol=%s",
		convertation: cfg.Convertation,
		stopChannel:  make(chan bool),
	}
}

func (pu *PriceUpdater) Start() {
	ticker := time.NewTicker(pu.interval)
	defer ticker.Stop()

	log.Printf("Чекер цен запущен с интервалом %v", pu.interval)

	for {
		select {
		case <-ticker.C:
			pu.updatePrices()
		case <-pu.stopChannel:
			log.Println("Остановка чекера цен")
			return
		}
	}
}

func (pu *PriceUpdater) Stop() {
	pu.stopChannel <- true
}

func (pu *PriceUpdater) updatePrices() {
	// Получаем список всех отслеживаемых валют
	var currencies []models.Currency
	if err := pu.db.Find(&currencies).Error; err != nil {
		log.Printf("ошибка при запросе списка отслеживаемых валют из бд: %v", err)
		return
	}

	if len(currencies) == 0 {
		log.Println("нет валют для отслеживания в бд")
		return
	}

	// Для каждой валюты получаем цену
	for _, currency := range currencies {
		price, err := pu.fetchPriceFromBinance(currency.Symbol)
		if err != nil {
			log.Printf("ошибка при запросе цены на %s: %v", currency.Symbol, err)
			continue
		}

		// Сохраняем цену в БД
		if err := pu.savePrice(currency.ID, price); err != nil {
			log.Printf("ошибка при сохранении цены на %s: %v", currency.Symbol, err)
			continue
		}

		log.Printf("Обновлена цена для %s: %f", currency.Symbol, price)
	}
}

func (pu *PriceUpdater) fetchPriceFromBinance(symbol string) (float64, error) {
	// * ВАЖНО! Здесь к запросу добавляется приставка USDT "+pu.convertation"
	// это нужно чтобы отображать цену криптовалют в USDT
	url := fmt.Sprintf(pu.binanceURL, symbol+pu.convertation)
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("неверный статус код: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка при чтении ответа запроса: %w", err)
	}

	var ticker struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	if err := json.Unmarshal(body, &ticker); err != nil {
		return 0, fmt.Errorf("ошибка при парсинге JSON: %w", err)
	}

	var price float64
	if _, err := fmt.Sscanf(ticker.Price, "%f", &price); err != nil {
		return 0, fmt.Errorf("ошибка при парсинге цены: %w", err)
	}

	return price, nil
}

func (pu *PriceUpdater) savePrice(currencyID uint, price float64) error {
	priceRecord := models.Price{
		CurrencyID: currencyID,
		Price:      price,
		Timestamp:  time.Now(),
	}

	if err := pu.db.Create(&priceRecord).Error; err != nil {
		return fmt.Errorf("ошибка при сохранении цены в бд: %w", err)
	}

	return nil
}
