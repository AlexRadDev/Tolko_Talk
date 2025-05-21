// Файл bot_usecase.go реализует бизнес-логику Telegram-бота. Содержит слой
// UseCase, который обрабатывает входящие сообщения, взаимодействует с репозиторием
// и отправляет ответы через Telegram Bot API. Соответствует принципам чистой
// архитектуры, разделяя бизнес-логику от инфраструктуры и доставки.

package tg_bot_user_case

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"tg_bot/internal/kafka/producer"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tg_bot/internal/repo_user_requests"
	"tg_bot/tools/logger"
)

// MainKeyboard Определяем главную клавиатуру бота
var MainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Выбрать канал"),
		tgbotapi.NewKeyboardButton("Скорость речи"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Период времени"),
		tgbotapi.NewKeyboardButton("Отправить"),
	),
)

// init Инициализируем настройки клавиатуры
func init() {
	MainKeyboard.ResizeKeyboard = true // Устанавливаем авторазмер клавиатуры
	MainKeyboard.Selective = false     // Отключаем выборочную видимость клавиатуры
}

// Структура UseCase содержит бизнес-логику бота
type UseCase struct {
	repo          *repo_user_requests.RepoUserRequests // repo — интерфейс репозитория для работы с данными
	kafkaProducer *producer.Producer                   // Указатель на Kafka-продюсер для отправки сообщений
}

// NewUseCase создаёт новый экземпляр UseCase
func NewUseCase(repo *repo_user_requests.RepoUserRequests, kafkaProducer *producer.Producer) *UseCase {
	return &UseCase{
		repo:          repo,
		kafkaProducer: kafkaProducer,
	}
}

