package config

import (
	"errors"
	"fmt"
	"os"
)

// Структура Config содержит конфигурационные параметры
type Config struct {
	ServerPort            string   // Порт HTTP-сервера
	GoogleCredentialsFile string   // Путь к файлу учетных данных Google Cloud
	KafkaPort             []string // адреса брокера Kafka
	TopicProdus           string   // Имя топика Kafka для отправки сообщений
	TopicConsum           string   // Имя топика Kafka для принятия сообщений
	KafkaGroupID          string   // Идентификатор группы консьюмеров Kafka
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

	// Получаем Имя топика продюсера
	topicConsum := os.Getenv("NAME_TOPIC_01")
	if topicConsum == "" {
		return nil, errors.New("NAME_TOPIC_01 не указан")
	}
	//myLogger.Info("Успешно прочитали KAFKA_TOPIC_01")

	topicProdus := os.Getenv("NAME_TOPIC_02")
	if topicProdus == "" {
		return nil, errors.New("NAME_TOPIC_02 не указан")
	}
	//myLogger.Info("Успешно прочитали KAFKA_TOPIC_01")

	// Получаем идентификатор группы консьюмеров Kafka
	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroupID == "" {
		return nil, fmt.Errorf("KAFKA_GROUP_ID не указан")
	}

	return &Config{
		ServerPort:            port,
		GoogleCredentialsFile: credentialsFile,
		KafkaPort:             kafkaPort,
		TopicProdus:           topicProdus,
		TopicConsum:           topicConsum,
		KafkaGroupID:          kafkaGroupID,
	}, nil
}
