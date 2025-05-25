package map_requests_storage

import (
	"sync"
	"tg_bot/internal/model/bot_request"
)

// RepoUserRequests хранилище состояния запросов пользователя
type RepoUserRequests struct {
	userRequests map[int64]*bot_request.TgBotRequest
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
	request, exists := r.userRequests[chatID]
	r.mutex.RUnlock()

	if exists {
		return request
	}

	request = &bot_request.TgBotRequest{}
	r.mutex.Lock()
	r.userRequests[chatID] = request
	r.mutex.Unlock()

	return request
}

// SaveRequest сохраняет запрос пользователя
func (r *RepoUserRequests) SaveRequest(chatID int64, request *bot_request.TgBotRequest) {
	r.mutex.Lock()
	r.userRequests[chatID] = request
	r.mutex.Unlock()
}
