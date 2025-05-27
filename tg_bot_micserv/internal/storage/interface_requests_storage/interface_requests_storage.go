package interface_requests_storage

import (
	"tg_bot/internal/model/model_kafka"
)

// UserRequestRepository интерфейс для хранилища состояния запросов пользователя
type UserRequestRepository interface {
	GetRequest(chatID int64) *model_kafka.KafkaMessageFromTGAPP
	SaveRequest(chatID int64, request *model_kafka.KafkaMessageFromTGAPP)
}
