package tg_bot_usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"tolko_talk/tools/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"tolko_talk/internal/model/tg_bot_model"
)

// UseCase хранит состояние пользователей
type UseCase struct {
	userRequests map[int64]*tg_bot_model.TgBotRequest
}

// Конструктор для UseCase — создаёт объект с пустым словарём userRequests
func NewUseCase() *UseCase {
	return &UseCase{
		userRequests: make(map[int64]*tg_bot_model.TgBotRequest),
	}
}

// HandleMessage Основной метод обработки входящего сообщения от пользователя
// принимает Telegram API клиент, ID чата и текст сообщения
// возвращает ошибку, если что-то пошло не так при отправке ответа
func (uc *UseCase) HandleMessage(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	const lbl = "internal/tg_bot_usecase/tg_bot_usecase.go/HandleMessage()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	if _, ok := uc.userRequests[chatID]; !ok {
		uc.userRequests[chatID] = &tg_bot_model.TgBotRequest{}
	}
	slog.Info("Обрабатываем сообщение", "text", text, "chatID", chatID)

	if text == "/start" {
		slog.Info("Обработка команды /start", "chatID", chatID)
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Выбрать канал"),
				tgbotapi.NewKeyboardButton("Скорость голоса"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Период времени"),
				tgbotapi.NewKeyboardButton("Отправить"),
			),
		)
		keyboard.ResizeKeyboard = true
		keyboard.Selective = false
		msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Выберите действие:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		slog.Info("Старая клавиатура удалена и отображена новая", "chatID", chatID)
		return nil
	}

	//--------------------------------------------------------------------------------------------------------------
	switch text {
	case "Выбрать канал":
		uc.userRequests[chatID].AwaitingChannelInput = true // Устанавливаем состояние ожидания ввода канала
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)
		keyboard.ResizeKeyboard = true
		keyboard.Selective = false
		msg := tgbotapi.NewMessage(chatID, "Введите имя ТГ канала в формате @name")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		return nil

	//--------------------------------------------------------------------------------------------------------------
	case "Скорость голоса":
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
		msg := tgbotapi.NewMessage(chatID, "Выберите скорость:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		return nil
	case "0.5", "0.75", "1.0", "1.2", "1.5", "2.0":
		rate, _ := strconv.ParseFloat(text, 64)
		uc.userRequests[chatID].SpeakingRate = rate
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)
		keyboard.ResizeKeyboard = true
		keyboard.Selective = false
		msg := tgbotapi.NewMessage(chatID, "Скорость сохранена: "+text)
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		return nil

		//--------------------------------------------------------------------------------------------------------------
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
		return nil
	case "1", "2", "3", "4", "5", "6":
		var period time.Duration
		period01, _ := strconv.Atoi(text)
		period = time.Duration(period01)
		uc.userRequests[chatID].TimePeriod = period
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)
		keyboard.ResizeKeyboard = true
		keyboard.Selective = false
		msg := tgbotapi.NewMessage(chatID, "Период сохранён: "+text)
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		return nil

		//--------------------------------------------------------------------------------------------------------------
	case "Назад":
		slog.Info("Возврат в главное меню", "chatID", chatID)
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Выбрать канал"),
				tgbotapi.NewKeyboardButton("Скорость речи"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Период времени"),
				tgbotapi.NewKeyboardButton("Отправить"),
			),
		)
		keyboard.ResizeKeyboard = true
		keyboard.Selective = false
		msg := tgbotapi.NewMessage(chatID, "Вернулись в главное меню. Выберите действие:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		return nil

		//--------------------------------------------------------------------------------------------------------------
	case "Отправить":
		request := uc.userRequests[chatID]

		// Валидация
		if request.NameChanel == "" {
			msg := tgbotapi.NewMessage(chatID, "Ошибка: Не выбран канал. Пожалуйста, выберите канал перед отправкой.")
			if _, err := bot.Send(msg); err != nil {
				return err
			}
			return nil
		}
		if request.TimePeriod == 0 {
			request.TimePeriod = 1 // Значение по умолчанию: 1 час
		}
		if request.SpeakingRate == 0 {
			request.SpeakingRate = 1.0 // Значение по умолчанию: 1.0
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			slog.Error("Ошибка сериализации JSON", "error", err)
			return err
		}

		resp, err := http.Post("http://localhost:4000/tgBotPost", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			slog.Error("Ошибка отправки запроса на бэкэнд", "error", err)
			msg := tgbotapi.NewMessage(chatID, "Ошибка при отправке запроса на сервер.")
			bot.Send(msg)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			slog.Warn("Бэкэнд вернул неуспешный статус", "status", resp.Status)
			msg := tgbotapi.NewMessage(chatID, "Бэкэнд вернул ошибку.")
			bot.Send(msg)
			return fmt.Errorf("неуспешный статус от бэкэнда: %d", resp.StatusCode)
		}

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Запрос успешно отправлен: Канал: %s, Скорость: %.1f, Период: %d", request.NameChanel, request.SpeakingRate, request.TimePeriod))
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		slog.Info("Запрос отправлен на бэкэнд", "request", request)

		// Очистка значений после успешной отправки
		uc.userRequests[chatID].TimePeriod = 0
		uc.userRequests[chatID].SpeakingRate = 0
		uc.userRequests[chatID].NameChanel = ""
		slog.Info("Значения очищены", "chatID", chatID)

		return nil

	default:
		// Обработка ввода имени канала, если ожидается
		if uc.userRequests[chatID].AwaitingChannelInput {
			if len(text) > 0 && text[0] == '@' {
				if len(text) == 1 {
					msg := tgbotapi.NewMessage(chatID, "Ошибка: Введите имя канала после @.")
					if _, err := bot.Send(msg); err != nil {
						return err
					}
					return nil
				}
				uc.userRequests[chatID].NameChanel = text
				uc.userRequests[chatID].AwaitingChannelInput = false // Сбрасываем состояние после успешного ввода
				keyboard := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Назад"),
					),
				)
				keyboard.ResizeKeyboard = true
				keyboard.Selective = false
				msg := tgbotapi.NewMessage(chatID, "Канал сохранён: "+text)
				msg.ReplyMarkup = keyboard
				if _, err := bot.Send(msg); err != nil {
					return err
				}
				return nil
			} else {
				msg := tgbotapi.NewMessage(chatID, "Ошибка: Имя канала должно начинаться с @. Введите заново.")
				if _, err := bot.Send(msg); err != nil {
					return err
				}
				return nil
			}
		}

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Выбрать канал"),
				tgbotapi.NewKeyboardButton("Скорость голоса"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Период времени"),
				tgbotapi.NewKeyboardButton("Отправить"),
			),
		)
		keyboard.ResizeKeyboard = true
		keyboard.Selective = false
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите действие:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		return nil
	}
}
