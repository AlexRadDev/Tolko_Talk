// Файл consumer.go реализует Kafka-консьюмер для получения запросов.
// Отвечает за чтение сообщений из Kafka, их обработку через бизнес-логику и отправку результатов через продюсер.

package consumer

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"text_to_speech_app/internal/app_text_to_speech"

	"text_to_speech_app/internal/model/model_text_to_speech"
	"text_to_speech_app/tools/logger"
)

// Consumer содержит Kafka-консьюмер
type Consumer struct {
	reader *kafka.Reader // объект для чтения сообщений из Kafka
}

// NewConsumer создаёт новый Kafka-консьюмер
func NewConsumer(brokers []string, nameTopic, groupID string) *Consumer {

	// Создаём объект Kafka-консьюмера
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,   // Указываем адреса брокеров Kafka
		Topic:    nameTopic, // Указываем топик
		GroupID:  groupID,   // Указываем идентификатор группы
		MinBytes: 10e3,      // 10KB - Устанавливаем минимальное количество байт для чтения
		MaxBytes: 10e6,      // 10MB - Устанавливаем максимальное количество байт для чтения
	})

	return &Consumer{reader: reader}
}

// Consume читает сообщения из Kafka и обрабатывает их
func (c *Consumer) Consume(ctx context.Context, ttsService *app_text_to_speech.Service) error {
	const lblConsumer = "internal/infrastructure/kafka/consumer.go"
	myLogger := logger.NewColorLogger(lblConsumer)

	// Бесконечный цикл для чтения сообщений
	for {
		// Читаем сообщение из Kafka
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			myLogger.Error("Ошибка чтения сообщения из Kafka", slog.Any("error", err))
			return err
		}
		myLogger.Info("Получено сообщение из Kafka", slog.String("topic", msg.Topic))

		// Создаём структуру для запроса
		var req model_text_to_speech.TextToSpeechRequest
		if err := json.Unmarshal(msg.Value, &req); err != nil {
			myLogger.Error("Ошибка десериализации сообщения", slog.Any("error", err))
			continue
		}
		myLogger.Info("Успешно десериализовали сообщение", slog.Any("request", req))

		// Вызываем бизнес-логику для синтеза речи
		resp, err := ttsService.Synthesize(ctx, &req)
		if err != nil {
			myLogger.Error("Ошибка синтеза речи", slog.Any("error", err))
			continue
		}
		myLogger.Info("Успешно обработали запрос")

		// Сериализуем ответ
		respData, err := json.Marshal(resp)
		if err != nil {
			myLogger.Error("Ошибка сериализации ответа", slog.Any("error", err))
			continue
		}

		// Отправляем ответ через продюсер
		if err := ttsService.KafkaProducer.SendMessage("text-to-speech-responses", respData); err != nil {
			myLogger.Error("Ошибка отправки ответа в Kafka", slog.Any("error", err))
			continue
		}
		myLogger.Info("Успешно отправили ответ в Kafka")
	}
}

// Close закрывает соединение с Kafka
func (c *Consumer) Close() error {
	const lblClose = "internal/infrastructure/kafka/consumer.go"
	myLogger := logger.NewColorLogger(lblClose)

	// Закрываем reader
	err := c.reader.Close()
	if err != nil {
		myLogger.Error("Ошибка закрытия Kafka-консьюмера", slog.Any("error", err))
		return err
	}
	myLogger.Info("Kafka-консьюмер успешно закрыт")
	return nil
}
