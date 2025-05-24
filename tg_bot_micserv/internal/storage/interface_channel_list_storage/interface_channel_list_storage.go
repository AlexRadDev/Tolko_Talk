package interface_channel_list_storage

type StorageInterface interface {
	AddChannel(chatID int64, nameChannel, linkChannel string)
	GetChannels(chatID int64) map[string]string
	DeleteChannels(chatID int64, nameChannel string)
}
