package config

import (
	"fmt"
	"os"

	"tg_bot/tools/logger"
)

// Структура Config содержит конфигурационные параметры
type Config struct {
	TGBotToken     string // токен Telegram-бота
	ServerPort     string // адрес HTTP-сервера
	KafkaPort      string // адрес брокера Kafka
	NameTopicKafka string // имя топика Kafka для отправки сообщений
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	const lblLoad = "tg_bot_micserv/internal/config/config.go/Load()"
	myLogger := logger.NewColorLogger(lblLoad)

	token := os.Getenv("TG_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TG_BOT_TOKEN не указан")
	}

	// Получаем адрес сервера
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return nil, fmt.Errorf("SERVER_PORT не указан")
	}

	// Получаем адрес брокера Kafka
	kafkaPort := os.Getenv("KAFKA_PORT")
	if kafkaPort == "" {
		return nil, fmt.Errorf("KAFKA_PORT не указан")
	}
	myLogger.Info(fmt.Sprintf("Успешно записали kafkaPort = %v", kafkaPort))

	// Получаем топик Kafka
	nameTopicKafka := os.Getenv("NAME_KAFKA_TOPIC")
	if nameTopicKafka == "" {
		return nil, fmt.Errorf("NAME_KAFKA_TOPIC не указан")
	}
	myLogger.Info(fmt.Sprintf("Успешно записали nameTopicKafka = %v", nameTopicKafka))

	return &Config{
		TGBotToken:     token,
		ServerPort:     serverPort,
		KafkaPort:      kafkaPort,
		NameTopicKafka: nameTopicKafka,
	}, nil
}
