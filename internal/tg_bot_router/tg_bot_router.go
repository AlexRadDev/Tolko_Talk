package tg_bot_router

import (
	"encoding/json"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tolko_talk/internal/tg_bot_handler"
)

// Определение структуры Router, которая содержит бота и обработчики
type Router struct {
	bot      *tgbotapi.BotAPI        // Указатель на экземпляр Telegram-бота
	handlers *tg_bot_handler.Handler // Указатель на структуру с обработчиками сообщений
}

// Конструктор NewRouter создаёт новый объект Router
func NewRouter(bot *tgbotapi.BotAPI, handlers *tg_bot_handler.Handler) *Router {
	return &Router{bot: bot, handlers: handlers}
}

// Метод HandleUpdate обрабатывает входящий HTTP-запрос с Telegram-обновлением
func (r *Router) HandleUpdate(w http.ResponseWriter, req *http.Request) {
	// Объявление переменной update для хранения данных от Telegram
	var update tgbotapi.Update

	// Декодирует JSON из тела запроса в структуру update, при ошибке возвращает 400 Bad Request
	if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Если обновление содержит сообщение (а не, например, callback), вызывается обработчик сообщений
	if update.Message != nil {
		r.handlers.HandleMessage(r.bot, update.Message) // Вызов обработчика сообщений, передаётся бот и само сообщение
	}
}
