// Файл tg_bot_router.go реализует HTTP-роутер для обработки входящих запросов от bota.
// Отвечает за десериализацию JSON-данных, вызов бизнес-логики и отправку ответов.
// Соответствует слою доставки в чистой архитектуре, изолируя HTTP-обработку от бизнес-логики.

package tg_bot_router

import (
	"encoding/json"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"tg_bot/internal/tg_bot_user_case"
	"tg_bot/tools/logger"
)

// Структура Router содержит зависимости для обработки запросов
type Router struct {
	tgBot         *tgbotapi.BotAPI
	tgBotUserCase *tg_bot_user_case.UseCase
}

// NewRouter создаёт новый HTTP-роутер
func NewRouter(tgBot *tgbotapi.BotAPI, tgBotUserCase *tg_bot_user_case.UseCase) *Router {
	return &Router{tgBot: tgBot, tgBotUserCase: tgBotUserCase}
}

// HandleUpdate обрабатывает входящий HTTP-запрос с Telegram-обновлением
func (r *Router) HandleUpdate(w http.ResponseWriter, req *http.Request) {
	const lblHandleUpdate = "tg_bot_micserv/internal/tg_bot_router/tg_bot_router.go/HandleUpdate()"
	myLogger := logger.NewColorLogger(lblHandleUpdate)

	// Создаём структуру для хранения обновления
	var update tgbotapi.Update

	// Декодируем JSON из тела запроса
	if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	myLogger.Info("Успешно декодировали JSON из тела запроса")

	// Проверяем, содержит ли обновление сообщение
	if update.Message != nil {
		// Передаём сообщение в бизнес-логику
		r.tgBotUserCase.HandleMessage(r.tgBot, update.Message.Chat.ID, update.Message.Text)
		myLogger.Info("Успешно передали запрос в бизнес-логику")
	}

	// Возвращаем статус 200 OK
	w.WriteHeader(http.StatusOK)
}

// ProcessUpdate обрабатывает обновление из Long Polling
func (r *Router) ProcessUpdate(update tgbotapi.Update) {
	const lblProcessUpdate = "tg_bot_micserv/internal/tg_bot_router/tg_bot_router.go/ProcessUpdate()"
	myLogger := logger.NewColorLogger(lblProcessUpdate)

	// Обработка обычных сообщений
	if update.Message != nil {
		r.tgBotUserCase.HandleMessage(r.tgBot, update.Message.Chat.ID, update.Message.Text)
		myLogger.Info("Успешно обработали сообщение")
		return
	}

	// Обработка нажатий на кнопки (CallbackQuery)
	if update.CallbackQuery != nil {
		// Отвечаем на callback, чтобы убрать часы загрузки с кнопки
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := r.tgBot.Request(callback); err != nil {
			myLogger.Error("Ошибка при ответе на callback", "error", err)
		}

		// Обрабатываем данные callback
		r.tgBotUserCase.HandleCallback(r.tgBot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
		myLogger.Info("Успешно обработали callback")
		return
	}
}