// HandleMessage обрабатывает входящее сообщение от пользователя
func (uc *UseCase) HandleMessage(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, text string) error {
	const lblHandleMessage = "tg_bot_micserv/internal/tg_bot_user_case/tg_bot_user_case.go/HandleMessage()"
	myLogger := logger.NewColorLogger(lblHandleMessage)
	//slog.SetDefault(myLogger)

	// Получаем запрос пользователя из репозитория
	request := uc.repo.GetRequest(chatID)

	// Обрабатываем команду /start
	if text == "/start" {
		myLogger.Info("Обработка команды /start", "chatID", chatID)
		msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Выберите действие:")
		// Устанавливаем главную клавиатуру
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		myLogger.Info("Отображена главная клавиатура", "chatID", chatID)
		uc.repo.SaveRequest(chatID, request) // Сохраняем обновлённый запрос
		return nil
	}

	switch text {
	case "Выбрать канал":

		request.AwaitingChannelInput = true    // Устанавливаем флаг ожидания ввода канала
		keyboard := tgbotapi.NewReplyKeyboard( // Создаём клавиатуру с кнопкой "Назад"
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)

		keyboard.ResizeKeyboard = true // Устанавливаем авторазмер клавиатуры
		keyboard.Selective = false     // Отключаем выборочную видимость
		msg := tgbotapi.NewMessage(chatID, "Введите имя ТГ канала, или ссылку на канал")
		msg.ReplyMarkup = keyboard               // Устанавливаем клавиатуру
		if _, err := bot.Send(msg); err != nil { // Отправляем сообщение
			return err
		}
		uc.repo.SaveRequest(chatID, request) // Сохраняем обновлённый запрос
		return nil

	case "Скорость речи":
		keyboard := tgbotapi.NewReplyKeyboard(
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
		keyboard.ResizeKeyboard = true
		keyboard.Selective = false
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Выберите скорость речи (время: %d)", time.Now().Unix()))
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil

	case "0.5", "0.75", "1.0", "1.2", "1.5", "2.0":
		rate, _ := strconv.ParseFloat(text, 64)
		request.SpeakingRate = rate
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Скорость сохранена: %s. Выберите действие:", text))
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		// Возвращаем nil, так как обработка успешна
		return nil

	case "Период времени":
		keyboard := tgbotapi.NewReplyKeyboard(
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
		keyboard.ResizeKeyboard = true
		keyboard.Selective = false
		msg := tgbotapi.NewMessage(chatID, "Выберите период:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil

	case "1", "2", "3", "4", "5", "6":
		period, _ := strconv.Atoi(text)
		request.TimePeriod = period
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Период сохранён: %s. Выберите действие:", text))
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil

	case "Назад":
		msg := tgbotapi.NewMessage(chatID, "Вернулись в главное меню. Выберите действие:")
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil

	case "Отправить":
		request.ChatID = chatID

		// Валидация
		if request.NameChanel == "" {
			msg := tgbotapi.NewMessage(chatID, "Ошибка: Не выбран канал. Пожалуйста, выберите канал перед отправкой.")
			bot.Send(msg)
			return nil
		}
		if request.TimePeriod == 0 {
			request.TimePeriod = 1
		}
		if request.SpeakingRate == 0 {
			request.SpeakingRate = 1.0
		}

		// Сериализация запроса в JSON
		jsonData, err := json.Marshal(request)
		if err != nil {
			myLogger.Error("Ошибка сериализации JSON", "error", err)
			return err
		}

		// Отправка в Kafka
		err = uc.kafkaProducer.SendMessage(ctx, uc.kafkaProducer.Writer.Topic, jsonData)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка отправки сообщения в Kafka: %v", err))
			msg := tgbotapi.NewMessage(chatID, "Ошибка отправки. Повторите позже.")
			bot.Send(msg)
			return err
		}
		myLogger.Info(fmt.Sprintf("Запрос под номнром: %v, успешно ушёл в kafka", request.ChatID))

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Запрос отправлен в обработку. Канал: %s, Скорость: %.1fx, Период: %d час.", request.NameChanel, request.SpeakingRate, request.TimePeriod))
		msg.ReplyMarkup = MainKeyboard
		bot.Send(msg)

		// Очистка состояния
		request.TimePeriod = 0
		request.SpeakingRate = 0
		request.NameChanel = ""
		uc.repo.SaveRequest(chatID, request)
		return nil

	default:
		if request.AwaitingChannelInput {
			if len(text) > 0 {
				//if text[0] != '@' {
				//	msg := tgbotapi.NewMessage(chatID, "Ошибка: Имя канала должно начинаться с @. Введите заново.")
				//	if _, err := bot.Send(msg); err != nil {
				//		return err
				//	}
				//	uc.repo.SaveRequest(chatID, request)
				//	return nil
				//}

				request.NameChanel = text
				request.AwaitingChannelInput = false
				msg := tgbotapi.NewMessage(chatID, "Канал сохранён: "+text+". Выберите действие:")
				msg.ReplyMarkup = MainKeyboard
				if _, err := bot.Send(msg); err != nil {
					return err
				}
				uc.repo.SaveRequest(chatID, request)
				return nil
			}

			msg := tgbotapi.NewMessage(chatID, "Ошибка: Введите имя канала")
			if _, err := bot.Send(msg); err != nil {
				return err
			}
			uc.repo.SaveRequest(chatID, request)
			return nil
		}

		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите действие:")
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil
	}
}

// TODO: Разобраться
// HandleCallback обрабатывает данные callback-запроса (нажатие на InlineKeyboard кнопки)
func (uc *UseCase) HandleCallback(bot *tgbotapi.BotAPI, chatID int64, callbackData string) error {
	const lblHandleCallback = "tg_bot_micserv/internal/tg_bot_user_case/tg_bot_user_case.go/HandleCallback()"
	myLogger := logger.NewColorLogger(lblHandleCallback)
	slog.SetDefault(myLogger)

	// Получаем текущий запрос пользователя
	request := uc.repo.GetRequest(chatID)

	// Логируем получение callback
	myLogger.Info("Получен callback", "chatID", chatID, "callbackData", callbackData)

	// В зависимости от данных callback выполняем разные действия
	// Здесь вы можете добавить обработку для конкретных callbackData ваших кнопок
	// Например:
	switch callbackData {
	case "speed_0.5", "speed_0.75", "speed_1.0", "speed_1.2", "speed_1.5", "speed_2.0":
		// Парсим скорость из callback-данных
		speedStr := callbackData[6:] // получаем часть строки после "speed_"
		speed, err := strconv.ParseFloat(speedStr, 64)
		if err != nil {
			myLogger.Error("Ошибка парсинга скорости", "error", err)
			return err
		}

		// Сохраняем выбранную скорость
		request.SpeakingRate = speed

		// Отправляем сообщение о выбранной скорости
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Скорость сохранена: %s", speedStr))
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}

	case "period_1", "period_2", "period_3", "period_4", "period_5", "period_6":
		// Парсим период из callback-данных
		periodStr := callbackData[7:] // получаем часть строки после "period_"
		period, err := strconv.Atoi(periodStr)
		if err != nil {
			myLogger.Error("Ошибка парсинга периода", "error", err)
			return err
		}

		// Сохраняем выбранный период
		request.TimePeriod = period

		// Отправляем сообщение о выбранном периоде
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Период сохранён: %d", period))
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}

	default:
		// Для неизвестных callback просто отправляем текст callback'а
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Получен callback: %s", callbackData))
		if _, err := bot.Send(msg); err != nil {
			return err
		}
	}

	// Сохраняем обновлённый запрос
	uc.repo.SaveRequest(chatID, request)
	return nil
}
