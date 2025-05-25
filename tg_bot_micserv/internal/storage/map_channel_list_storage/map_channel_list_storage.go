package map_channel_list_storage

import (
	"sync"
)

// MapStorage хранилище в map
type MapStorage struct {
	MapStorage map[int64]map[string]string
	Mutex      sync.RWMutex
}

// NewMapStorage инициализация объекта MapStorage
func NewMapStorage() *MapStorage {
	myMapStorage := make(map[int64]map[string]string, 0)
	return &MapStorage{myMapStorage, sync.RWMutex{}}
}

// AddChannel добавление канала в избранное
func (s *MapStorage) AddChannel(chatID int64, nameChannel, linkChannel string) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if _, exists := s.MapStorage[chatID]; !exists {
		s.MapStorage[chatID] = make(map[string]string)
	}
	s.MapStorage[chatID][nameChannel] = linkChannel

	return nil
}

// GetChannels получения списка избранных каналов
func (s *MapStorage) GetChannels(chatID int64) map[string]string {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	channels, exists := s.MapStorage[chatID]
	if !exists {
		return map[string]string{}
	}
	return channels
}

// DeleteChannels удаление канала из списка
func (s *MapStorage) DeleteChannels(chatID int64, nameChannel string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if daughterMap, ok := s.MapStorage[chatID]; ok { // daughterMap - дочерняя мапа по ключу chatID
		delete(daughterMap, nameChannel) // удаляет ключ(nameChannel) и значение из дочерней мапы daughterMap
		if len(daughterMap) == 0 {
			delete(s.MapStorage, chatID)
		}
	}
}
