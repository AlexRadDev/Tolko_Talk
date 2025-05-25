package interface_channel_list_storage

// StorageInterface интерфейс для хранилища списка избранных каналов пользователя
type StorageInterface interface {
	AddChannel(chatID int64, nameChannel, linkChannel string) error
	GetChannels(chatID int64) map[string]string
	DeleteChannels(chatID int64, nameChannel string)
}
