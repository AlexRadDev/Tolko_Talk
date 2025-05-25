package interface_requests_storage

import (
	"tg_bot/internal/model/bot_request"
)

// UserRequestRepository интерфейс для хранилища состояния запросов пользователя
type UserRequestRepository interface {
	GetRequest(chatID int64) *bot_request.TgBotRequest
	SaveRequest(chatID int64, request *bot_request.TgBotRequest)
}
