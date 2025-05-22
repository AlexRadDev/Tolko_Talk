package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/segmentio/kafka-go"

	"tg_bot/internal/config"
	"tg_bot/internal/model/model_kafka"
	"tg_bot/tools/logger"
)

// Consumer представляет Kafka-консьюмер.
type Consumer struct {
	Reader *kafka.Reader
}

// NewConsumer создаёт новый экземпляр Kafka-консьюмера.
func NewConsumer(cfg *config.Config) *Consumer { // serviceParser *service_parser.ServiceParser
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.KafkaPort,     // Список брокеров
		Topic:       cfg.NameTopic02,   // Топик для чтения
		GroupID:     "group_01",        // ID группы консьюмеров
		StartOffset: kafka.FirstOffset, // читать с начала
		MinBytes:    10e3,              // Минимальный размер батча (10KB)
		MaxBytes:    10e6,              // Максимальный размер батча (10MB)
	})

	return &Consumer{
		Reader: reader,
	}
}

// Consume читает сообщения из Kafka и передаёт их в service_parser.PostParser.
func (c *Consumer) Consume(ctx context.Context, tgBot *tgbotapi.BotAPI) error {
	const lblConsume = "tg_bot_micserv/internal/kafka/consumer/consumer.go/Consume()"
	myLogger := logger.NewColorLogger(lblConsume)

	myLogger.Info("Подписались на топик Kafaka: \"texttospeech-tgbot\"")
	for {
		// Читаем сообщение из Kafka
		msg, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка чтения сообщения из Kafka: %v", err))
			return fmt.Errorf("чтение сообщения: %w", err)
		}
		myLogger.Info(fmt.Sprintf("Получено сообщение из топика: %s", msg.Topic))

		var messageFromTGB *model_kafka.KafkaRequestForTGBOT // Структура для Unmarshal сообщения от TextToSpeech из Kafka
		if err := json.Unmarshal(msg.Value, &messageFromTGB); err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка десериализации сообщения: %v", err))
			continue
		}

		// Создаём аудио сообщение для Telegram
		audio := tgbotapi.NewAudio(messageFromTGB.ChatID, tgbotapi.FileBytes{
			Name:  "audio.mp3", // Имя файла для Telegram
			Bytes: messageFromTGB.AudioData,
		})

		// Отправляем аудио в Telegram
		if _, err := tgBot.Send(audio); err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка отправки аудио в Telegram, chatID: %d: %v", messageFromTGB.ChatID, err))
			continue
		}
		myLogger.Info(fmt.Sprintf("Аудио отправлено в Telegram, chatID: %d", messageFromTGB.ChatID))
	}
}

// Close закрывает Kafka Reader.
func (c *Consumer) Close() error {
	const lblClose = "tg_bot_micserv/internal/kafka/consumer/consumer.go/Close()"
	mylogger := logger.NewColorLogger(lblClose)

	if err := c.Reader.Close(); err != nil {
		mylogger.Error(fmt.Sprintf("Ошибка закрытия Kafka Reader: %v", err))
		return fmt.Errorf("закрытие Reader: %w", err)
	}
	mylogger.Info("Kafka Consumer успешно закрыт")
	return nil
}
