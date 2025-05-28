package tg_bot_user_case

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"tg_bot/internal/kafka/producer"
	"tg_bot/internal/metrics/metrics_prometheus"
	"tg_bot/internal/metrics/metrics_time"
	"tg_bot/internal/model/model_kafka"
	"tg_bot/internal/storage/interface_channel_list_storage"
	"tg_bot/internal/storage/interface_requests_storage"
	"tg_bot/tools/logger"
)

// Константы для клавиатур
var (
	MainKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Выбрать канал из списка"),
			tgbotapi.NewKeyboardButton("Выбрать скорость речи"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Выбрать период времени"),
			tgbotapi.NewKeyboardButton("Отправить"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Добавить канал в избранное"),
			tgbotapi.NewKeyboardButton("Удалить канал из избранного"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Вставить канал по ссылке"),
		),
	)

	BackKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)

	SpeechRateKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("0.5"),
			tgbotapi.NewKeyboardButton("0.75"),
			tgbotapi.NewKeyboardButton("1.0"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("1.2"),
			tgbotapi.NewKeyboardButton("1.5"),
			tgbotapi.NewKeyboardButton("2.0"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)

	TimePeriodKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("1"),
			tgbotapi.NewKeyboardButton("2"),
			tgbotapi.NewKeyboardButton("3"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("4"),
			tgbotapi.NewKeyboardButton("5"),
			tgbotapi.NewKeyboardButton("6"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
)

func init() {
	MainKeyboard.ResizeKeyboard = true
	BackKeyboard.ResizeKeyboard = true
	SpeechRateKeyboard.ResizeKeyboard = true
	TimePeriodKeyboard.ResizeKeyboard = true
}

type UseCase struct {
	mylogger          *logger.Logger
	repo              interface_requests_storage.UserRequestRepository
	kafkaProducer     *producer.Producer
	storage           interface_channel_list_storage.StorageInterface
	MetricsTime       *metrics_time.MetricsTime
	MetricsPrometheus *metrics_prometheus.TGBotMetricsPrometheus
}

const lblname = "[myMicro] tg_bot/internal/tg_bot_user_case" // Метка места для логирования

func NewUseCase(repo interface_requests_storage.UserRequestRepository,
	kafkaProducer *producer.Producer,
	storage interface_channel_list_storage.StorageInterface,
	metricsTime *metrics_time.MetricsTime,
	metricsPrometheus *metrics_prometheus.TGBotMetricsPrometheus) *UseCase {
	return &UseCase{
		mylogger:          logger.NewColorLogger(lblname),
		repo:              repo,
		kafkaProducer:     kafkaProducer,
		storage:           storage,
		MetricsTime:       metricsTime,
		MetricsPrometheus: metricsPrometheus,
	}
}

// HandleMessage обрабатывает входящее сообщение от пользователя
func (uc *UseCase) HandleMessage(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, text string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("контекст отменён: %w", ctx.Err())
	}

	request := uc.repo.GetRequest(chatID)
	switch text {
	case "/start":
		return uc.handleStart(bot, chatID)
	case "Добавить канал в избранное", "Вставить канал по ссылке":
		return uc.handleAddChannel(bot, chatID, request)
	case "Выбрать канал из списка":
		return uc.handleSelectChannel(bot, chatID)
	case "Удалить канал из избранного":
		return uc.handleDeleteChannel(bot, chatID)
	case "Выбрать скорость речи":
		return uc.handleSelectSpeechRate(bot, chatID)
	case "0.5", "0.75", "1.0", "1.2", "1.5", "2.0":
		return uc.handleSpeechRateInput(bot, chatID, text, request)
	case "Выбрать период времени":
		return uc.handleSelectTimePeriod(bot, chatID)
	case "1", "2", "3", "4", "5", "6":
		return uc.handleTimePeriodInput(bot, chatID, text, request)
	case "Назад":
		return uc.handleBack(bot, chatID, request)
	case "Отправить":
		return uc.handleSendRequest(ctx, bot, chatID, request)
	default:
		return uc.handleDefault(ctx, bot, chatID, text, request)
	}
}

// handleStart обрабатывает команду /start
func (uc *UseCase) handleStart(bot *tgbotapi.BotAPI, chatID int64) error {
	uc.mylogger.Info(fmt.Sprintf("Обработка команды /start, chatID = %v", chatID))
	msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Выберите действие:")
	msg.ReplyMarkup = MainKeyboard
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения в handleStart: %w", err)
	}
	return nil
}

// handleAddChannel запрашивает ввод канала
func (uc *UseCase) handleAddChannel(bot *tgbotapi.BotAPI, chatID int64, request *model_kafka.KafkaMessageFromTGAPP) error {
	request.AwaitingChannelInput = true
	msg := tgbotapi.NewMessage(chatID, "Введите имя и ссылку на канал в формате: Название|Ссылка (Пример: MyChannel|https://t.me/mychannel)")
	msg.ReplyMarkup = BackKeyboard
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}

	uc.repo.SaveRequest(chatID, request)

	return nil
}

