package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	tg_post_model "tg_app_micserv/internal/model"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"

	"tg_app_micserv/internal/config"
	"tg_app_micserv/internal/kafka/producer"
	"tg_app_micserv/internal/metrics/metrics_prometheus"
	"tg_app_micserv/internal/model/model_kafka"
	"tg_app_micserv/internal/service_parser"
	"tg_app_micserv/internal/tg_app_init"
	"tg_app_micserv/tools/logger"
)

// Consumer обрабатывает сообщения из Kafka с использованием воркер-пула и Circuit Breaker.
type Consumer struct {
	reader         *kafka.Reader                              `json:"reader,omitempty"`
	metrics        *metrics_prometheus.TGBotMetricsPrometheus `json:"metrics,omitempty"`
	tgClient       *tg_app_init.TGAppClient                   `json:"tg_client,omitempty"`
	myLogger       *slog.Logger                               `json:"myLogger,omitempty"`
	bufferPool     *sync.Pool                                 `json:"buffer_pool,omitempty"`     // Пул для оптимизации аллокаций буферов
	workerPoolSize chan struct{}                              `json:"worker_pool,omitempty"`     // Ограничение параллельных обработок
	CircuitBreaker *gobreaker.CircuitBreaker                  `json:"circuit_breaker,omitempty"` // Circuit Breaker для устойчивости
}

// NewConsumer создаёт новый экземпляр Consumer с воркер-пулом и Circuit Breaker.
func NewConsumer(cfg *config.Config, metrics *metrics_prometheus.TGBotMetricsPrometheus, tgClient *tg_app_init.TGAppClient) *Consumer {
	const lblConsumer = "tg_app_micserv/internal/kafka/consumer"
	mylogger := logger.NewColorLogger(lblConsumer)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.KafkaPort,
		Topic:       cfg.KafkaTopic_01,
		GroupID:     "group_01",
		StartOffset: kafka.FirstOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     100 * time.Millisecond,
	})

	// Настройка Circuit Breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "tg-parser-circuit-breaker",
		MaxRequests: 2,                // Количество запросов в полуоткрытом состоянии
		Interval:    30 * time.Second, // Интервал сброса статистики
		Timeout:     10 * time.Second, // Время ожидания в открытом состоянии
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 3 // Срабатывание после 3 ошибок
		},
	})

	mylogger.Info("Kafka Consumer создан для топика %s с брокерами %v", cfg.KafkaTopic_01, cfg.KafkaPort)
	sizePool, err := strconv.Atoi(cfg.WorkerPoolSize)
	if err != nil {
		mylogger.Error("Ошибка перевода строки в инт", "err", err)
	}

	return &Consumer{
		reader:         reader,
		metrics:        metrics,
		tgClient:       tgClient,
		myLogger:       mylogger,
		bufferPool:     &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
		workerPoolSize: make(chan struct{}, sizePool),
		CircuitBreaker: cb,
	}
}

func (c *Consumer) Consume(ctx context.Context, cfg *config.Config, producer *producer.Producer) error {
	var wg sync.WaitGroup
	workerCount, err := strconv.Atoi(cfg.WorkerPoolSize) // Количество воркеров из конфигурации
	if err != nil {
		c.myLogger.Error("Ошибка перевода строки в инт", "err", err)
	}

	if workerCount <= 0 {
		workerCount = 1
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			c.myLogger.Debug(fmt.Sprintf("Запущен воркер %v", workerID))
			err = c.StartWorker(ctx, cfg, producer)
			if err != nil {
				c.myLogger.Error(fmt.Sprintf("Ошибка: %v", err))
			}
		}(i + 1)
	}

	wg.Wait()
	return nil
}

func (c *Consumer) StartWorker(ctx context.Context, cfg *config.Config, producer *producer.Producer) error {
	for {
		select {
		case <-ctx.Done():
			c.myLogger.Info("Получен сигнал завершения, остановка Consumer")
			return ctx.Err()
		default:
			if err := c.processMessage(ctx, cfg, producer); err != nil {
				c.myLogger.Error(fmt.Sprintf("Ошибка обработки сообщения: %v", err))
				continue
			}
		}
	}
}

