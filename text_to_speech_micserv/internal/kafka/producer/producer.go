// Файл producer.go реализует Kafka-продюсер для отправки сообщений в брокер.
// Отвечает за настройку соединения с Kafka и асинхронную отправку сообщений в указанный топик

package producer

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"

	"text_to_speech_app/tools/logger"
)

// Producer содержит Kafka-продюсер
type Producer struct {
	writer *kafka.Writer // объект для записи сообщений в Kafka
}

// NewProducer создаёт новый Kafka-продюсер
func NewProducer(portKafka []string, nameTopic string) *Producer {

	// Создаём объект Kafka-продюсера
	writer := &kafka.Writer{
		Addr:     kafka.TCP(portKafka...), // Указываем адреса брокеров Kafka
		Topic:    nameTopic,               // Указываем топик для отправки сообщений
		Balancer: &kafka.LeastBytes{},     // Используем балансировку по ключу
		Async:    true,                    // Включаем асинхронную отправку
	}

	return &Producer{writer: writer}
}

// SendMessage отправляет сообщение в Kafka
func (p *Producer) SendMessage(nameTopic string, data []byte) error {
	const lblNewSendMessage = "internal/infrastructure/kafka/producer.go"
	myLogger := logger.NewColorLogger(lblNewSendMessage)

	// Создаём сообщение для Kafka
	msg := kafka.Message{
		Topic: nameTopic,
		Value: data,
	}

	// Отправляем сообщение в Kafka
	err := p.writer.WriteMessages(context.Background(), msg)
	if err != nil {
		myLogger.Error("Ошибка отправки сообщения в Kafka", slog.Any("error", err), slog.String("nameTopic", nameTopic))
		return err
	}
	myLogger.Info("Сообщение успешно отправлено в Kafka", slog.String("nameTopic", nameTopic))
	return nil
}

// Close закрывает соединение с Kafka
func (p *Producer) Close() error {
	const lblClose = "internal/infrastructure/kafka/producer.go"
	myLogger := logger.NewColorLogger(lblClose)

	// Закрываем writer
	err := p.writer.Close()
	if err != nil {
		myLogger.Error("Ошибка закрытия Kafka-продюсера", slog.Any("error", err))
		return err
	}
	myLogger.Info("Kafka-продюсер успешно закрыт")
	return nil
}
