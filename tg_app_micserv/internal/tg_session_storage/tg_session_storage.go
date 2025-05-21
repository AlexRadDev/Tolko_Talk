// Хранение сеансов

package tg_session_storage

import "github.com/gotd/td/telegram"

// FileSessionStorage обертка для telegram.FileSessionStorage
type FileSessionStorage struct {
	*telegram.FileSessionStorage
}

// NewFileSessionStorage создает новый FileSessionStorage
func NewFileSessionStorage(path string) *FileSessionStorage {
	return &FileSessionStorage{
		FileSessionStorage: &telegram.FileSessionStorage{
			Path: path,
		},
	}
}
