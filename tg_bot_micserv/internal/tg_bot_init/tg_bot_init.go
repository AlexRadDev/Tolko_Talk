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

// SetupLongPolling настраивает и возвращает канал обновлений для Long Polling
func SetupLongPolling(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	const lblSetupLP = "tg_bot_micserv/internal/tg_bot_init/tg_bot_init.go/SetupLongPolling()"
	myLogger := logger.NewColorLogger(lblSetupLP)

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60 // Тайм-аут в секундах
	// Создается конфигурацию запроса для получения обновлений:
	// 0 означает, что бот хочет все новые обновления.
	// Timeout = 60 — держит соединение открытым до 60 секунд, ожидая обновление от Telegram.

	// Начинаем получать обновления
	updates := bot.GetUpdatesChan(u)
	myLogger.Info("Long Polling успешно настроен")

	return updates
}
