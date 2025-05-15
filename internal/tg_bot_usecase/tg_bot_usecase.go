package tg_bot_usecase

import (
	"fmt"
	"log/slog"

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
	if _, ok := uc.userRequests[chatID]; !ok {
		uc.userRequests[chatID] = &tg_bot_model.TgBotRequest{}
	}
	slog.Info("Обрабатываем сообщение", "text", text, "chatID", chatID)

	if text == "/start" {
		slog.Info("Обработка команды /start", "chatID", chatID)
		removeMsg := tgbotapi.NewMessage(chatID, "Обновляем интерфейс...")
		removeMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		if _, err := bot.Send(removeMsg); err != nil {
			return err
		}
		slog.Info("Старая клавиатура удалена", "chatID", chatID)

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
		msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Выберите действие:")
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		return err
	}

	switch text {
	case "Выбрать канал":
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("@intercargon5"),
				tgbotapi.NewKeyboardButton("@technewsdaily"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "Выберите канал:")
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		return err
	case "@intercargon5", "@technewsdaily":
		uc.userRequests[chatID].NameChanel = text
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "Канал сохранён: "+text)
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		return err
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
		msg := tgbotapi.NewMessage(chatID, "Выберите скорость:")
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		return err
	case "0.5", "0.75", "1.0", "1.2", "1.5", "2.0":
		rate, _ := strconv.ParseFloat(text, 64)
		uc.userRequests[chatID].SpeakingRate = rate
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "Скорость сохранена: "+text)
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		return err
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
		msg := tgbotapi.NewMessage(chatID, "Выберите период:")
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		return err
	case "1", "2", "3", "4", "5", "6":
		period, _ := strconv.Atoi(text)
		uc.userRequests[chatID].TimePeriod = period
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "Период сохранён: "+text)
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		return err
	case "Назад":
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Выбрать канал"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Скорость голоса"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Период времени"),
				tgbotapi.NewKeyboardButton("Отправить"),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "Вернулись в главное меню. Выберите действие:")
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		return err
	case "Отправить":
		request := uc.userRequests[chatID]
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Отправлен запрос: Канал: %s, Скорость: %.1f, Период: %d", request.NameChanel, request.SpeakingRate, request.TimePeriod))
		_, err := bot.Send(msg)
		if err != nil {
			return err
		}
		// Здесь можно добавить отправку на бэкэнд, например, через HTTP-запрос
		slog.Info("Запрос отправлен", "request", request)
		return nil
	default:
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Выбрать канал"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Скорость голоса"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Период времени"),
				tgbotapi.NewKeyboardButton("Отправить"),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, выберите действие:")
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		return err
	}
}
