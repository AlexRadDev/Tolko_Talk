package tg_bot_init

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewBot(token string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = true // отключить в проде
	slog.Info(fmt.Sprintf("Бот успешно авторизован как: @%s", bot.Self.UserName))
	return bot, nil
}
