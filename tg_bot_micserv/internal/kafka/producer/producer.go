package producer

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"

	"tg_bot/internal/config"
	"tg_bot/tools/logger"
)

// Producer структура Kafka-продюсера
type Producer struct {
	Writer *kafka.Writer // Объект для записи сообщений в Kafka.
}

// NewProducer создаёт новый Kafka-продюсер
func NewProducer(cfg *config.Config) *Producer {
	writer := &kafka.Writer{
		Addr:       kafka.TCP(cfg.KafkaPort[0]), // Указываем адреса брокеров Kafka
		Topic:      cfg.NameTopic01,             // Указываем топик для отправки сообщений по умолчанию
		Balancer:   &kafka.LeastBytes{},         // Используем балансировку по ключу для равномерного распределения сообщений
		Async:      true,                        // Включаем асинхронную отправку сообщений
		BatchBytes: 10 * 1024 * 1024,            // 10 MB
	}

	return &Producer{Writer: writer}
}

// SendMessage отправляет сообщение в топик Kafka.
func (p *Producer) SendMessage(ctx context.Context, nameTopic string, data []byte) error {
	const lblSendMessage = "[myMicro] tg_bot_micserv/internal/kafka/producer/SendMessage()"
	myLogger := logger.NewColorLogger(lblSendMessage)

	// Проверяем, существует ли топик в кластере Kafka и создаем если нет
	err := p.CheckingAndCreatingTopics(ctx, nameTopic)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Value: data,
	}

	// Отправляем сообщение в Kafka
	err = p.Writer.WriteMessages(context.Background(), msg)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка отправки сообщения в Kafka: %v, Name Topic: %v", err, nameTopic))
		return err
	}

	return nil
}

// CheckingAndCreatingTopics проверяет, существует ли топик в кластере Kafka и создает если нет.
func (p *Producer) CheckingAndCreatingTopics(ctx context.Context, nameTopic string) error {
	const lbltopicExists = "[myMicro] tg_bot_micserv/internal/kafka/producer/CheckingAndCreatingTopics()"
	myLogger := logger.NewColorLogger(lbltopicExists)

	// Проверка существования топика
	exists, err := p.topicExists(ctx, nameTopic)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка при проверке существования топика %s: %v", nameTopic, err))
		return fmt.Errorf("Ошибка при проверке существования топика: %w", err)
	}

	// Если топик не существует, создаём его
	if !exists {
		myLogger.Info(fmt.Sprintf("Топик %s не существует, создаём новый", nameTopic))
		err = p.CreateTopic(ctx, nameTopic)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка при создании топика %s: %v", nameTopic, err))
			return fmt.Errorf("Ошибка при создании топика: %w", err)
		}
		myLogger.Info(fmt.Sprintf("Топик %s успешно создан", nameTopic))
	}

	return nil
}

// topicExists проверяет, существует ли топик в кластере Kafka.
func (p *Producer) topicExists(ctx context.Context, topic string) (bool, error) {
	const lbltopicExists = "[myMicro] tg_bot_micserv/internal/kafka/producer/topicExists()"
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
	for _, val := range partitions {
		if val.Topic == topic {
			return true, nil
		}
	}
	return false, nil
}

// CreateTopic создаёт новый топик, с заданной конфигурацией.
func (p *Producer) CreateTopic(ctx context.Context, topic string) error {
	const lblcreateTopic = "[myMicro] tg_bot_micserv/internal/kafka/producer/CreateTopic()"
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

	// Проверяем, что топик создан
	exists, err := p.topicExists(ctx, topic)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка проверки существования топика после создания: %v", err))
		return fmt.Errorf("проверка топика после создания: %w", err)
	}
	if !exists {
		return fmt.Errorf("топик %s не был создан", topic)
	}
	myLogger.Info(fmt.Sprintf("Топик %s успешно создан", topic))

	return nil
}

// getControllerConn устанавливает соединение с контроллером Kafka
func (p *Producer) getControllerConn(ctx context.Context) (*kafka.Conn, error) {
	const lbl = "[myMicro] tg_bot_micserv/internal/kafka/producer/getControllerConn()"
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
		conn.Close() // Закрываем соединение при ошибке
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
	const lbl = "[myMicro] tg_bot_micserv/internal/kafka/producer/Close()"
	myLogger := logger.NewColorLogger(lbl)

	err := p.Writer.Close()
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка закрытия Kafka-продюсера %v", err))
		return err
	}
	myLogger.Info("Kafka-продюсер успешно закрыт")
	return nil
}