// handleSelectChannel отображает список каналов
func (uc *UseCase) handleSelectChannel(bot *tgbotapi.BotAPI, chatID int64) error {
	channels := uc.storage.GetChannels(chatID)
	if len(channels) == 0 {
		msg := tgbotapi.NewMessage(chatID, "У вас нет добавленных каналов. Сначала добавьте канал в избранное.")
		msg.ReplyMarkup = MainKeyboard
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf("ошибка отправки сообщения: %w", err)
		}
		return nil
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for name, link := range channels {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(name, fmt.Sprintf("channel_%s|%s", name, link)),
		))
	}
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_main"),
	))

	msg := tgbotapi.NewMessage(chatID, "Выберите канал из списка:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}
	return nil
}

// handleDeleteChannel отображает список каналов для удаления
func (uc *UseCase) handleDeleteChannel(bot *tgbotapi.BotAPI, chatID int64) error {
	channels := uc.storage.GetChannels(chatID)
	if len(channels) == 0 {
		msg := tgbotapi.NewMessage(chatID, "У вас нет добавленных каналов. Сначала добавьте канал в избранное.")
		msg.ReplyMarkup = MainKeyboard
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf("ошибка отправки сообщения: %w", err)
		}
		return nil
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for name := range channels {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(name, fmt.Sprintf("delete_channel_%s", name)),
		))
	}
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_main"),
	))

	msg := tgbotapi.NewMessage(chatID, "Выберите канал для удаления:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}
	return nil
}

// handleSelectSpeechRate отображает выбор скорости речи
func (uc *UseCase) handleSelectSpeechRate(bot *tgbotapi.BotAPI, chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Выберите скорость речи:")
	msg.ReplyMarkup = SpeechRateKeyboard
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}
	return nil
}

// handleSpeechRateInput сохраняет выбранную скорость речи
func (uc *UseCase) handleSpeechRateInput(bot *tgbotapi.BotAPI, chatID int64, text string, request *model_kafka.KafkaMessageFromTGAPP) error {
	rate, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return fmt.Errorf("Ошибка конвертации скорости речи из string в float64 %q: %w", text, err)
	}
	request.SpeakingRate = rate
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Скорость сохранена: %s.\nВыберите действие:", text))
	msg.ReplyMarkup = MainKeyboard
	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("Ошибка отправки сообщения: %w", err)
	}

	uc.repo.SaveRequest(chatID, request)

	return nil
}

// handleSelectTimePeriod отображает выбор периода времени
func (uc *UseCase) handleSelectTimePeriod(bot *tgbotapi.BotAPI, chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Выберите период времени:")
	msg.ReplyMarkup = TimePeriodKeyboard
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}
	return nil
}

