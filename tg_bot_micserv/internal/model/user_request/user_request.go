// Файл user_request.go определяет интерфейс репозитория для работы с состоянием запросов пользователей.
// Интерфейс абстрагирует доступ к данным, и позволяет легко заменить реализацию репозитория (например, на базу данных).

package user_request

import (
	"tg_bot/internal/model/bot_request"
)

// Интерфейс UserRequestRepository определяет методы для работы с запросами пользователей
type UserRequestRepository interface {
	GetRequest(chatID int64) *bot_request.TgBotRequest           // Возвращает запрос пользователя по chatID, либо создаёт новый
	SaveRequest(chatID int64, request *bot_request.TgBotRequest) // Сохраняет запрос пользователя
}
