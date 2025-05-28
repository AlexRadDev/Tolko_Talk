package tg_bot_router

import (
	"context"
	"encoding/json"
	"fmt"
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
	ctx := req.Context()

	const lblHandleUpdate = "[myMicro] tg_bot_micserv/internal/tg_bot_router/HandleUpdate()"
	myLogger := logger.NewColorLogger(lblHandleUpdate)

	var update tgbotapi.Update
	if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	myLogger.Info("Успешно декодировали JSON из тела запроса")

	if update.Message != nil {
		err := r.tgBotUserCase.HandleMessage(ctx, r.tgBot, update.Message.Chat.ID, update.Message.Text)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка обработки запроса в HandleMessage: %v", err))
		}
		myLogger.Info("Успешно передали запрос в бизнес-логику")
	}
	w.WriteHeader(http.StatusOK)
}

// ProcessUpdate обрабатывает обновление из Long Polling
func (r *Router) ProcessUpdate(update tgbotapi.Update) {
	ctx := context.Background()

	const lblProcessUpdate = "[myMicro] tg_bot_micserv/internal/tg_bot_router/ProcessUpdate()"
	myLogger := logger.NewColorLogger(lblProcessUpdate)

	if update.Message != nil {
		err := r.tgBotUserCase.HandleMessage(ctx, r.tgBot, update.Message.Chat.ID, update.Message.Text)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка обработки запроса в HandleMessage: %v", err))
		}
		myLogger.Info("Успешно обработали сообщение")
		return
	}

	// Обработка нажатий на кнопки CallbackQuery
	if update.CallbackQuery != nil {

		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := r.tgBot.Request(callback); err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка при ответе на callback: %v", err))
		}

		err := r.tgBotUserCase.HandleCallback(r.tgBot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
		if err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка обработки действий в HandleCallback: %v", err))
		}
		myLogger.Info("Успешно обработали callback")

		return
	}
}