// processMessage читает и обрабатывает одно сообщение из Kafka.
func (c *Consumer) processMessage(ctx context.Context, cfg *config.Config, producer *producer.Producer) error {
	c.workerPoolSize <- struct{}{} // Захватываем слот воркера
	defer func() { <-c.workerPoolSize }()

	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return fmt.Errorf("чтение сообщения: %w", err)
	}
	c.myLogger.Debug("Получено сообщение из топика %s", msg.Topic)

	var tgBotMessage *model_kafka.KafkaMessageFromTGB
	if err := json.Unmarshal(msg.Value, &tgBotMessage); err != nil {
		return fmt.Errorf("десериализация сообщения: %w", err)
	}

	// Валидация входных данных
	if tgBotMessage.RequestID == "" || tgBotMessage.ChatID == 0 || tgBotMessage.NameChanel == "" {
		return fmt.Errorf("некорректные входные данные: %+v", tgBotMessage)
	}

	// Записываем метрики
	c.metrics.RequestID = tgBotMessage.RequestID
	c.metrics.ChatID = tgBotMessage.ChatID
	c.metrics.TimeStartRequest = time.Now()

	// Форматируем имя канала
	channelName := service_parser.ExtractChannelName(tgBotMessage.NameChanel)
	timePeriod := time.Duration(tgBotMessage.TimePeriod) * time.Hour

	// Парсим посты из Telegram с использованием Circuit Breaker
	posts, err := c.CircuitBreaker.Execute(func() (interface{}, error) {
		return c.tgClient.PostParser(ctx, channelName, timePeriod)
	})
	if err != nil {
		return fmt.Errorf("извлечение постов из канала %s: %w", channelName, err)
	}

	// Форматируем текст для TTS
	formattedText, err := service_parser.FormationTextForTTS(posts.([]tg_post_model.Message))
	if err != nil {
		return fmt.Errorf("форматирование текста: %w", err)
	}

	// Подготавливаем и отправляем запрос
	buf := c.bufferPool.Get().(*bytes.Buffer)
	defer c.bufferPool.Put(buf)
	buf.Reset()

	var targetTopic string
	var request interface{}

	if len(formattedText) == 0 {
		targetTopic = cfg.KafkaTopic_03
		request = model_kafka.InitKafkaRequestForTGBot(tgBotMessage, "В телеграм канале "+channelName+" постов за указанный период нет.")
	} else {
		targetTopic = cfg.KafkaTopic_02
		request = model_kafka.InitKafkaRequestForTTS(tgBotMessage, formattedText)
	}

	if err := json.NewEncoder(buf).Encode(request); err != nil {
		return fmt.Errorf("сериализация JSON: %w", err)
	}

	// Отправка с ретраями
	err = c.sendWithRetry(ctx, producer, targetTopic, buf.Bytes())
	if err != nil {
		return fmt.Errorf("отправка сообщения в топик %s: %w", targetTopic, err)
	}

	c.myLogger.Debug(fmt.Sprintf("Сообщение успешно отправлено в топик %s", targetTopic))
	if len(formattedText) > 0 {
		c.metrics.ObserveDuration(c.metrics.RequestID, fmt.Sprintf("%d", c.metrics.ChatID), c.metrics.TimeRequestCalculation())
	}

	return nil
}

// sendWithRetry отправляет сообщение в Kafka с ретраями.
func (c *Consumer) sendWithRetry(ctx context.Context, producer *producer.Producer, topic string, data []byte) error {
	b := backoff.NewExponentialBackOff()

	// TODO: Вынести в конфиг значения и сбалансировать
	b.MaxElapsedTime = 5 * time.Second     // Максимальное время ретраев
	b.MaxInterval = 500 * time.Millisecond // Максимальный интервал между попытками

	return backoff.Retry(func() error {
		return producer.SendMessage(ctx, topic, data)
	}, backoff.WithContext(b, ctx))
}

// Close закрывает Kafka Reader.
func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("закрытие Reader: %w", err)
	}
	c.myLogger.Info("Kafka Consumer успешно закрыт")
	return nil
}

