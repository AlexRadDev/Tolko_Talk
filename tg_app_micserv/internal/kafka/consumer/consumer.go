package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"tg_app_micserv/internal/config"
	"tg_app_micserv/internal/kafka/producer"
	"tg_app_micserv/internal/model/model_kafka"
	"tg_app_micserv/internal/service_parser"
	"tg_app_micserv/internal/tg_init_parser"
	"tg_app_micserv/internal/tg_session_storage"
	"tg_app_micserv/tools/logger"
)

// Consumer представляет Kafka-консьюмер.
type Consumer struct {
	Reader *kafka.Reader
}

// NewConsumer создаёт новый экземпляр Kafka-консьюмера.
func NewConsumer(cfg *config.Config) *Consumer { // serviceParser *service_parser.ServiceParser
	const lblNewConsumer = "tg_app_micserv/internal/kafka/consumer/consumer.go/NewConsumer()"
	mylogger := logger.NewColorLogger(lblNewConsumer)

	// Конфигурация Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.KafkaPort,     // Список брокеров
		Topic:       cfg.KafkaTopic_01, // Топик для чтения
		GroupID:     "group_01",        // ID группы консьюмеров
		StartOffset: kafka.FirstOffset, // читать с начала
		MinBytes:    10e3,              // Минимальный размер батча (10KB)
		MaxBytes:    10e6,              // Максимальный размер батча (10MB)
	})

	mylogger.Info(fmt.Sprintf("Kafka Consumer создан для топика %s с брокерами %v", cfg.KafkaTopic_01, cfg.KafkaPort))
	return &Consumer{
		Reader: reader,
	}
}

// Consume читает сообщения из Kafka и передаёт их в service_parser.PostParser.
func (c *Consumer) Consume(ctx context.Context, cfg *config.Config, sessionStorage *tg_session_storage.FileSessionStorage, producer *producer.Producer, pathSessionJson string) error {
	const lblConsume = "tg_app_micserv/internal/kafka/consumer/consumer.go/Consume()"
	myLogger := logger.NewColorLogger(lblConsume)
	myLogger.Info("Подписались на топик Kafaka: \"tgbot-tgapp\"")

	for {
		// Создаем новый клиент для каждой итерации
		tgClient := tg_init_parser.NewClient(cfg, sessionStorage, pathSessionJson)
		myLogger.Info("Успешно создали tg клиента")
		serviceParser := service_parser.NewServiceParser(tgClient)

		// Читаем сообщение из Kafka
		msg, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка чтения сообщения из Kafka: %v", err))
			return fmt.Errorf("чтение сообщения: %w", err)
		}
		myLogger.Info(fmt.Sprintf("Получено сообщение из топика %s: %s", msg.Topic, string(msg.Value)))

		var messageFromTGB *model_kafka.KafkaMessageFromTGB // Структура для Unmarshal сообщения от ТГ-бота из Kafka
		if err := json.Unmarshal(msg.Value, &messageFromTGB); err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка десериализации сообщения: %v", err))
			continue // Пропускаем некорректное сообщение
		}

		timePeriod := time.Duration(messageFromTGB.TimePeriod) * time.Hour

		// Передаём данные в service_parser.PostParser
		textForAudio, err1 := serviceParser.PostParser(ctx, messageFromTGB.NameChanel, timePeriod)
		if err1 != nil {
			myLogger.Error(fmt.Sprintf("Ошибка обработки сообщения в PostParser: %v", err1))
			// Продолжаем чтение, чтобы не останавливать консьюмер
			continue
		}
		//myLogger.Info(fmt.Sprintf("Мапа textForAudio: %v", textForAudio))

		// Проверяем спарсили ли мы текст
		if len(textForAudio) == 0 {
			// Новостей за этот период нет
			// TODO: Реализовать отправку обратно в TG сообщение что постов за этот период нет
		} else {
			// отправляем мапу с текстом дольше в kafka в tgapp-texttospeech
			request := model_kafka.InitKafkaRequestForTTS(messageFromTGB, textForAudio)

			jsonData, err01 := json.Marshal(request)
			if err01 != nil {
				myLogger.Error(fmt.Sprintf("Ошибка сериализации JSON: %v", err01))
			}

			err01 = producer.SendMessage(ctx, cfg.KafkaTopic_02, jsonData)
			if err01 != nil {
				myLogger.Error(fmt.Sprintf("Не удалось отправить сообщение в Kafka(tgapp-texttospeech): %v", err01))
			}
		}
		myLogger.Info("Сообщение в Kafka(tgapp-texttospeech) успешно отправлено")

		myLogger.Info("Завершение Telegram-клиента")
		if err02 := tgClient.Close(ctx); err != nil {
			myLogger.Error("Ошибка при закрытии Telegram-клиента", "error", err02)
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
