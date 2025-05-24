package storage

import "sync"

type MapStorage struct {
	MapStorage map[int64]map[string]string
	Mutex      sync.RWMutex
}

// NewMapStorage инициализация объекта MapStorage
func NewMapStorage() *MapStorage {
	myMapStorage := make(map[int64]map[string]string, 0)
	return &MapStorage{myMapStorage, sync.RWMutex{}}
}

// Метод для добавления канала пользователю
func (s *MapStorage) AddChannel(chatID int64, nameChannel, linkChannel string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if _, exists := s.MapStorage[chatID]; !exists {
		s.MapStorage[chatID] = make(map[string]string)
	}
	s.MapStorage[chatID][nameChannel] = linkChannel
}

// Метод для получения списка каналов пользователя
func (s *MapStorage) GetChannels(chatID int64) map[string]string {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	channels, exists := s.MapStorage[chatID]
	if !exists {
		return map[string]string{}
	}
	return channels
}
