package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"

	"text_to_speech_app/internal/config"
	"text_to_speech_app/internal/kafka/producer"
	"text_to_speech_app/internal/metrics/metrics_prometheus"
	"text_to_speech_app/internal/model/model_kafka"
	"text_to_speech_app/internal/text_to_audio/interface_text_to_audio"
	"text_to_speech_app/tools/logger"
)

// Consumer представляет Kafka-консьюмер.
type Consumer struct {
	Reader            *kafka.Reader
	Producer          *producer.Producer
	TextToAudio       interface_text_to_audio.InterfaceTextToAudio
	MetricsPrometheus *metrics_prometheus.TGBotMetricsPrometheus
}

// NewConsumer создаёт новый экземпляр Kafka-консьюмера.
func NewConsumer(cfg *config.Config, producerKafka *producer.Producer,
	interfaceTextToAudio interface_text_to_audio.InterfaceTextToAudio,
	metricsPrometheus *metrics_prometheus.TGBotMetricsPrometheus) *Consumer { // serviceParser *service_parser.ServiceParser

	// Конфигурация Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.KafkaPort,          // Список брокеров
		Topic:       cfg.TopicConsum,        // Топик для чтения
		GroupID:     "group_01",             // ID группы консьюмеров
		StartOffset: kafka.FirstOffset,      // читать с начала
		MinBytes:    1,                      // Минимальный размер батча (10KB)
		MaxBytes:    10e6,                   // Максимальный размер батча (10MB)
		MaxWait:     100 * time.Millisecond, // ждать не более 100мс
	})

	return &Consumer{
		Reader:            reader,
		Producer:          producerKafka,
		TextToAudio:       interfaceTextToAudio,
		MetricsPrometheus: metricsPrometheus,
	}
}

// Consume читает сообщения из Kafka и передаёт их в service_parser.PostParser.
func (c *Consumer) Consume(ctx context.Context, cfg *config.Config) error {
	const lblConsume = "text_to_speech_micserv/internal/kafka/consumer/Consume()"
	myLogger := logger.NewColorLogger(lblConsume)
	myLogger.Info("Подписались на топик Kafaka: \"tgapp-texttospeech\"")

	for {
		// Читаем сообщение из Kafka
		msg, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка чтения сообщения из Kafka: %v", err))
			return fmt.Errorf("чтение сообщения: %w", err)
		}
		myLogger.Info(fmt.Sprintf("Получено сообщение из топика %s: %s", msg.Topic, string(msg.Value)))

		var messageFromTGAPP *model_kafka.KafkaMessageFromTGAPP

		if err := json.Unmarshal(msg.Value, &messageFromTGAPP); err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка десериализации сообщения: %v", err))
			continue
		}

		// Фиксируем значения для метрик в Prometheus
		c.MetricsPrometheus.RequestID = messageFromTGAPP.RequestID
		c.MetricsPrometheus.ChatID = messageFromTGAPP.ChatID
		c.MetricsPrometheus.TimeStartRequest = time.Now()

		responseForTGAPP, err1 := c.TextToAudio.Synthesize(ctx, messageFromTGAPP)
		if err1 != nil {
			myLogger.Error(fmt.Sprintf("Ошибка обработки сообщения в PostParser: %v", err1))
			continue
		}

		jsonData, err02 := json.Marshal(responseForTGAPP)
		if err02 != nil {
			myLogger.Error(fmt.Sprintf("Ошибка сериализации JSON: %v", err02))
		}

		err03 := c.Producer.SendMessage(ctx, cfg.TopicProdus, jsonData)
		if err03 != nil {
			myLogger.Error(fmt.Sprintf("Не удалось отправить сообщение в Kafka(tgapp-texttospeech): %v", err03))
		}

		strChatID := strconv.FormatInt(c.MetricsPrometheus.ChatID, 10)
		c.MetricsPrometheus.ObserveDuration(c.MetricsPrometheus.RequestID, strChatID, c.MetricsPrometheus.TimeRequestCalculation())
		myLogger.Info("Сообщение в Kafka(texttospeech-tgapp) успешно отправлено")
	}

	return nil
}

// Close закрывает Kafka Reader.
func (c *Consumer) Close() error {
	const lblClose = "text_to_speech_micserv/internal/kafka/consumer/Close()"
	mylogger := logger.NewColorLogger(lblClose)

	if err := c.Reader.Close(); err != nil {
		mylogger.Error(fmt.Sprintf("Ошибка закрытия Kafka Reader: %v", err))
		return fmt.Errorf("закрытие Reader: %w", err)
	}
	mylogger.Info("Kafka Consumer успешно закрыт")
	return nil
}
