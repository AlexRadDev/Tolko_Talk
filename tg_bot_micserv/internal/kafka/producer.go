// Файл producer.go реализует Kafka-продюсер для отправки сообщений в брокер Kafka.
// Отвечает за настройку соединения с Kafka и асинхронную отправку сообщений в указанный топик

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
	Writer *kafka.Writer
}

// NewProducer создаёт новый Kafka-продюсер
func NewProducer(portKafka string, nameTopicKafka string) *Producer {
	//const lblNewProducer = "internal/infrastructure/kafka/producer.go/NewProducer()"
	//logger := logger.NewColorLogger(lblNewProducer)

	// Создаём объект Kafka-продюсера
	writer := &kafka.Writer{
		Addr:     kafka.TCP(portKafka), // Указываем адреса брокеров Kafka
		Topic:    nameTopicKafka,       // Указываем топик для отправки сообщений
		Balancer: &kafka.LeastBytes{},  // Используем балансировку по ключу для равномерного распределения сообщений
		Async:    true,                 // Включаем асинхронную отправку сообщений
	}

	return &Producer{Writer: writer}
}

// SendMessage отправляет сообщение в Kafka
func (p *Producer) SendMessage(nameTopic string, data []byte) error {
	const lblSendMessage = "internal/infrastructure/kafka/producer.go/SendMessage()"
	myLogger := logger.NewColorLogger(lblSendMessage)

	// 1. Проверка существования топика
	exists, err := p.topicExists(nameTopic)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка при проверке существования топика %s: %v", nameTopic, err))
		return fmt.Errorf("проверка существования топика: %w", err)
	}

	// 2. Если топик не существует, создаём его
	if !exists {
		myLogger.Info(fmt.Sprintf("Топик %s не существует, создаём новый", nameTopic))
		err = p.createTopic(nameTopic)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка при создании топика %s: %v", nameTopic, err))
			return fmt.Errorf("создание топика: %w", err)
		}
		myLogger.Info(fmt.Sprintf("Топик %s успешно создан", nameTopic))
	}

	msg := kafka.Message{
		//Topic: nameTopic, // Указываем топик
		Value: data, // Передаём сериализованные данные
	}

	// Отправляем сообщение в Kafka
	err = p.Writer.WriteMessages(context.Background(), msg)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка отправки сообщения в Kafka: %v, Name Topic: %v", err, nameTopic))
		return err
	}

	return nil
}

// topicExists проверяет, существует ли топик в кластере Kafka.
func (p *Producer) topicExists(topic string) (bool, error) {
	// Подключаемся к контроллеру Kafka
	conn, err := kafka.Dial("tcp", p.Writer.Addr.String())
	if err != nil {
		return false, fmt.Errorf("подключение к Kafka: %w", err)
	}
	defer conn.Close()

	// Получаем список топиков
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return false, fmt.Errorf("чтение партиций: %w", err)
	}

	// Проверяем, есть ли топик в списке
	for _, p := range partitions {
		if p.Topic == topic {
			return true, nil
		}
	}
	return false, nil
}

// createTopic создаёт новый топик с заданной конфигурацией.
func (p *Producer) createTopic(topic string) error {
	// Подключаемся к контроллеру Kafka
	conn, err := kafka.Dial("tcp", p.Writer.Addr.String())
	if err != nil {
		return fmt.Errorf("подключение к Kafka: %w", err)
	}
	defer conn.Close()

	// Конфигурация топика
	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     2, // Кол-во партиций
		ReplicationFactor: 3, // Репликация на все брокеры
	}

	// Создаём топик
	err = conn.CreateTopics(topicConfig)
	if err != nil {
		return fmt.Errorf("создание топика %s: %w", topic, err)
	}

	return nil
}

// Close закрывает соединение с Kafka
func (p *Producer) Close() error {
	err := p.Writer.Close()
	if err != nil {
		slog.Error("Ошибка закрытия Kafka-продюсера", "error", err)
		return err
	}
	slog.Info("Kafka-продюсер успешно закрыт")
	return nil
}
