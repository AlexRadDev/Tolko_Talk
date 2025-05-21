package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"tg_app_micserv/config"
	"tg_app_micserv/tools/logger"
)

// MessageProducer определяет интерфейс для отправки сообщений в Kafka
type MessageProducer interface {
	ProduceMessages(ctx context.Context, messages map[int64]string) error
}

// Producer реализует отправку сообщений в Kafka
type Producer struct {
	writer *kafka.Writer
}

// NewProducer создает новый Kafka Producer
func NewProducer(cfg *config.Config) (*Producer, error) {
	kafkaPort := cfg.KafkaPort
	nameTopicKafka := cfg.KafkaTopic

	// Создание нового экземпляра kafka.Writer для записи сообщений в Kafka
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(strings.Split(kafkaPort, ",")...), // Установка адресов брокеров Kafka
		Topic:                  nameTopicKafka,                              // Указание названия Kafka-топика, в который будут отправляться сообщения
		Balancer:               &kafka.LeastBytes{},                         // Установка стратегии балансировки нагрузки: LeastBytes выбирает партицию с наименьшим количеством данных
		MaxAttempts:            3,                                           // Максимальное количество попыток записи сообщения при неудаче
		WriteTimeout:           1 * time.Second,                             // Таймаут для попытки записи сообщения
		RequiredAcks:           kafka.RequireOne,                            // Требование подтверждения получения сообщения хотя бы одним брокером
		Async:                  false,                                       // Указывает, что запись будет происходить синхронно (false — запись блокирует до завершения)
		AllowAutoTopicCreation: true,                                        // Автоматически создавать топик, если он не существует (требует поддержки со стороны брокера)
	}

	// Возвращаем указатель на новый объект Producer с созданным writer
	return &Producer{
		writer: writer,
	}, nil
}

// ProduceMessages отправляет карту сообщений в Kafka
func (p *Producer) ProduceMessages(ctx context.Context, mapTextForAudio map[int64]string) error {
	const lbl = "tg_app_micserv/internal/kafka/producer.go/ProduceMessages()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	var kafkaMessages []kafka.Message
	// Преобразуем карту в сообщения Kafka
	for timestamp, text := range mapTextForAudio {
		value, err := json.Marshal(map[string]interface{}{
			"timestamp": timestamp,
			"text":      text,
		})
		if err != nil {
			slog.Error(fmt.Sprintf("Ошибка при сериализации сообщения, error: %v, timestamp: %v", err, timestamp))
			continue
		}

		kafkaMessages = append(kafkaMessages, kafka.Message{
			Key:   []byte(fmt.Sprintf("%d", timestamp)),
			Value: value,
		})
	}
	slog.Info("Карта постов для Kafka заполнена")

	if len(kafkaMessages) == 0 {
		slog.Error("Карта постов для Kafka пуста")
		return nil
	}

	// Отправляем сообщения в Kafka
	err := p.writer.WriteMessages(ctx, kafkaMessages...)
	if err != nil {
		return fmt.Errorf("ошибка при записи сообщений в Kafka: %w", err)
	}

	//p.logger.Info("Сообщения успешно отправлены в Kafka", "count", len(kafkaMessages))
	return nil
}

// Close закрывает Kafka-продюсер
func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		//p.logger.Error("Ошибка при закрытии Kafka-продюсера", "error", err)
		return fmt.Errorf("ошибка при закрытии Kafka-продюсера: %w", err)
	}
	//p.logger.Info("Kafka-продюсер успешно закрыт")
	return nil
}