// handleTimePeriodInput сохраняет выбранный период времени
func (uc *UseCase) handleTimePeriodInput(bot *tgbotapi.BotAPI, chatID int64, text string, request *model_kafka.KafkaMessageFromTGAPP) error {
	period, err := strconv.Atoi(text)
	if err != nil {
		return fmt.Errorf("ошибка парсинга периода времени %q: %w", text, err)
	}
	request.TimePeriod = period
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Период времени сохранён: %s ч.\nВыберите действие:", text))
	msg.ReplyMarkup = MainKeyboard
	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}

	uc.repo.SaveRequest(chatID, request)

	return nil
}

// handleBack возвращает в главное меню
func (uc *UseCase) handleBack(bot *tgbotapi.BotAPI, chatID int64, request *model_kafka.KafkaMessageFromTGAPP) error {
	msg := tgbotapi.NewMessage(chatID, "Вернулись в главное меню.\nВыберите действие:")
	msg.ReplyMarkup = MainKeyboard
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}

	uc.repo.SaveRequest(chatID, request)

	return nil
}

// handleSendRequest отправляет запрос в Kafka
func (uc *UseCase) handleSendRequest(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, request *model_kafka.KafkaMessageFromTGAPP) error {
	request.ChatID = chatID
	if request.NameChanel == "" {
		msg := tgbotapi.NewMessage(chatID, "Ошибка: Канал не выбран.\nПожалуйста, выберите канал перед отправкой.")
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf("ошибка отправки сообщения: %w", err)
		}
		return nil
	}
	if request.TimePeriod == 0 {
		request.TimePeriod = 1
	}
	if request.SpeakingRate == 0 {
		request.SpeakingRate = 1.2
	}

	// Фиксируем значения для метрик в прометэус
	uc.MetricsPrometheus.RequestID++
	uc.MetricsPrometheus.ChatID = chatID
	uc.MetricsPrometheus.TimeStartRequest = time.Now()

	uc.MetricsTime.TimeRequest = time.Now()
	uc.MetricsTime.RequestID++
	uc.MetricsTime.ChatID = chatID

	request.RequestID = strconv.FormatInt(uc.MetricsPrometheus.RequestID, 10)
	uc.MetricsPrometheus.IncUpdate() // считает количество запросов

	jsonData, err := json.Marshal(request)
	if err != nil {
		uc.mylogger.Error("Ошибка сериализации JSON", "chatID", chatID, "error", err)
		return fmt.Errorf("ошибка сериализации JSON: %w", err)
	}

	if err := uc.kafkaProducer.SendMessage(ctx, uc.kafkaProducer.Writer.Topic, jsonData); err != nil {
		uc.mylogger.Error("Ошибка отправки сообщения в Kafka", "chatID", chatID, "error", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка отправки запроса.\nПожалуйста, попробуйте позже.")
		bot.Send(msg)
		return fmt.Errorf("ошибка отправки в Kafka: %w", err)
	}

	uc.mylogger.Info("Запрос успешно отправлен в Kafka", "chatID", chatID)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
		"Ваш запрос отправлен на обработку:\nКанал: %s\nСкорость аудио: %.1fx\nПериод времени: %d ч.\nПожалуйста, дождитесь ответа...",
		request.NameChanel, request.SpeakingRate, request.TimePeriod,
	))
	msg.ReplyMarkup = MainKeyboard
	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}

	// Очистка состояния
	request.TimePeriod = 0
	request.SpeakingRate = 0
	request.NameChanel = ""

	uc.repo.SaveRequest(chatID, request)

	return nil
}

