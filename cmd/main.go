package main

import (
	"avito_coin/internal/config"
	"avito_coin/internal/db"
	"avito_coin/internal/handler"
	"avito_coin/internal/repository"
	"avito_coin/internal/service"
	"os"
	"os/signal"
	"syscall"
	_ "time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Infof(".env file not found: %v", err)
	}
	logrus.Info("Config has been successfuly loaded")

	// Подключение к БД
	DB, err := db.NewPostgresDB(cfg)
	if err != nil {
		logrus.Fatalf("Failed to connect to DB: %v", err)
	}
	DB.SetMaxOpenConns(1000)
	logrus.Info("Database has been successfuly connected")

	// Создание слоя репозитория
	repo := repository.NewRepository(DB)

	// Создание слоя сервиса
	service := service.NewCoinService(repo)

	// Новый экземрляр Echo и задаем sli времени ответа
	e := echo.New()

	// Создание слоя обработчика
	handler.NewCoinHandler(e, service)

	// Запускаем сервер
	go func() {
		if err := e.Start(":8080"); err != nil {
			logrus.Fatalf("error starting server: %v", err)
		}
	}()

	// Канал для обработки сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала завершения
	<-stop
	logrus.Info("Received shutdown signal. Gracefully shutting down...")
}
