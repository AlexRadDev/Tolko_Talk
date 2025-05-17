// Загрузка (парсинг) кофига

package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"tg_app_micserv/tools/logger"
)

// Config содержит конфиг микросервиса
type Config struct {
	API_ID         int
	API_Hash       string
	Phone          string
	Two_F_Password string
	Port           string
	KafkaPort      string
	KafkaTopic     string
}

// Load загружает данные из переменных среды
func Load() (*Config, error) {
	const lbl = "tg_app_micserv/config/config.go/Load()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	apiIDStr := os.Getenv("TELEGRAM_API_ID")
	if apiIDStr == "" {
		return nil, errors.New("TELEGRAM_API_ID не указан")
	}
	apiID, err := strconv.Atoi(apiIDStr)

	if err != nil {
		return nil, fmt.Errorf("TELEGRAM_API_ID не указан: %w", err)
	}
	slog.Info("Успешно прочитали TELEGRAM_API_ID")

	apiHash := os.Getenv("TELEGRAM_API_HASH")
	if apiHash == "" {
		return nil, errors.New("TELEGRAM_API_HASH не указан")
	}
	slog.Info("Успешно прочитали TELEGRAM_API_HASH")

	phone := os.Getenv("PHONE")
	if phone == "" {
		return nil, errors.New("PHONE не указан")
	}
	slog.Info("Успешно прочитали PHONE")

	password := os.Getenv("TELEGRAM_PASSWORD")
	if password == "" {
		return nil, errors.New("TELEGRAM_PASSWORD не указан")
	}
	slog.Info("Успешно прочитали TELEGRAM_PASSWORD")

	port := os.Getenv("PORT")
	if port == "" {
		return nil, errors.New("PORT не указан")
	}
	slog.Info("Успешно прочитали PORT")

	kafkaPort := os.Getenv("KAFKA_PORT")
	if kafkaPort == "" {
		return nil, errors.New("KAFKA_PORT не указан")
	}
	slog.Info("Успешно прочитали KAFKA_PORT")

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		return nil, errors.New("KAFKA_TOPIC не указан")
	}
	slog.Info("Успешно прочитали KAFKA_TOPIC")

	return &Config{
		API_ID:         apiID,
		API_Hash:       apiHash,
		Phone:          phone,
		Two_F_Password: password,
		Port:           port,
		KafkaPort:      kafkaPort,
		KafkaTopic:     kafkaTopic,
	}, nil
}
