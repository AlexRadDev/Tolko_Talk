// Загрузка (парсинг) кофига

package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Config содержит конфиг микросервиса
type Config struct {
	API_ID         int
	API_Hash       string
	Phone          string
	Two_F_Password string
	Port           string
	KafkaPort      []string
	KafkaTopic_01  string
	KafkaTopic_02  string
}

// Load загружает данные из переменных среды
func Load() (*Config, error) {
	//const lbl = "tg_app_micserv/config/config.go/Load()"
	//myLogger := logger.NewColorLogger(lbl)

	apiIDStr := os.Getenv("TELEGRAM_API_ID")
	if apiIDStr == "" {
		return nil, errors.New("TELEGRAM_API_ID не указан")
	}
	apiID, err := strconv.Atoi(apiIDStr)
	if err != nil {
		return nil, fmt.Errorf("TELEGRAM_API_ID не указан: %w", err)
	}
	//myLogger.Info("Успешно прочитали TELEGRAM_API_ID")

	apiHash := os.Getenv("TELEGRAM_API_HASH")
	if apiHash == "" {
		return nil, errors.New("TELEGRAM_API_HASH не указан")
	}
	//myLogger.Info("Успешно прочитали TELEGRAM_API_HASH")

	phone := os.Getenv("PHONE")
	if phone == "" {
		return nil, errors.New("PHONE не указан")
	}
	//myLogger.Info("Успешно прочитали PHONE")

	password := os.Getenv("TELEGRAM_PASSWORD")
	if password == "" {
		return nil, errors.New("TELEGRAM_PASSWORD не указан")
	}
	//myLogger.Info("Успешно прочитали TELEGRAM_PASSWORD")

	port := os.Getenv("PORT")
	if port == "" {
		return nil, errors.New("PORT не указан")
	}
	//myLogger.Info("Успешно прочитали PORT")

	kafkaPort := make([]string, 0)
	kafkaPort = append(kafkaPort, os.Getenv("KAFKA_PORT_01"))
	if kafkaPort[0] == "" {
		return nil, errors.New("KAFKA_PORT_01 не указан")
	}
	kafkaPort = append(kafkaPort, os.Getenv("KAFKA_PORT_02"))
	if kafkaPort[1] == "" {
		return nil, errors.New("KAFKA_PORT_02 не указан")
	}
	kafkaPort = append(kafkaPort, os.Getenv("KAFKA_PORT_03"))
	if kafkaPort[2] == "" {
		return nil, errors.New("KAFKA_PORT_03 не указан")
	}
	//myLogger.Info("Успешно прочитали все KAFKA_PORT")

	kafkaTopic_01 := os.Getenv("KAFKA_TOPIC_01")
	if kafkaTopic_01 == "" {
		return nil, errors.New("KAFKA_TOPIC_01 не указан")
	}
	//myLogger.Info("Успешно прочитали KAFKA_TOPIC_01")

	kafkaTopic_02 := os.Getenv("KAFKA_TOPIC_02")
	if kafkaTopic_02 == "" {
		return nil, errors.New("KAFKA_TOPIC_02 не указан")
	}
	//myLogger.Info("Успешно прочитали KAFKA_TOPIC_01")

	return &Config{
		API_ID:         apiID,
		API_Hash:       apiHash,
		Phone:          phone,
		Two_F_Password: password,
		Port:           port,
		KafkaPort:      kafkaPort,
		KafkaTopic_01:  kafkaTopic_01,
		KafkaTopic_02:  kafkaTopic_02,
	}, nil
}
