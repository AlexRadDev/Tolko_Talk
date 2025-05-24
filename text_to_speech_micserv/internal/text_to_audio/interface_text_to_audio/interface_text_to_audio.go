package interface_text_to_audio

import (
	"context"

	"text_to_speech_app/internal/model/model_kafka"
)

// InterfaceTextToAudio интерфейс для превращения текста в аудио
type InterfaceTextToAudio interface {
	Synthesize(ctx context.Context, messageFromTGAPP *model_kafka.KafkaMessageFromTGAPP) (*model_kafka.KafkaRequestForTGAPP, error)
}
