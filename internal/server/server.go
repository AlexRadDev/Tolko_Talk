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
				msg := tgbotapi.NewMessage(request.ChatID, "Ошибка при парсинге канала: "+err.Error())
				bot.Send(msg)
				return
			}
			if textNews == "" || textNews == " " {
				slog.Error("Текст новостей пуст")
				msg := tgbotapi.NewMessage(request.ChatID, "Текст новостей пуст")
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

//func HandleSubmit(cfg *config.Config, bot *tgbotapi.BotAPI) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		const lbl = "internal/server/server.go/HandleSubmit()"
//		logger := logger.NewColorLogger(lbl)
//		slog.SetDefault(logger)
//
//		// Проверяем, что метод запроса является POST, иначе возвращаем ошибку 405
//		if r.Method != http.MethodPost {
//			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
//			return
//		}
//
//		// Декодируем JSON из тела запроса в структуру request, при ошибке возвращаем 400
//		var request tg_bot_model.TgBotRequest // var request TgBotRequest
//		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
//			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
//			return
//		}
//		slog.Info(fmt.Sprintf("Тело запроса от TG Bota: %+v\n", request))
//
//		// Убираем @ из имени канала
//		if len(request.NameChanel) > 0 && request.NameChanel[0] == '@' {
//			request.NameChanel = request.NameChanel[1:]
//		}
//
//		// Запускаем парсинг канала телеги
//		textNews, err := app_telega.RunTelegaApp(cfg.TG_Api_ID, cfg.TG_Api_API_Hash, request.NameChanel, cfg.MyHpone_for_App, cfg.Two_Factor_Auth, request.TimePeriod)
//		if err != nil {
//			slog.Error(err.Error())
//			log.Fatalf("Ошибка функции RunTelegaApp: %v", err)
//		}
//		if textNews == "" || textNews == " " {
//			slog.Error("Текст новостей пуст")
//			return
//		}
//		slog.Info(fmt.Sprintf("Текст из канала: %+v\n", textNews))
//
//		// Синтезируем речь в MP3
//		mp3Paht := "sound_001.mp3"
//		audioData, err := app_text_to_speech.SynthesizeText(textNews, cfg.Google_Key_To_Speech, mp3Paht, request.SpeakingRate)
//		if err != nil {
//			slog.Error(fmt.Sprintf("Ошибка функции SynthesizeText: %v", err))
//			return
//			//log.Fatalf("Ошибка функции SynthesizeText: %v", err)
//		}
//		slog.Info("Аудио записано")
//
//		// Создаем объект для отправки аудио
//		audioBytes := tgbotapi.FileBytes{
//			Name:  "audio.mp3",
//			Bytes: audioData,
//		}
//
//		// Создаем конфигурацию для отправки аудио
//		audioConfig := tgbotapi.NewAudio(request.ChatID, audioBytes)
//		audioConfig.Title = "News Audio" // Название аудио (опционально)
//		audioConfig.Duration = 0         // Длительность в секундах (можно оставить 0, Telegram определит сам)
//		audioConfig.Performer = "Bot"    // Исполнитель (опционально)
//
//		// Отправляем аудио в чат
//		_, err = bot.Send(audioConfig)
//		if err != nil {
//			slog.Error(fmt.Sprintf("Ошибка отправки аудио: %v", err))
//			http.Error(w, fmt.Sprintf("Ошибка отправки аудио: %v", err), http.StatusInternalServerError)
//			return
//		}
//
//		// Ответ бэкэнда
//		w.WriteHeader(http.StatusOK)
//		w.Write([]byte("Аудио отправлено в чат"))
//	}
//}

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