// handleDefault обрабатывает неизвестные команды
func (uc *UseCase) handleDefault(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, text string, request *model_kafka.KafkaMessageFromTGAPP) error {
	if !request.AwaitingChannelInput {
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите действие:")
		msg.ReplyMarkup = MainKeyboard
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf("ошибка отправки сообщения: %w", err)
		}

		uc.repo.SaveRequest(chatID, request)

		return nil
	}

	parts := strings.SplitN(text, "|", 2)
	if len(parts) != 2 ||
		strings.TrimSpace(parts[0]) == "" ||
		strings.TrimSpace(parts[0]) == " " ||
		strings.TrimSpace(parts[1]) == "" ||
		strings.TrimSpace(parts[1]) == " " ||
		!strings.HasPrefix(parts[1], "https://t.me/") {
		msg := tgbotapi.NewMessage(chatID, "Ошибка: Введите имя и ссылку в формате: Название|Ссылка (например: MyChannel|https://t.me/mychannel)")
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf("ошибка отправки сообщения: %w", err)
		}
		return nil
	}

	channelName, channelLink := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	if err := uc.storage.AddChannel(chatID, channelName, channelLink); err != nil {
		uc.mylogger.Error("Ошибка сохранения канала", "chatID", chatID, "channelName", channelName, "error", err)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка на сервере, канал %s не сохранён.\nВыберите действие:", channelName))
		msg.ReplyMarkup = MainKeyboard
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf("ошибка отправки сообщения: %w", err)
		}
		return nil
	}

	request.AwaitingChannelInput = false
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Канал сохранён в избранное: %s\nВыберите действие:", channelName))
	msg.ReplyMarkup = MainKeyboard
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}

	uc.repo.SaveRequest(chatID, request)

	return nil
}

// HandleCallback обрабатывает callback-запросы
func (uc *UseCase) HandleCallback(bot *tgbotapi.BotAPI, chatID int64, callbackData string) error {
	uc.mylogger.Info("Получен callback", "chatID", chatID, "callbackData", callbackData)
	request := uc.repo.GetRequest(chatID)

	switch {
	case callbackData == "back_to_main":
		msg := tgbotapi.NewMessage(chatID, "Вернулись в главное меню.\nВыберите действие:")
		msg.ReplyMarkup = MainKeyboard
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf("ошибка отправки сообщения: %w", err)
		}
	case strings.HasPrefix(callbackData, "channel_"):
		parts := strings.SplitN(callbackData[8:], "|", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			uc.mylogger.Error("Неверный формат callbackData", "callbackData", callbackData)
			msg := tgbotapi.NewMessage(chatID, "Ошибка: Неверный формат данных канала.\nПожалуйста, проверьте ввод.")
			_, err := bot.Send(msg)
			if err != nil {
				return fmt.Errorf("ошибка отправки сообщения: %w", err)
			}
			return nil
		}
		channelName, channelLink := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		request.NameChanel = channelLink
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Канал выбран: %s\nВыберите действие:", channelName))
		msg.ReplyMarkup = MainKeyboard
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf("ошибка отправки сообщения: %w", err)
		}
	case strings.HasPrefix(callbackData, "delete_channel_"):
		channelName := strings.TrimSpace(callbackData[15:])
		
		err := uc.storage.DeleteChannels(chatID, channelName)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка на сервере.\nКанал %s не удалён из избранного.\nПопробуйте позже.\nВыберите действие:", channelName))
			msg.ReplyMarkup = MainKeyboard
			_, err := bot.Send(msg)
			if err != nil {
				return fmt.Errorf("ошибка отправки сообщения: %w", err)
			}
		} else {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Канал %s удалён из избранного.\nВыберите действие:", channelName))
			msg.ReplyMarkup = MainKeyboard
			_, err := bot.Send(msg)
			if err != nil {
				return fmt.Errorf("ошибка отправки сообщения: %w", err)
			}
		}

	default:
		uc.mylogger.Warn("Получен неизвестный callback", "callbackData", callbackData)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Получен неизвестный callback: %s", callbackData))
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf("ошибка отправки сообщения: %w", err)
		}
	}

	uc.repo.SaveRequest(chatID, request)

	return nil
}
