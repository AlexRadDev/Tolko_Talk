// Файл bot_usecase.go обрабатывает входящие сообщения, взаимодействует с репозиторием.
// Отправляет ответы через Telegram Bot API.

package tg_bot_user_case

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	//"log/slog"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"tg_bot/internal/kafka/producer"
	"tg_bot/internal/repo_user_requests"
	"tg_bot/internal/storage"
	"tg_bot/tools/logger"
)

// MainKeyboard Определяем главную клавиатуру бота
var MainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Добавить канал в избранное"),
		tgbotapi.NewKeyboardButton("Выбрать канал из списка"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Вставить канал по ссылке"),
		tgbotapi.NewKeyboardButton("Скорость речи"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Период времени"),
		tgbotapi.NewKeyboardButton("Отправить"),
	),
)

// init Инициализируем настройки клавиатуры
func init() {
	MainKeyboard.ResizeKeyboard = true // Устанавливаем авто размер клавиатуры
	MainKeyboard.Selective = false     // Отключаем выборочную видимость клавиатуры
}

// Структура UseCase содержит бизнес-логику бота
type UseCase struct {
	repo          *repo_user_requests.RepoUserRequests // Интерфейс репозитория для работы с данными
	kafkaProducer *producer.Producer
	storage       storage.StorageInterface
}

// NewUseCase создаёт новый экземпляр UseCase
func NewUseCase(repo *repo_user_requests.RepoUserRequests, kafkaProducer *producer.Producer, storage storage.StorageInterface) *UseCase {
	return &UseCase{
		repo:          repo,
		kafkaProducer: kafkaProducer,
		storage:       storage,
	}
}

// HandleMessage обрабатывает входящее сообщение от пользователя
func (uc *UseCase) HandleMessage(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, text string) error {
	const lblHandleMessage = "tg_bot_micserv/internal/tg_bot_user_case/tg_bot_user_case.go/HandleMessage()"
	myLogger := logger.NewColorLogger(lblHandleMessage)

	// Получаем запрос пользователя из репозитория
	request := uc.repo.GetRequest(chatID)

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
	case "Добавить канал в избранное":
		request.AwaitingChannelInput = true // Устанавливаем флаг ожидания ввода канала
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)
		keyboard.ResizeKeyboard = true
		keyboard.Selective = false
		msg := tgbotapi.NewMessage(chatID, "Введите имя и ссылку на канал в формате: Название|Ссылка (например, MyChannel|https://t.me/mychannel)")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			myLogger.Error("Ошибка отправки сообщения", "error", err)
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil

	case "Выбрать канал из списка":
		channels := uc.storage.GetChannels(chatID)
		if len(channels) == 0 {
			msg := tgbotapi.NewMessage(chatID, "У вас нет добавленных каналов. Сначала добавьте канал в избранное.")
			msg.ReplyMarkup = MainKeyboard
			if _, err := bot.Send(msg); err != nil {
				myLogger.Error("Ошибка отправки сообщения", "error", err)
				return err
			}
			return nil
		}

		// Создаём инлайн-клавиатуру со списком каналов
		var buttons [][]tgbotapi.InlineKeyboardButton
		for name, link := range channels {
			button := tgbotapi.NewInlineKeyboardButtonData(name, fmt.Sprintf("channel_%s|%s", name, link))
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
		}
		// Добавляем кнопку "Назад"
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_main"),
		))

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
		msg := tgbotapi.NewMessage(chatID, "Выберите канал из списка:")
		msg.ReplyMarkup = inlineKeyboard
		if _, err := bot.Send(msg); err != nil {
			myLogger.Error("Ошибка отправки сообщения с инлайн-клавиатурой", "error", err)
			return err
		}
		return nil

	case "Вставить канал по ссылке":

		request.AwaitingChannelInput = true // Устанавливаем флаг ожидания ввода канала
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)

		keyboard.ResizeKeyboard = true // Устанавливаем авто размер клавиатуры
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
		msg := tgbotapi.NewMessage(chatID, "Выберите скорость речи:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil

	case "0.5", "0.75", "1.0", "1.2", "1.5", "2.0":
		rate, _ := strconv.ParseFloat(text, 64)
		request.SpeakingRate = rate
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Скорость сохранена: %s.\nВыберите действие:", text))
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
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
		msg := tgbotapi.NewMessage(chatID, "Выберите период времени:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil

	case "1", "2", "3", "4", "5", "6":
		period, _ := strconv.Atoi(text)
		request.TimePeriod = period
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Период времени сохранён: %s ч..\nВыберите действие:", text))
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil

	case "Назад":
		msg := tgbotapi.NewMessage(chatID, "Вернулись в главное меню.\nВыберите действие:")
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil

	case "Отправить":
		request.ChatID = chatID

		// Валидация запроса клиента
		if request.NameChanel == "" {
			msg := tgbotapi.NewMessage(chatID, "Ошибка: Канал не выбран.\nПожалуйста, введите ссылку на желаемый телеграмм канала перед отправкой.")
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
			msg := tgbotapi.NewMessage(chatID, "Ошибка отправки запроса.\nПожалуйста попробуйте позже.")
			bot.Send(msg)
			return err
		}
		myLogger.Info(fmt.Sprintf("Запрос под номером: %v, успешно отправлен в kafka", request.ChatID))

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ваш запрос отправлен на обработку: \nКанал: %s, \nСкорость аудио: %.1fx, \nПериод времени: %d час., \nПожалуйста дождитесь ответа...", request.NameChanel, request.SpeakingRate, request.TimePeriod))
		msg.ReplyMarkup = MainKeyboard
		bot.Send(msg)

		// Очистка состояния запроса
		request.TimePeriod = 0
		request.SpeakingRate = 0
		request.NameChanel = ""
		uc.repo.SaveRequest(chatID, request)
		return nil

	default:
		if request.AwaitingChannelInput {
			if len(text) > 0 {
				// Проверяем формат ввода: Название|Ссылка
				parts := strings.SplitN(text, "|", 2)
				if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
					msg := tgbotapi.NewMessage(chatID, "Ошибка: Введите имя и ссылку в формате: Название|Ссылка")
					if _, err := bot.Send(msg); err != nil {
						myLogger.Error("Ошибка отправки сообщения", "error", err)
						return err
					}
					return nil
				}
				nameChannel, linkChannel := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
				uc.storage.AddChannel(chatID, nameChannel, linkChannel)
				request.AwaitingChannelInput = false
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Канал сохранён в избранное: %s\nВыберите действие:", nameChannel))
				msg.ReplyMarkup = MainKeyboard
				if _, err := bot.Send(msg); err != nil {
					myLogger.Error("Ошибка отправки сообщения", "error", err)
					return err
				}
				uc.repo.SaveRequest(chatID, request)
				return nil
			}

			msg := tgbotapi.NewMessage(chatID, "Ошибка: Введите имя и ссылку в формате: Название|Ссылка")
			if _, err := bot.Send(msg); err != nil {
				myLogger.Error("Ошибка отправки сообщения", "error", err)
				return err
			}
			return nil
		}

		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите действие:")
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			myLogger.Error("Ошибка отправки сообщения", "error", err)
			return err
		}
		uc.repo.SaveRequest(chatID, request)
		return nil
	}
}

