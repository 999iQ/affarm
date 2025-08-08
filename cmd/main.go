package main

import (
	"affarm/config"
	"affarm/internal/database"
	"affarm/internal/handlers"
	services "affarm/internal/service"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.Load("config.yml")
	if err != nil {
		log.Fatal(err)
	}
	// Подключение к БД
	db, err := database.GetGormDB()
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация роутера
	r := handlers.NewRouter(db)

	// Создаем чекер цен с заданным интервалом
	priceUpdater := services.NewPriceUpdater(db, cfg)
	// Запускаем чекер цен в отдельной горутине
	go priceUpdater.Start()
	defer priceUpdater.Stop()

	// Настройка сервера
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Println("Сервер запущен по пути http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
