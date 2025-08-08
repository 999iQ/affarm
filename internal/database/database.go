package database

import (
	"affarm/internal/models"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strconv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные окружения системы")
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("Для переменной %v установлено значение по умолчанию: %v", key, value)
		return defaultValue
	}
	return value
}

// GetGormDB возвращает подключение к PostgreSQL с GORM
func GetGormDB() (*gorm.DB, error) {
	port, _ := strconv.Atoi(getEnv("PG_PORT", "5432"))

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		getEnv("PG_HOST", "localhost"),
		getEnv("PG_USER", "postgres"),
		getEnv("PG_PASS", "1234"),
		getEnv("PG_DBNAME", "affarm"),
		port,
		"disable",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		//Logger: logger.Default.LogMode(logger.Info), // полное логирование запросов
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к бд: %w", err)
	}

	// Автомиграция структур
	err = db.AutoMigrate(
		&models.Currency{},
		&models.Price{},
	)
	if err != nil {
		panic("ошибка при миграции бд")
	}

	// Настройка пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке подключения к бд: %w", err)
	}

	sqlDB.SetMaxOpenConns(20) // макс соединений открытых одновременно
	sqlDB.SetMaxIdleConns(5)  // макс соединений в режиме ожидания

	return db, nil
}
