package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"

	"tg_app_micserv/internal/config"
	"tg_app_micserv/internal/kafka/producer"
	"tg_app_micserv/internal/metrics/metrics_prometheus"
	"tg_app_micserv/internal/model/model_kafka"
	"tg_app_micserv/internal/service_parser"
	"tg_app_micserv/internal/tg_app_init"
	"tg_app_micserv/tools/logger"
)

// Consumer представляет Kafka-консьюмер.
type Consumer struct {
	Reader            *kafka.Reader
	MetricsPrometheus *metrics_prometheus.TGBotMetricsPrometheus
	AppTGPostParser   *tg_app_init.TGAppClient
}

// NewConsumer создаёт новый экземпляр Kafka-консьюмера.
func NewConsumer(cfg *config.Config,
	metricsPrometheus *metrics_prometheus.TGBotMetricsPrometheus,
	appTGPostParser *tg_app_init.TGAppClient) *Consumer {

	const lblNewConsumer = "tg_app_micserv/internal/kafka/consumer/NewConsumer()"
	mylogger := logger.NewColorLogger(lblNewConsumer)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.KafkaPort,          // Список брокеров
		Topic:       cfg.KafkaTopic_01,      // Топик для чтения
		GroupID:     "group_01",             // ID группы консьюмеров
		StartOffset: kafka.FirstOffset,      // читать с начала
		MinBytes:    1,                      // Минимальный размер батча (10KB)
		MaxBytes:    10e6,                   // Максимальный размер батча (10MB)
		MaxWait:     100 * time.Millisecond, // ждать не более 100мс
	})

	mylogger.Info(fmt.Sprintf("Kafka Consumer создан для топика %s с брокерами %v", cfg.KafkaTopic_01, cfg.KafkaPort))
	return &Consumer{
		Reader:            reader,
		MetricsPrometheus: metricsPrometheus,
		AppTGPostParser:   appTGPostParser,
	}
}

// Consume читает сообщения из Kafka и передаёт их в service_parser.PostParser.
func (c *Consumer) Consume(ctx context.Context,
	cfg *config.Config,
	producer *producer.Producer) error {

	const lblConsume = "tg_app_micserv/internal/kafka/consumer/consumer.go/Consume()"
	myLogger := logger.NewColorLogger(lblConsume)
	myLogger.Info("Подписались на топик Kafaka: \"tgbot-tgapp\"")

	// tgClient := tg_init_parser.NewClient(cfg, sessionStorage, pathSessionJson)
	// myLogger.Info("Успешно создали tg клиента")
	// serviceParser := service_parser.NewServiceParser(tgClient)

	for {
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
			continue
		}

		// Фиксируем значения для метрик в Prometheus
		c.MetricsPrometheus.RequestID = messageFromTGB.RequestID
		c.MetricsPrometheus.ChatID = messageFromTGB.ChatID
		c.MetricsPrometheus.TimeStartRequest = time.Now()

		// Форматируем имя канала, превращаем ссылку в нужный формат для TG API PARSER
		channelNameForParserFunc := service_parser.ExtractChannelName(messageFromTGB.NameChanel)

		timePeriod := time.Duration(messageFromTGB.TimePeriod) * time.Hour

		// Парсим посты из ТГ канала
		tgPosts, err := c.AppTGPostParser.PostParser(ctx, channelNameForParserFunc, timePeriod)
		if err != nil {
			myLogger.Error(err.Error())
			return err
		}

		// Форматируем посты в одно сообщение для Google TTS
		textForAudio, err := service_parser.FormationTextForTTS(tgPosts)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка форматирования текста постов: %v", err))
			continue // Продолжаем чтение, чтобы не останавливать консьюмер
		}
		myLogger.Info(fmt.Sprintf("Длинна мапы (textForAudio) равна: %v", len(textForAudio)))
		//myLogger.Info(fmt.Sprintf("Содержимое мапы (textForAudio) равно: %s", textForAudio))

		if len(textForAudio) < 1 || textForAudio == nil { // Если новостей за этот период нет
			textForUsers := "В телеграм канале: " + channelNameForParserFunc + " постов за указанный период нет."

			request := model_kafka.InitKafkaRequestForTGBot(messageFromTGB, textForUsers)

			jsonData, err01 := json.Marshal(request)
			if err01 != nil {
				myLogger.Error(fmt.Sprintf("Ошибка сериализации JSON: %v", err01))
			}

			err01 = producer.SendMessage(ctx, cfg.KafkaTopic_03, jsonData)
			if err01 != nil {
				myLogger.Error(fmt.Sprintf("Не удалось отправить сообщение в Kafka(tgapp-tgbot): %v", err01))
			}
			myLogger.Info("Сообщение в Kafka в TG-Bot успешно отправлено")

		} else {
			request := model_kafka.InitKafkaRequestForTTS(messageFromTGB, textForAudio)

			jsonData, err01 := json.Marshal(request)
			if err01 != nil {
				myLogger.Error(fmt.Sprintf("Ошибка сериализации JSON: %v", err01))
			}

			err01 = producer.SendMessage(ctx, cfg.KafkaTopic_02, jsonData)
			if err01 != nil {
				myLogger.Error(fmt.Sprintf("Не удалось отправить сообщение в Kafka(tgapp-texttospeech): %v", err01))
			}
			myLogger.Info("Сообщение в Kafka в TTS успешно отправлено")

			strChatID := strconv.FormatInt(c.MetricsPrometheus.ChatID, 10)
			c.MetricsPrometheus.ObserveDuration(c.MetricsPrometheus.RequestID, strChatID, c.MetricsPrometheus.TimeRequestCalculation())
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

// Создаем новый клиент для каждой итерации
// tgClient := tg_init_parser.NewClient(cfg, sessionStorage, pathSessionJson)
// myLogger.Info("Успешно создали tg клиента")

//serviceParser := service_parser.NewServiceParser(tgClient)

// myLogger.Info("Завершение Telegram-клиента")
// if err02 := tgClient.Close(ctx); err != nil {
//	myLogger.Error("Ошибка при закрытии Telegram-клиента", "error", err02)
// }
