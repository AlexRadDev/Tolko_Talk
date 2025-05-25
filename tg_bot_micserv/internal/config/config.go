package config

import (
	"fmt"
	"os"
	"strings"
)

// Структура Config содержит конфигурационные параметры
type Config struct {
	TGBotToken       string   // токен Telegram-бота
	ServerPort       string   // адрес HTTP-сервера
	KafkaPort        []string // адрес брокера Kafka
	NameTopic01      string   // имя топика Kafka для отправки сообщений
	NameTopic02      string   // имя топика Kafka для принятия сообщений
	NameTopic03      string   // имя топика Kafka для принятия сообщений
	NameTopic04      string   // имя топика Kafka для принятия сообщений
	KafkaTopicSlise  []string // Срез имен топиков Kafka
	DBHost           string   // хост базы данных внутри сети докер
	DBPort           string   // порт базы данных внутри контейнера
	PostgresDBName   string   // имя базы данных
	PostgresUser     string   // имя юзера БД
	PostgresPassword string   // пароль к БД
	DBSSLMOD         string
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

	// Получаем адрес брокерров Kafka
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

	// База данных Postgres
	DB_Host := os.Getenv("DB_HOST")
	if DB_Host == "" {
		return nil, fmt.Errorf("DB_HOST не указан")
	}

	DB_Port := os.Getenv("DB_PORT")
	if DB_Port == "" {
		return nil, fmt.Errorf("DB_PORT не указан")
	}

	PostgresDB_Name := os.Getenv("POSTGRES_DB_NAME")
	if PostgresDB_Name == "" {
		return nil, fmt.Errorf("POSTGRES_DB_NAME не указан")
	}

	PostgresDB_User := os.Getenv("POSTGRES_USER")
	if PostgresDB_User == "" {
		return nil, fmt.Errorf("POSTGRES_USER не указан")
	}

	PostgresDB_Password := os.Getenv("POSTGRES_PASSWORD")
	if PostgresDB_Password == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD не указан")
	}

	DB_SSL_MODE := os.Getenv("DB_SSL_MODE")
	if DB_SSL_MODE == "" {
		return nil, fmt.Errorf("DB_SSL_MODE не указан")
	}

	return &Config{
		TGBotToken:       token,
		ServerPort:       serverPort,
		KafkaPort:        kafkaPort,
		NameTopic01:      nameTopic01,
		NameTopic02:      nameTopic02,
		NameTopic03:      nameTopic03,
		NameTopic04:      nameTopic04,
		KafkaTopicSlise:  kafkaTopicSlise,
		DBHost:           DB_Host,
		DBPort:           DB_Port,
		PostgresDBName:   PostgresDB_Name,
		PostgresUser:     PostgresDB_User,
		PostgresPassword: PostgresDB_Password,
		DBSSLMOD:         DB_SSL_MODE,
	}, nil
}
