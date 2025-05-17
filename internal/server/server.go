package server

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"

	//"log"
	"log/slog"
	"net/http"
	"tolko_talk/internal/app_telega"
	"tolko_talk/internal/app_text_to_speech"
	"tolko_talk/tools/logger"

	"tolko_talk/internal/config"
	"tolko_talk/internal/model/tg_bot_model"
)

// HandleSubmit принимает конфиг и экземпляр бота Telegram API,
// обрабатывает запрос /tgBotPost
func HandleSubmit(cfg *config.Config, bot *tgbotapi.BotAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const lbl = "internal/server/server.go/HandleSubmit()"
		logger := logger.NewColorLogger(lbl)
		slog.SetDefault(logger)

		if r.Method != http.MethodPost {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		var request tg_bot_model.TgBotRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
			return
		}

		slog.Info(fmt.Sprintf("Запрос от TG Bota: %+v\n", request))

		if request.ChatID == 0 {
			slog.Error("ChatID пустой или равен 0")
			http.Error(w, "ChatID не указан в запросе", http.StatusBadRequest)
			return
		}

		channelName := ""
		if len(request.NameChanel) > 0 && request.NameChanel[0] == '@' {
			channelName = request.NameChanel[1:]
		}
		const prefix = "https://t.me/"
		if len(request.NameChanel) > 0 && strings.HasPrefix(request.NameChanel, prefix) {
			channelName = strings.TrimPrefix(request.NameChanel, prefix)
		}
		slog.Info("Имя канала успешно извлечено", "channel_name", channelName)

		// Сразу возвращаем статус 202 Accepted, чтобы бот мог отправить подтверждение
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Запрос принят и обрабатывается, пожалуйста подождите немного."))

		// Обрабатываем запрос асинхронно
		go func(cfg *config.Config, bot *tgbotapi.BotAPI, channelName string) {
			// Парсинг канала
			textNews, err := app_telega.RunTelegaApp(cfg.TG_Api_ID, cfg.TG_Api_API_Hash, channelName, cfg.MyHpone_for_App, cfg.Two_Factor_Auth, request.TimePeriod)
			if err != nil {
				slog.Error(err.Error())
				msg := tgbotapi.NewMessage(request.ChatID, "Приносим извинения за неудобства! На удаленном сервере произошла ошибка, пожалуйста, повторите попытку позже.")
				bot.Send(msg)
				return
			}
			if textNews == "" || textNews == " " {
				slog.Error("Текст новостей пуст")
				msg := tgbotapi.NewMessage(request.ChatID, "За это время в канале нету новостей")
				bot.Send(msg)
				return
			}
			slog.Info(fmt.Sprintf("Текст из канала: %+v\n", textNews))

			// Синтез аудио
			outputFile := "sound_001.mp3"
			audioData, err := app_text_to_speech.SynthesizeText(textNews, cfg.Google_Key_To_Speech, outputFile, request.SpeakingRate)
			if err != nil {
				slog.Error(fmt.Sprintf("Ошибка функции SynthesizeText: %v", err))
				msg := tgbotapi.NewMessage(request.ChatID, "Приносим извинения за неудобства! На удаленном сервере произошла ошибка, пожалуйста, повторите попытку позже.")
				bot.Send(msg)
				return
			}
			slog.Info("Аудио синтезировано")

			// Отправка аудио
			audioBytes := tgbotapi.FileBytes{
				Name:  "audio.mp3",
				Bytes: audioData,
			}
			audioConfig := tgbotapi.NewAudio(request.ChatID, audioBytes)
			audioConfig.Title = "News Audio"
			audioConfig.Duration = 0
			audioConfig.Performer = "Bot"
			if _, err := bot.Send(audioConfig); err != nil {
				slog.Error(fmt.Sprintf("Ошибка отправки аудио: %v", err))
				msg := tgbotapi.NewMessage(request.ChatID, "Приносим извинения за неудобства! На удаленном сервере произошла ошибка, пожалуйста, повторите попытку позже.")
				bot.Send(msg)
				return
			}
			slog.Info("Аудио отправлено в чат")
		}(cfg, bot, channelName)
	}
}

func StartServer(cfg *config.Config, bot *tgbotapi.BotAPI) {
	const lbl = "internal/server/server.go/StartServer()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	http.HandleFunc("/tgBotPost", HandleSubmit(cfg, bot))

	slog.Info(fmt.Sprintf("Сервер запущен на порту: %v", cfg.ServerPort))
	if err := http.ListenAndServe(cfg.ServerPort, nil); err != nil {
		slog.Error(fmt.Sprintf("Ошибка запуска сервера: %v", err))
	}
}
