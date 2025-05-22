package config

import (
	"fmt"
	"os"
	"strings"
)

// Структура Config содержит конфигурационные параметры
type Config struct {
	TGBotToken      string   // токен Telegram-бота
	ServerPort      string   // адрес HTTP-сервера
	KafkaPort       []string // адрес брокера Kafka
	NameTopic01     string   // имя топика Kafka для отправки сообщений
	NameTopic02     string   // имя топика Kafka для принятия сообщений
	NameTopic04     string   // имя топика Kafka для принятия сообщений
	KafkaTopicSlise []string // Срез имен топиков Kafka
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
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
	kafkaPort := make([]string, 0)
	kafkaPort = append(kafkaPort, os.Getenv("KAFKA_PORT_01"))
	if kafkaPort[0] == "" {
		return nil, fmt.Errorf("KAFKA_PORT_01 не указан")
	}
	kafkaPort = append(kafkaPort, os.Getenv("KAFKA_PORT_02"))
	if kafkaPort[1] == "" {
		return nil, fmt.Errorf("KAFKA_PORT_02 не указан")
	}
	kafkaPort = append(kafkaPort, os.Getenv("KAFKA_PORT_03"))
	if kafkaPort[2] == "" {
		return nil, fmt.Errorf("KAFKA_PORT_03 не указан")
	}

	// Получаем имена топиков Kafka
	kafkaTopicSlise := make([]string, 0)

	nameTopic01 := os.Getenv("NAME_TOPIC_01")
	if nameTopic01 == "" {
		return nil, fmt.Errorf("NAME_TOPIC_01 не указан")
	}
	kafkaTopicSlise = append(kafkaTopicSlise, nameTopic01)

	nameTopic02 := os.Getenv("NAME_TOPIC_02")
	if nameTopic02 == "" {
		return nil, fmt.Errorf("NAME_TOPIC_02 не указан")
	}
	kafkaTopicSlise = append(kafkaTopicSlise, nameTopic02)

	nameTopic03 := os.Getenv("NAME_TOPIC_03")
	if nameTopic03 == "" {
		return nil, fmt.Errorf("NAME_TOPIC_03 не указан")
	}
	kafkaTopicSlise = append(kafkaTopicSlise, nameTopic03)
	nameTopic04 := os.Getenv("NAME_TOPIC_04")
	if nameTopic04 == "" {
		return nil, fmt.Errorf("NAME_TOPIC_04 не указан")
	}
	kafkaTopicSlise = append(kafkaTopicSlise, nameTopic04)

	// Валидация имени топика
	for _, name := range kafkaTopicSlise {
		if name == "" || len(name) > 253 {
			return nil, fmt.Errorf("Недопустимое имя топика: %s", name)
		}
		if strings.ContainsAny(name, " !@#$%^&*()+=[]{}|;:,<>?") {
			return nil, fmt.Errorf("Имя топика содержит недопустимые символы: %s", name)
		}
	}

	return &Config{
		TGBotToken:      token,
		ServerPort:      serverPort,
		KafkaPort:       kafkaPort,
		NameTopic01:     nameTopic01,
		NameTopic02:     nameTopic02,
		NameTopic04:     nameTopic04,
		KafkaTopicSlise: kafkaTopicSlise,
	}, nil
}
