// Файл producer.go реализует Kafka-продюсер для отправки сообщений в брокер Kafka.
// Отвечает за настройку соединения с Kafka и асинхронную отправку сообщений в указанный топик.
// Соответствует инфраструктурному слою чистой архитектуры, изолируя работу с Kafka от бизнес-логики.

package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"tg_bot/tools/logger"
)

// Структура Producer содержит Kafka-продюсер
type Producer struct {
	// writer — объект для записи сообщений в Kafka
	writer *kafka.Writer
}

// NewProducer создаёт новый Kafka-продюсер
func NewProducer(portKafka []string, nameTopicKafka string) *Producer {
	//const lblNewProducer = "internal/infrastructure/kafka/producer.go/NewProducer()"
	//logger := logger.NewColorLogger(lblNewProducer)

	// Создаём объект Kafka-продюсера
	writer := &kafka.Writer{
		Addr:     kafka.TCP(portKafka...), // Указываем адреса брокеров Kafka
		Topic:    nameTopicKafka,          // Указываем топик для отправки сообщений
		Balancer: &kafka.LeastBytes{},     // Используем балансировку по ключу для равномерного распределения сообщений
		Async:    true,                    // Включаем асинхронную отправку сообщений
	}

	// Возвращаем новый объект Producer
	return &Producer{writer: writer}
}

// SendMessage отправляет сообщение в Kafka
func (p *Producer) SendMessage(nameTopic string, data []byte) error {
	const lblSendMessage = "internal/infrastructure/kafka/producer.go/SendMessage()"
	myLogger := logger.NewColorLogger(lblSendMessage)
	//slog.SetDefault(logger)

	msg := kafka.Message{
		// Указываем топик
		Topic: nameTopic,
		// Передаём сериализованные данные
		Value: data,
	}
	// Отправляем сообщение в Kafka
	err := p.writer.WriteMessages(context.Background(), msg)
	if err != nil {
		// Логируем ошибку отправки сообщения
		myLogger.Error(fmt.Sprintf("Ошибка отправки сообщения в Kafka: %v, Name Topic: %v", err, nameTopic))
		// Возвращаем ошибку
		return err
	}

	return nil
}

// Close закрывает соединение с Kafka
func (p *Producer) Close() error {
	// Закрываем writer Kafka-продюсера
	err := p.writer.Close()
	if err != nil {
		// Логируем ошибку закрытия соединения
		slog.Error("Ошибка закрытия Kafka-продюсера", "error", err)
		// Возвращаем ошибку
		return err
	}
	// Логируем успешное закрытие соединения
	slog.Info("Kafka-продюсер успешно закрыт")
	// Возвращаем nil, так как закрытие успешно
	return nil
}
