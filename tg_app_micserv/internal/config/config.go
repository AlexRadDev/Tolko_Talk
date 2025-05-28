package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	API_ID         int
	API_Hash       string
	Phone          string
	Two_F_Password string
	ServerPort     string
	KafkaPort      []string
	KafkaTopic_01  string
	KafkaTopic_02  string
	KafkaTopic_03  string
	WorkerPoolSize string
}

// Load загружает данные из переменных среды
func Load() (*Config, error) {
	apiIDStr := os.Getenv("TELEGRAM_API_ID")
	if apiIDStr == "" {
		return nil, errors.New("TELEGRAM_API_ID не указан")
	}
	apiID, err := strconv.Atoi(apiIDStr)
	if err != nil {
		return nil, fmt.Errorf("TELEGRAM_API_ID не указан: %w", err)
	}

	apiHash := os.Getenv("TELEGRAM_API_HASH")
	if apiHash == "" {
		return nil, errors.New("TELEGRAM_API_HASH не указан")
	}

	phone := os.Getenv("PHONE")
	if phone == "" {
		return nil, errors.New("PHONE не указан")
	}

	password := os.Getenv("TELEGRAM_PASSWORD")
	if password == "" {
		return nil, errors.New("TELEGRAM_PASSWORD не указан")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return nil, errors.New("SERVER_PORT не указан")
	}

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

	kafkaTopic_01 := os.Getenv("KAFKA_TOPIC_01")
	if kafkaTopic_01 == "" {
		return nil, errors.New("KAFKA_TOPIC_01 не указан")
	}

	kafkaTopic_02 := os.Getenv("KAFKA_TOPIC_02")
	if kafkaTopic_02 == "" {
		return nil, errors.New("KAFKA_TOPIC_02 не указан")
	}

	kafkaTopic_03 := os.Getenv("KAFKA_TOPIC_03")
	if kafkaTopic_03 == "" {
		return nil, errors.New("KAFKA_TOPIC_03 не указан")
	}

	workerPoolSize := os.Getenv("WORKER_POOL_SIZE")
	if kafkaTopic_03 == "" {
		return nil, errors.New("KAFKA_TOPIC_03 не указан")
	}

	return &Config{
		API_ID:         apiID,
		API_Hash:       apiHash,
		Phone:          phone,
		Two_F_Password: password,
		ServerPort:     serverPort,
		KafkaPort:      kafkaPort,
		KafkaTopic_01:  kafkaTopic_01,
		KafkaTopic_02:  kafkaTopic_02,
		KafkaTopic_03:  kafkaTopic_03,
		WorkerPoolSize: workerPoolSize,
	}, nil
}