//type Consumer struct {
//	Reader            *kafka.Reader
//	MetricsPrometheus *metrics_prometheus.TGBotMetricsPrometheus
//	AppTGPostParser   *tg_app_init.TGAppClient
//}
//
//func NewConsumer(cfg *config.Config,
//	metricsPrometheus *metrics_prometheus.TGBotMetricsPrometheus,
//	appTGPostParser *tg_app_init.TGAppClient) *Consumer {
//
//	const lblNewConsumer = "tg_app_micserv/internal/kafka/consumer/NewConsumer()"
//	mylogger := myLogger.NewColorLogger(lblNewConsumer)
//
//	reader := kafka.NewReader(kafka.ReaderConfig{
//		Brokers:     cfg.KafkaPort,          // Список брокеров
//		Topic:       cfg.KafkaTopic_01,      // Топик для чтения
//		GroupID:     "group_01",             // ID группы консьюмеров
//		StartOffset: kafka.FirstOffset,      // читать с начала
//		MinBytes:    1,                      // Минимальный размер батча (10KB)
//		MaxBytes:    10e6,                   // Максимальный размер батча (10MB)
//		MaxWait:     100 * time.Millisecond, // ждать не более 100мс
//	})
//
//	mylogger.Info(fmt.Sprintf("Kafka Consumer создан для топика %s с брокерами %v", cfg.KafkaTopic_01, cfg.KafkaPort))
//	return &Consumer{
//		Reader:            reader,
//		MetricsPrometheus: metricsPrometheus,
//		AppTGPostParser:   appTGPostParser,
//	}
//}
//
//func (c *Consumer) Consume(ctx context.Context,
//	cfg *config.Config,
//	producer *producer.Producer) error {
//
//	const lblConsume = "[myMicro] tg_app_micserv/internal/kafka/consumer/Consume()"
//	myLogger := myLogger.NewColorLogger(lblConsume)
//	myLogger.Info("Подписались на топик Kafaka: \"tgbot-tgapp\"")
//
//	for {
//		msg, err := c.Reader.ReadMessage(ctx)
//		if err != nil {
//			myLogger.Error(fmt.Sprintf("Ошибка чтения сообщения из Kafka: %v", err))
//			return fmt.Errorf("чтение сообщения: %w", err)
//		}
//		myLogger.Info(fmt.Sprintf("Получено сообщение из топика %s: %s", msg.Topic, string(msg.Value)))
//
//		var messageFromTGB *model_kafka.KafkaMessageFromTGB // Структура для Unmarshal сообщения от ТГ-бота из Kafka
//		if err := json.Unmarshal(msg.Value, &messageFromTGB); err != nil {
//			myLogger.Error(fmt.Sprintf("Ошибка десериализации сообщения: %v", err))
//			continue
//		}
//
//		// Фиксируем значения для метрик в Prometheus
//		c.MetricsPrometheus.RequestID = messageFromTGB.RequestID
//		c.MetricsPrometheus.ChatID = messageFromTGB.ChatID
//		c.MetricsPrometheus.TimeStartRequest = time.Now()
//
//		// Форматируем имя канала, превращаем ссылку в нужный формат для TG API PARSER
//		channelNameForParserFunc := service_parser.ExtractChannelName(messageFromTGB.NameChanel)
//
//		timePeriod := time.Duration(messageFromTGB.TimePeriod) * time.Hour
//
//		// Парсим посты из ТГ канала
//		tgPosts, err := c.AppTGPostParser.PostParser(ctx, channelNameForParserFunc, timePeriod)
//		if err != nil {
//			myLogger.Error(fmt.Sprintf("Ошибка извлечения постов из канала: %v", err))
//			return err
//		}
//
//		// Форматируем посты в одно сообщение для Google TTS
//		textForAudio, err := service_parser.FormationTextForTTS(tgPosts)
//		if err != nil {
//			myLogger.Error(fmt.Sprintf("Ошибка форматирования текста постов: %v", err))
//			continue
//		}
//		myLogger.Info(fmt.Sprintf("Длинна мапы (textForAudio) равна: %v", len(textForAudio)))
//
//		if len(textForAudio) < 1 || textForAudio == nil { // Если новостей за этот период нет
//			textForUsers := "В телеграм канале: " + channelNameForParserFunc + " постов за указанный период нет."
//
//			request := model_kafka.InitKafkaRequestForTGBot(messageFromTGB, textForUsers)
//
//			jsonData, err01 := json.Marshal(request)
//			if err01 != nil {
//				myLogger.Error(fmt.Sprintf("Ошибка сериализации JSON: %v", err01))
//			}
//
//			err01 = producer.SendMessage(ctx, cfg.KafkaTopic_03, jsonData)
//			if err01 != nil {
//				myLogger.Error(fmt.Sprintf("Не удалось отправить сообщение в Kafka(tgapp-tgbot): %v", err01))
//			}
//			myLogger.Info("Сообщение в Kafka в TG-Bot успешно отправлено")
//
//		} else {
//			request := model_kafka.InitKafkaRequestForTTS(messageFromTGB, textForAudio)
//
//			jsonData, err01 := json.Marshal(request)
//			if err01 != nil {
//				myLogger.Error(fmt.Sprintf("Ошибка сериализации JSON: %v", err01))
//			}
//
//			err01 = producer.SendMessage(ctx, cfg.KafkaTopic_02, jsonData)
//			if err01 != nil {
//				myLogger.Error(fmt.Sprintf("Не удалось отправить сообщение в Kafka(tgapp-texttospeech): %v", err01))
//			}
//			myLogger.Info("Сообщение в Kafka в TTS успешно отправлено")
//
//			strChatID := strconv.FormatInt(c.MetricsPrometheus.ChatID, 10)
//			c.MetricsPrometheus.ObserveDuration(c.MetricsPrometheus.RequestID, strChatID, c.MetricsPrometheus.TimeRequestCalculation())
//		}
//	}
//}
//
//// Close закрывает Kafka Reader.
//func (c *Consumer) Close() error {
//	const lblClose = "[myMicro] tg_app_micserv/internal/kafka/consumer/Consume()"
//	mylogger := myLogger.NewColorLogger(lblClose)
//
//	if err := c.Reader.Close(); err != nil {
//		mylogger.Error(fmt.Sprintf("Ошибка закрытия Kafka Reader: %v", err))
//		return fmt.Errorf("закрытие Reader: %w", err)
//	}
//	mylogger.Info("Kafka Consumer успешно закрыт")
//	return nil
//}
