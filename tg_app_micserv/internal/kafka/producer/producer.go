package producer

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"

	"tg_app_micserv/internal/config"
	"tg_app_micserv/tools/logger"
)

// MessageProducer определяет интерфейс для отправки сообщений в Kafka
type MessageProducer interface {
	ProduceMessages(ctx context.Context, messages map[int64]string) error
}

// Структура Producer содержит Kafka-продюсер
type Producer struct {
	Writer *kafka.Writer // Объект для записи сообщений в Kafka
}

// NewProducer создаёт новый Kafka-продюсер
func NewProducer(cfg *config.Config) *Producer { // (portKafka string, nameTopicKafka string)
	writer := &kafka.Writer{
		Addr:       kafka.TCP(cfg.KafkaPort[0]), // Указываем адреса брокеров Kafka
		Topic:      cfg.KafkaTopic_02,           // Указываем топик для отправки сообщений
		Balancer:   &kafka.LeastBytes{},         // Используем балансировку по ключу для равномерного распределения сообщений
		Async:      true,                        // Включаем асинхронную отправку сообщений
		BatchBytes: 5 * 1024 * 1024,             // 5 MB
	}

	return &Producer{Writer: writer}
}

// SendMessage отправляет сообщение в Kafka
func (p *Producer) SendMessage(ctx context.Context, nameTopic string, data []byte) error {
	const lblSendMessage = "internal/kafka/producer/producer.go/SendMessage()"
	myLogger := logger.NewColorLogger(lblSendMessage)

	// Проверка существования топика
	exists, err := p.topicExists(ctx, nameTopic)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка при проверке существования топика %s: %v", nameTopic, err))
		return fmt.Errorf("проверка существования топика: %w", err)
	}

	// Если топик не существует, создаём его
	if !exists {
		myLogger.Info(fmt.Sprintf("Топик %s не существует, создаём новый", nameTopic))
		err = p.createTopic(ctx, nameTopic)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка при создании топика %s: %v", nameTopic, err))
			return fmt.Errorf("создание топика: %w", err)
		}
		myLogger.Info(fmt.Sprintf("Топик %s успешно создан", nameTopic))
	}

	msg := kafka.Message{
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
func (p *Producer) topicExists(ctx context.Context, topic string) (bool, error) {
	const lbltopicExists = "internal/kafka/producer/producer.go/topicExists()"
	myLogger := logger.NewColorLogger(lbltopicExists)

	// Получаем соединение с контроллером
	ctrlConn, err := p.getControllerConn(ctx)
	if err != nil {
		return false, err
	}
	defer ctrlConn.Close()

	// Получаем список топиков
	partitions, err := ctrlConn.ReadPartitions()
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка чтения партиций: %v", err))
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
func (p *Producer) createTopic(ctx context.Context, topic string) error {
	const lblcreateTopic = "internal/kafka/producer/producer.go/topicExists()"
	myLogger := logger.NewColorLogger(lblcreateTopic)

	// Получаем соединение с контроллером
	ctrlConn, err := p.getControllerConn(ctx)
	if err != nil {
		return err
	}
	defer ctrlConn.Close()

	// Конфигурация топика
	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     2, // Кол-во партиций
		ReplicationFactor: 3, // Репликация на все брокеры
		ConfigEntries: []kafka.ConfigEntry{ // Максимальный размер сообщения в топике
			{
				ConfigName:  "max.message.bytes",
				ConfigValue: "10485760", // 10MB
			},
		},
	}

	// Создаём топик
	err = ctrlConn.CreateTopics(topicConfig)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка создания топика: %v", err))
		return fmt.Errorf("создание топика %s: %w", topic, err)
	}

	// Ждем 3 секунды для распространения метаданных о новом топике
	time.Sleep(3 * time.Second)

	return nil
}

// getControllerConn устанавливает соединение с контроллером Kafka
func (p *Producer) getControllerConn(ctx context.Context) (*kafka.Conn, error) {
	const lbl = "internal/kafka/producer/producer.go/getControllerConn()"
	myLogger := logger.NewColorLogger(lbl)

	// Подключаемся к Kafka
	conn, err := kafka.DialContext(ctx, "tcp", p.Writer.Addr.String())
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка подключения к Kafka: %v", err))
		return nil, fmt.Errorf("подключение к Kafka: %w", err)
	}

	// Получаем контроллера кластера
	controller, err := conn.Controller()
	if err != nil {
		conn.Close()
		myLogger.Error(fmt.Sprintf("Не удалось получить контроллер: %v", err))
		return nil, fmt.Errorf("получение контроллера: %w", err)
	}

	controllerAddr := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	myLogger.Info(fmt.Sprintf("Найден контроллер: %v", controllerAddr))

	// Подключаемся к контроллеру
	ctrlConn, err := kafka.DialContext(ctx, "tcp", controllerAddr)
	if err != nil {
		conn.Close() // Закрываем первое соединение при ошибке
		myLogger.Error(fmt.Sprintf("Не удалось подключиться к контроллеру: %v", err))
		return nil, fmt.Errorf("подключение к контроллеру: %w", err)
	}

	conn.Close()

	return ctrlConn, nil
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