// HandleCallback обрабатывает данные callback-запроса (нажатие на InlineKeyboard кнопки)
func (uc *UseCase) HandleCallback(bot *tgbotapi.BotAPI, chatID int64, callbackData string) error {
	const lblHandleCallback = "tg_bot_micserv/internal/tg_bot_user_case/tg_bot_user_case.go/HandleCallback()"
	myLogger := logger.NewColorLogger(lblHandleCallback)

	// Получаем текущий запрос пользователя
	request := uc.repo.GetRequest(chatID)
	myLogger.Info("Получен callback", "chatID", chatID, "callbackData", callbackData)

	switch callbackData {
	case "back_to_main":
		// Возвращение в главное меню
		msg := tgbotapi.NewMessage(chatID, "Вернулись в главное меню.\nВыберите действие:")
		msg.ReplyMarkup = MainKeyboard
		if _, err := bot.Send(msg); err != nil {
			myLogger.Error("Ошибка отправки сообщения", "error", err)
			return err
		}

	default:
		if strings.HasPrefix(callbackData, "channel_") {
			// Извлекаем имя и ссылку канала из callbackData
			parts := strings.SplitN(callbackData[8:], "|", 2)
			if len(parts) != 2 {
				myLogger.Error("Неверный формат callbackData", "callbackData", callbackData)
				msg := tgbotapi.NewMessage(chatID, "Ошибка: Неверный формат данных канала.")
				if _, err := bot.Send(msg); err != nil {
					return err
				}
				return nil
			}
			nameChannel, linkChannel := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			//request.NameChanel = nameChannel
			request.NameChanel = linkChannel
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Канал выбран: %s\nВыберите действие:", nameChannel))
			msg.ReplyMarkup = MainKeyboard
			if _, err := bot.Send(msg); err != nil {
				myLogger.Error("Ошибка отправки сообщения", "error", err)
				return err
			}
		} else {
			// Для неизвестных callback отправляем текст callback'а
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Получен неизвестный callback: %s", callbackData))
			if _, err := bot.Send(msg); err != nil {
				myLogger.Error("Ошибка отправки сообщения", "error", err)
				return err
			}
		}
	}

	// Сохраняем обновлённый запрос
	uc.repo.SaveRequest(chatID, request)
	return nil
}
