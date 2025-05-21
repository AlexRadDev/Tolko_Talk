// repo_user_requests.go реализует репозиторий для хранения состояния запросов пользователей.
// Использует map для хранения данных в памяти.
// Реализация соответствует интерфейсу UserRequestRepository, обеспечивая независимость слоя бизнес-логики от хранения.

package repo_user_requests

import (
	"sync"
	"tg_bot/internal/model/bot_request"
)

// Структура RepoUserRequests реализует in-memory репозиторий
type RepoUserRequests struct {
	userRequests map[int64]*bot_request.TgBotRequest // карта для хранения запросов пользователей
	mutex        sync.RWMutex
}

// NewRepoUserRequests создаёт хранилище для запросов пользователей
func NewRepoUserRequests() *RepoUserRequests {
	return &RepoUserRequests{
		userRequests: make(map[int64]*bot_request.TgBotRequest),
		mutex:        sync.RWMutex{},
	}
}

// GetRequest возвращает запрос пользователя по chatID или создаёт новый
func (r *RepoUserRequests) GetRequest(chatID int64) *bot_request.TgBotRequest {
	r.mutex.RLock()
	request, exists := r.userRequests[chatID] // Проверяем, существует ли запрос для данного chatID
	r.mutex.RUnlock()

	// Если запрос существует, возвращаем его
	if exists {
		return request
	}

	request = &bot_request.TgBotRequest{} // Создаём новый запрос
	r.mutex.Lock()
	r.userRequests[chatID] = request // Сохраняем новый запрос в карту
	r.mutex.Unlock()

	// Возвращаем новый запрос
	return request
}

// SaveRequest сохраняет запрос пользователя
func (r *RepoUserRequests) SaveRequest(chatID int64, request *bot_request.TgBotRequest) {
	r.mutex.Lock()
	r.userRequests[chatID] = request
	r.mutex.Unlock()
}
