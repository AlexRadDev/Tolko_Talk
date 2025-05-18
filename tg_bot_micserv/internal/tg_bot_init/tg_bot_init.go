// Файл tg_bot_init.go отвечает за инициализацию Telegram-бота с использованием предоставленного токена.
// Настраивает бот с отключённым режимом отладки (для продакшена).

package tg_bot_init

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"tg_bot/tools/logger"
)

// NewBot создаёт новый Telegram-бот
func NewBot(token string) (*tgbotapi.BotAPI, error) {
	const lblNewBot = "tg_bot_micserv/internal/tg_bot_init/tg_bot_init.go/NewBot()"
	myLogger := logger.NewColorLogger(lblNewBot)

	// Создаём новый бот с токеном
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка создания бота: %v", err))
		return nil, err
	}

	bot.Debug = false // Отключаем режим отладки (для продакшена)

	return bot, nil
}
