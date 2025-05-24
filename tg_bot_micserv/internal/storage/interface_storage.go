package storage

type StorageInterface interface {
	AddChannel(chatID int64, nameChannel, linkChannel string)
	GetChannels(chatID int64) map[string]string
	// TODO: Сделать удаление канала из списка
}
