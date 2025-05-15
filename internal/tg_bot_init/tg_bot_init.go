package tg_bot_init

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"tolko_talk/tools/logger"
)

// Функция NewBot создаёт нового Telegram-бота с использованием переданного токена
func NewBot(token string) (*tgbotapi.BotAPI, error) {
	const lbl = "internal/tg_bot_init/tg_bot_init.go/NewBot()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	// Создаём нового бота с помощью токена
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	slog.Info(fmt.Sprintf("Бот успешно авторизован как: @%s", bot.Self.UserName))

	// Включаем режим отладки для бота (в продакшене рекомендуется отключить)
	bot.Debug = false // отключить на проде

	// Возвращаем указатель на созданного бота и nil в качестве ошибки
	return bot, nil
}
