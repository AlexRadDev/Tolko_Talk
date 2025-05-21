package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"tg_app_micserv/internal/config"
	"tg_app_micserv/internal/service_parser"
	"tg_app_micserv/internal/tg_init_parser"
	"tg_app_micserv/internal/tg_session_storage"
	"tg_app_micserv/tools/logger"
)

// Consumer представляет Kafka-консьюмер.
type Consumer struct {
	Reader *kafka.Reader
}

// KafkaRequest представляет структуру данных сообщения из Kafka.
type KafkaRequest struct {
	ChatID               int64   `json:"ChatID"`               // идентификатор чата Telegram
	NameChanel           string  `json:"NameChanel"`           // имя или ссылка на Telegram-канал
	SpeakingRate         float64 `json:"SpeakingRate"`         // скорость речи
	TimePeriod           int     `json:"TimePeriod"`           // период времени в часах
	AwaitingChannelInput bool    `json:"AwaitingChannelInput"` // флаг, указывающий, ожидается ли ввод имени канала
}

// NewConsumer создаёт новый экземпляр Kafka-консьюмера.
func NewConsumer(cfg *config.Config) *Consumer { // serviceParser *service_parser.ServiceParser
	const lblNewConsumer = "tg_app_micserv/internal/kafka/consumer/consumer.go/NewConsumer()"
	mylogger := logger.NewColorLogger(lblNewConsumer)

	// Конфигурация Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.KafkaPort,          // Список брокеров
		Topic:       cfg.KafkaTopic,         // Топик для чтения
		GroupID:     "tg_app_parser_group4", // ID группы консьюмеров
		StartOffset: kafka.FirstOffset,      // читать с начала
		MinBytes:    10e3,                   // Минимальный размер батча (10KB)
		MaxBytes:    10e6,                   // Максимальный размер батча (10MB)
	})

	mylogger.Info(fmt.Sprintf("Kafka Consumer создан для топика %s с брокерами %v", cfg.KafkaTopic, cfg.KafkaPort))
	return &Consumer{
		Reader: reader,
		//ServiceParser: serviceParser,
	}
}

// Consume читает сообщения из Kafka и передаёт их в service_parser.PostParser.
func (c *Consumer) Consume(ctx context.Context, cfg *config.Config, sessionStorage *tg_session_storage.FileSessionStorage) error {
	const lblConsume = "tg_app_micserv/internal/kafka/consumer/consumer.go/Consume()"
	mylogger := logger.NewColorLogger(lblConsume)
	mylogger.Info("Подписались на топик Kafaka: \"tgbot-tgapp\"")

	for {
		// Создаем новый клиент для каждой итерации
		tgClient := tg_init_parser.NewClient(cfg.API_ID, cfg.API_Hash, cfg.Phone, cfg.Two_F_Password, sessionStorage)
		serviceParser := service_parser.NewServiceParser(tgClient)

		// Читаем сообщение из Kafka
		msg, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			mylogger.Error(fmt.Sprintf("Ошибка чтения сообщения из Kafka: %v", err))
			return fmt.Errorf("чтение сообщения: %w", err)
		}
		mylogger.Info(fmt.Sprintf("Получено сообщение из топика %s: %s", msg.Topic, string(msg.Value)))

		var request KafkaRequest
		if err := json.Unmarshal(msg.Value, &request); err != nil {
			mylogger.Error(fmt.Sprintf("Ошибка десериализации сообщения: %v", err))
			continue // Пропускаем некорректное сообщение
		}

		timePeriod := time.Duration(request.TimePeriod) * time.Hour

		// Передаём данные в service_parser.PostParser
		textForAudio, err1 := serviceParser.PostParser(ctx, request.NameChanel, timePeriod)
		if err1 != nil {
			mylogger.Error(fmt.Sprintf("Ошибка обработки сообщения в PostParser: %v", err1))
			// Продолжаем чтение, чтобы не останавливать консьюмер
			continue
		}

		mylogger.Info(fmt.Sprintf("!!! Сообщение успешно обработано PostParser, МАПА ТЕКСТ ДЛЯ АУДИО → textForAudio = %v", textForAudio))

		mylogger.Info("Завершение Telegram-клиента")
		if err := tgClient.Close(ctx); err != nil {
			mylogger.Error("Ошибка при закрытии Telegram-клиента", "error", err)
		}
	}
}

// Close закрывает Kafka Reader.
func (c *Consumer) Close() error {
	const lblClose = "tg_app_micserv/internal/kafka/consumer/consumer.go/Consume()"
	mylogger := logger.NewColorLogger(lblClose)

	if err := c.Reader.Close(); err != nil {
		mylogger.Error(fmt.Sprintf("Ошибка закрытия Kafka Reader: %v", err))
		return fmt.Errorf("закрытие Reader: %w", err)
	}
	mylogger.Info("Kafka Consumer успешно закрыт")
	return nil
}
