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
	}

	// Возвращаем статус 200 OK
	w.WriteHeader(http.StatusOK)
}
