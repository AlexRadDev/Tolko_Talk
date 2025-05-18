// Файл bot_usecase.go реализует бизнес-логику Telegram-бота. Содержит слой
// UseCase, который обрабатывает входящие сообщения, взаимодействует с репозиторием
// и отправляет ответы через Telegram Bot API. Соответствует принципам чистой
// архитектуры, разделяя бизнес-логику от инфраструктуры и доставки.

package tg_bot_user_case

import (
	"encoding/json"
	"fmt"
	"strconv"
	"tg_bot/internal/kafka"
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
	kafkaProducer *kafka.Producer                      // Указатель на Kafka-продюсер для отправки сообщений
}

// NewUseCase создаёт новый экземпляр UseCase
func NewUseCase(repo *repo_user_requests.RepoUserRequests, kafkaProducer *kafka.Producer) *UseCase {
	return &UseCase{
		repo:          repo,
		kafkaProducer: kafkaProducer,
	}
}

// HandleMessage обрабатывает входящее сообщение от пользователя
func (uc *UseCase) HandleMessage(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	const lblHandleMessage = "tg_bot_micserv/internal/tg_bot_user_case/tg_bot_user_case.go/HandleMessage()"
	myLogger := logger.NewColorLogger(lblHandleMessage)

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
		// Устанавливаем chatID в запросе
		request.ChatID = chatID

		// Проверяем, указан ли канал
		if request.NameChanel == "" {
			// Создаём сообщение об ошибке
			msg := tgbotapi.NewMessage(chatID, "Ошибка: Не выбран канал. Пожалуйста, выберите канал перед отправкой.")
			// Отправляем сообщение
			if _, err := bot.Send(msg); err != nil {
				// Возвращаем ошибку отправки сообщения
				return err
			}
			// Сохраняем обновлённый запрос
			uc.repo.SaveRequest(chatID, request)
			// Возвращаем nil, так как обработка успешна
			return nil
		}
		// Устанавливаем значение по умолчанию для периода, если не указано
		if request.TimePeriod == 0 {
			request.TimePeriod = 1
		}
		// Устанавливаем значение по умолчанию для скорости, если не указано
		if request.SpeakingRate == 0 {
			request.SpeakingRate = 1.0
		}

		// Сериализуем запрос в JSON
		jsonData, err := json.Marshal(request)
		if err != nil {
			// Логируем ошибку сериализации
			myLogger.Error("Ошибка сериализации JSON", "error", err)
			// Возвращаем ошибку
			return err
		}

		// Отправляем сообщение в Kafka-продюсер
		err = uc.kafkaProducer.SendMessage("tg-bot-requests", jsonData)
		if err != nil {
			// Логируем ошибку отправки сообщения в Kafka
			myLogger.Error("Ошибка отправки сообщения в Kafka", "error", err)
			// Создаём сообщение об ошибке для пользователя
			msg := tgbotapi.NewMessage(chatID, "Приносим извинения за неудобства! Ошибка при отправке запроса. Повторите попытку позже.")
			// Отправляем сообщение
			bot.Send(msg)
			// Возвращаем ошибку
			return err
		}

		// Создаём сообщение об успешной отправке
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Запрос успешно отправлен: Канал: %s, Скорость: %.1fx, Период: %dчас., пожалуйста дождитесь ответа.", request.NameChanel, request.SpeakingRate, request.TimePeriod))
		// Отправляем сообщение
		if _, err := bot.Send(msg); err != nil {
			// Возвращаем ошибку отправки сообщения
			return err
		}
		// Логируем успешную отправку запроса
		myLogger.Info("Запрос отправлен в Kafka", "request", request)

		// Очищаем значения запроса
		request.TimePeriod = 0
		request.SpeakingRate = 0
		request.NameChanel = ""

		myLogger.Info("Значения очищены", "chatID", chatID)
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

//package tg_bot_user_case
//
//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"log/slog"
//	"net/http"
//	"strconv"
//	"time"
//
//	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
//
//	"tg_bot/internal/repo_user_requests"
//	"tg_bot/tools/logger"
//)
//
//const logLabel = "tg_bot_micserv/internal/tg_bot_user_case/tg_bot_user_case.go/HandleMessage()"
//
//// Определяем главную клавиатуру бота
//var MainKeyboard = tgbotapi.NewReplyKeyboard(
//	tgbotapi.NewKeyboardButtonRow(
//		tgbotapi.NewKeyboardButton("Выбрать канал"),
//		tgbotapi.NewKeyboardButton("Скорость речи"),
//	),
//	tgbotapi.NewKeyboardButtonRow(
//		tgbotapi.NewKeyboardButton("Период времени"),
//		tgbotapi.NewKeyboardButton("Отправить"),
//	),
//)
//
//// Инициализируем настройки клавиатуры
//func init() {
//	MainKeyboard.ResizeKeyboard = true // Устанавливаем авторазмер клавиатуры
//	MainKeyboard.Selective = false     // Отключаем выборочную видимость клавиатуры
//}
//
//// Структура UseCase содержит бизнес-логику бота
//type UseCase struct {
//	repo *repo_user_requests.RepoUserRequests // repo — интерфейс репозитория для работы с данными
//}
//
//// NewUseCase создаёт новый экземпляр UseCase
//func NewUseCase(repo *repo_user_requests.RepoUserRequests) *UseCase {
//	return &UseCase{repo: repo}
//}
//
//// HandleMessage обрабатывает входящее сообщение от пользователя
//func (uc *UseCase) HandleMessage(bot *tgbotapi.BotAPI, chatID int64, text string) error {
//	logger := logger.NewColorLogger(logLabel)
//	slog.SetDefault(logger)
//
//	// Получаем запрос пользователя из репозитория
//	request := uc.repo.GetRequest(chatID)
//
//	// Обрабатываем команду /start
//	if text == "/start" {
//		slog.Info("Обработка команды /start", "chatID", chatID)
//		msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Выберите действие:")
//		// Устанавливаем главную клавиатуру
//		msg.ReplyMarkup = MainKeyboard
//		if _, err := bot.Send(msg); err != nil {
//			return err
//		}
//		slog.Info("Отображена главная клавиатура", "chatID", chatID)
//		uc.repo.SaveRequest(chatID, request) // Сохраняем обновлённый запрос
//		return nil
//	}
//
//	switch text {
//	case "Выбрать канал":
//
//		request.AwaitingChannelInput = true    // Устанавливаем флаг ожидания ввода канала
//		keyboard := tgbotapi.NewReplyKeyboard( // Создаём клавиатуру с кнопкой "Назад"
//			tgbotapi.NewKeyboardButtonRow(
//				tgbotapi.NewKeyboardButton("Назад"),
//			),
//		)
//
//		keyboard.ResizeKeyboard = true // Устанавливаем авторазмер клавиатуры
//		keyboard.Selective = false     // Отключаем выборочную видимость
//		msg := tgbotapi.NewMessage(chatID, "Введите имя ТГ канала, или ссылку на канал")
//		msg.ReplyMarkup = keyboard               // Устанавливаем клавиатуру
//		if _, err := bot.Send(msg); err != nil { // Отправляем сообщение
//			return err
//		}
//		uc.repo.SaveRequest(chatID, request) // Сохраняем обновлённый запрос
//		return nil
//
//	case "Скорость речи":
//		keyboard := tgbotapi.NewReplyKeyboard(
//			tgbotapi.NewKeyboardButtonRow(
//				tgbotapi.NewKeyboardButton("0.5"),
//				tgbotapi.NewKeyboardButton("0.75"),
//				tgbotapi.NewKeyboardButton("1.0"),
//			),
//			tgbotapi.NewKeyboardButtonRow(
//				tgbotapi.NewKeyboardButton("1.2"),
//				tgbotapi.NewKeyboardButton("1.5"),
//				tgbotapi.NewKeyboardButton("2.0"),
//			),
//			tgbotapi.NewKeyboardButtonRow(
//				tgbotapi.NewKeyboardButton("Назад"),
//			),
//		)
//		keyboard.ResizeKeyboard = true
//		keyboard.Selective = false
//		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Выберите скорость речи (время: %d)", time.Now().Unix()))
//		msg.ReplyMarkup = keyboard
//		if _, err := bot.Send(msg); err != nil {
//			return err
//		}
//		uc.repo.SaveRequest(chatID, request)
//		return nil
//
//	case "0.5", "0.75", "1.0", "1.2", "1.5", "2.0":
//		rate, _ := strconv.ParseFloat(text, 64)
//		request.SpeakingRate = rate
//		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Скорость сохранена: %s. Выберите действие:", text))
//		msg.ReplyMarkup = MainKeyboard
//		if _, err := bot.Send(msg); err != nil {
//			return err
//		}
//		uc.repo.SaveRequest(chatID, request)
//		// Возвращаем nil, так как обработка успешна
//		return nil
//
//	case "Период времени":
//		keyboard := tgbotapi.NewReplyKeyboard(
//			tgbotapi.NewKeyboardButtonRow(
//				tgbotapi.NewKeyboardButton("1"),
//				tgbotapi.NewKeyboardButton("2"),
//				tgbotapi.NewKeyboardButton("3"),
//			),
//			tgbotapi.NewKeyboardButtonRow(
//				tgbotapi.NewKeyboardButton("4"),
//				tgbotapi.NewKeyboardButton("5"),
//				tgbotapi.NewKeyboardButton("6"),
//			),
//			tgbotapi.NewKeyboardButtonRow(
//				tgbotapi.NewKeyboardButton("Назад"),
//			),
//		)
//		keyboard.ResizeKeyboard = true
//		keyboard.Selective = false
//		msg := tgbotapi.NewMessage(chatID, "Выберите период:")
//		msg.ReplyMarkup = keyboard
//		if _, err := bot.Send(msg); err != nil {
//			return err
//		}
//		uc.repo.SaveRequest(chatID, request)
//		return nil
//
//	case "1", "2", "3", "4", "5", "6":
//		period, _ := strconv.Atoi(text)
//		request.TimePeriod = period
//		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Период сохранён: %s. Выберите действие:", text))
//		msg.ReplyMarkup = MainKeyboard
//		if _, err := bot.Send(msg); err != nil {
//			return err
//		}
//		uc.repo.SaveRequest(chatID, request)
//		return nil
//
//	case "Назад":
//		msg := tgbotapi.NewMessage(chatID, "Вернулись в главное меню. Выберите действие:")
//		msg.ReplyMarkup = MainKeyboard
//		if _, err := bot.Send(msg); err != nil {
//			return err
//		}
//		uc.repo.SaveRequest(chatID, request)
//		return nil
//
//	case "Отправить":
//		request.ChatID = chatID
//
//		// Валидация
//		if request.NameChanel == "" {
//			msg := tgbotapi.NewMessage(chatID, "Ошибка: Не выбран канал. Пожалуйста, выберите канал перед отправкой.")
//			if _, err := bot.Send(msg); err != nil {
//				return err
//			}
//			uc.repo.SaveRequest(chatID, request)
//			return nil
//		}
//		if request.TimePeriod == 0 {
//			request.TimePeriod = 1
//		}
//		if request.SpeakingRate == 0 {
//			request.SpeakingRate = 1.0
//		}
//
//		jsonData, err := json.Marshal(request) // Сериализуем запрос в JSON
//		if err != nil {
//			slog.Error("Ошибка сериализации JSON", "error", err)
//			return err
//		}
//
//		// Отправляем HTTP-запрос на сервер
//		resp, err := http.Post("http://localhost:4000/post_parser", "application/json", bytes.NewBuffer(jsonData))
//		if err != nil {
//			slog.Error("Ошибка отправки запроса на сервер", "error", err)
//			msg := tgbotapi.NewMessage(chatID, "Приносим извинения за неудобства! Ошибка при отправке запроса на сервер. Повторите попытку позже.")
//			bot.Send(msg)
//			return err
//		}
//		defer resp.Body.Close() // Закрываем тело ответа после использования
//
//		if resp.StatusCode != http.StatusAccepted { // Проверяем статус ответа
//			slog.Warn("Бэкэнд вернул неуспешный статус", "status", resp.Status)
//			msg := tgbotapi.NewMessage(chatID, "Приносим извинения за неудобства! На удаленном сервере произошла ошибка, пожалуйста, повторите попытку позже.")
//			bot.Send(msg)
//			return fmt.Errorf("неуспешный статус от бэкэнда: %d", resp.StatusCode)
//		}
//
//		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Запрос успешно отправлен: Канал: %s, Скорость: %.1fx, Период: %dчас., пожалуйста дождитесь ответа.", request.NameChanel, request.SpeakingRate, request.TimePeriod))
//		if _, err := bot.Send(msg); err != nil {
//			return err
//		}
//		slog.Info("Запрос отправлен на сервер", "request", request)
//
//		// Очищаем значения запроса
//		request.TimePeriod = 0
//		request.SpeakingRate = 0
//		request.NameChanel = ""
//
//		slog.Info("Значения очищены", "chatID", chatID)
//
//		uc.repo.SaveRequest(chatID, request)
//
//		return nil
//
//	default:
//		if request.AwaitingChannelInput {
//			if len(text) > 0 {
//				//if text[0] != '@' {
//				//	msg := tgbotapi.NewMessage(chatID, "Ошибка: Имя канала должно начинаться с @. Введите заново.")
//				//	if _, err := bot.Send(msg); err != nil {
//				//		return err
//				//	}
//				//	uc.repo.SaveRequest(chatID, request)
//				//	return nil
//				//}
//
//				request.NameChanel = text
//				request.AwaitingChannelInput = false
//				msg := tgbotapi.NewMessage(chatID, "Канал сохранён: "+text+". Выберите действие:")
//				msg.ReplyMarkup = MainKeyboard
//				if _, err := bot.Send(msg); err != nil {
//					return err
//				}
//				uc.repo.SaveRequest(chatID, request)
//				return nil
//			}
//
//			msg := tgbotapi.NewMessage(chatID, "Ошибка: Введите имя канала")
//			if _, err := bot.Send(msg); err != nil {
//				return err
//			}
//			uc.repo.SaveRequest(chatID, request)
//			return nil
//		}
//
//		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите действие:")
//		msg.ReplyMarkup = MainKeyboard
//		if _, err := bot.Send(msg); err != nil {
//			return err
//		}
//		uc.repo.SaveRequest(chatID, request)
//		return nil
//	}
//}
