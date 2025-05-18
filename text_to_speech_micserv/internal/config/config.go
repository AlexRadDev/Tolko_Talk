package config

import (
	"fmt"
	"os"
)

// Структура Config содержит конфигурационные параметры
type Config struct {
	ServerPort            string // Порт HTTP-сервера
	GoogleCredentialsFile string // Путь к файлу учетных данных Google Cloud
	KafkaPort             string // адрес брокера Kafka
	NameTopicProdus       string // Имя топика Kafka для отправки сообщений
	NameTopicConsum       string // Имя топика Kafka для принятия сообщений
	KafkaGroupID          string // Идентификатор группы консьюмеров Kafka
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8080"
	}

	// Получаем путь к файлу учетных данных Google Cloud
	credentialsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	if credentialsFile == "" {
		return nil, fmt.Errorf("GOOGLE_CREDENTIALS_FILE не указан")
	}

	// Получаем адрес брокера Kafka
	kafkaPort := os.Getenv("KAFKA_PORT")
	if kafkaPort == "" {
		return nil, fmt.Errorf("KAFKA_PORT не указан")
	}

	// Получаем Имя топика продюсера
	kafkaTopic := os.Getenv("NAME_TOPIC_KAFKA")
	if kafkaTopic == "" {
		return nil, fmt.Errorf("NAME_TOPIC_KAFKA не указан")
	}

	// Получаем идентификатор группы консьюмеров Kafka
	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroupID == "" {
		return nil, fmt.Errorf("KAFKA_GROUP_ID не указан")
	}
	
	return &Config{
		ServerPort:            port,
		GoogleCredentialsFile: credentialsFile,
		KafkaPort:             kafkaPort,
		NameTopicProdus:       kafkaTopic,
		KafkaGroupID:          kafkaGroupID,
	}, nil
}
