package app_telega

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"tolko_talk/tools/logger"

	//"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// customCodeAuthenticator реализует auth.CodeAuthenticator для ввода номера телефона и кода
type customCodeAuthenticator struct {
	phone    string
	password string
}

func (c *customCodeAuthenticator) Phone(_ context.Context) (string, error) {
	return c.phone, nil
}

func (c *customCodeAuthenticator) Code(_ context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Println("Введите код подтверждения:")
	var code string
	fmt.Scanln(&code)
	return code, nil
}

func (c *customCodeAuthenticator) Password(_ context.Context) (string, error) {
	// Возвращаем пароль из переменной окружения или заданное значение
	password := os.Getenv("TWO_FACTOR_AUTH")
	if password == "" {
		slog.Info("Пароль TWO_FACTOR_AUTH пустой")
	}
	return password, nil
}

func (c *customCodeAuthenticator) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	fmt.Printf("Условия использования Telegram: %s\n", tos.Text)
	fmt.Println("Условия автоматически приняты.")
	return nil
}

func (c *customCodeAuthenticator) SignUp(_ context.Context) (auth.UserInfo, error) {
	fmt.Println("Требуется регистрация нового аккаунта.")
	fmt.Println("Введите ваше имя:")
	var firstName string
	fmt.Scanln(&firstName)

	fmt.Println("Введите вашу фамилию (или оставьте пустым):")
	var lastName string
	fmt.Scanln(&lastName)

	return auth.UserInfo{
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}

// ---------------------------------------------------------------------------------------------------------------------

// RunTelegaApp запускает клиент телеги и парсит нужный канал
func RunTelegaApp(apiID int, apiHash, channelName, phone, password string, timePeriod time.Duration) (string, error) {
	const lbl = "internal/app_telega/app_telega.go/RunTelegaApp()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	var resultMessage strings.Builder // Используем strings.Builder для сборки текста
	resultText := ""

	// Настройка хранилища сессии в файл
	sessionStorage := &telegram.FileSessionStorage{
		Path: "session.json", // Путь к файлу для хранения сессии
	}

	// Создаем клиент с хранилищем сессии
	client := telegram.NewClient(apiID, apiHash, telegram.Options{
		SessionStorage: sessionStorage,
	})
	slog.Info("Создали клиента Telega App")

	// Запускаем клиент
	err := client.Run(context.Background(), func(ctx context.Context) error {
		// Проверяем статус авторизации
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("не удалось проверить статус авторизации: %w", err)
		}

		// Если пользователь не авторизован, выполняем авторизацию
		if !status.Authorized {
			authenticator := &customCodeAuthenticator{
				phone:    phone,
				password: password,
			}
			flow := auth.NewFlow(authenticator, auth.SendCodeOptions{})

			// Запускаем процесс авторизации
			if err := flow.Run(ctx, client.Auth()); err != nil {
				return fmt.Errorf("ошибка авторизации: %w", err)
			}
			slog.Info("Успешно авторизовались")
		} else {
			slog.Info("Уже авторизованы, используем существующую сессию")
		}

		// Получаем объект канала
		api := client.API()
		usernameRequest := &tg.ContactsResolveUsernameRequest{
			Username: channelName,
		}
		resolved, err := api.ContactsResolveUsername(ctx, usernameRequest)
		if err != nil {
			return fmt.Errorf("не удалось найти канал: %w", err)
		}

		// Извлекаем InputPeer из результата
		var inputPeer tg.InputPeerClass
		if len(resolved.Chats) > 0 {
			chat, ok := resolved.Chats[0].(*tg.Channel)
			if !ok {
				return fmt.Errorf("ожидался тип Channel, получен %T", resolved.Chats[0])
			}
			inputPeer = &tg.InputPeerChannel{
				ChannelID:  chat.ID,
				AccessHash: chat.AccessHash,
			}
		} else if len(resolved.Users) > 0 {
			user, ok := resolved.Users[0].(*tg.User)
			if !ok {
				return fmt.Errorf("ожидался тип User, получен %T", resolved.Users[0])
			}
			inputPeer = &tg.InputPeerUser{
				UserID:     user.ID,
				AccessHash: user.AccessHash,
			}
		} else {
			return fmt.Errorf("канал или пользователь не найден")
		}

		// Определяем временной порог
		timeThreshold := time.Now().Add(-timePeriod * time.Hour)

		// Запрашиваем сообщения
		messages, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:  inputPeer,
			Limit: 100,
		})
		if err != nil {
			return fmt.Errorf("не удалось получить сообщения: %w", err)
		}
		slog.Info("Посты из тг канала успешно спаршены")

		// Обрабатываем сообщения
		var msgSlice []tg.MessageClass
		switch m := messages.(type) {
		case *tg.MessagesMessages:
			msgSlice = m.Messages
		case *tg.MessagesMessagesSlice:
			msgSlice = m.Messages
		case *tg.MessagesChannelMessages:
			msgSlice = m.Messages
		default:
			return fmt.Errorf("неожиданный тип сообщений: %T", m)
		}

		for _, msg := range msgSlice {
			message, ok := msg.(*tg.Message)
			if !ok {
				continue
			}

			// Преобразуем время сообщения
			msgTime := time.Unix(int64(message.Date), 0)
			if msgTime.Before(timeThreshold) {
				break
			}

			//messageText := fmt.Sprintf("ID: %d, Дата: %s, Текст: %s\n", message.ID, msgTime, cleanMessage)
			//messageText := fmt.Sprintf("%s, %s\n", msgTime.Format("15:04"), cleanMessage)
			messageText := fmt.Sprintf("%s\n", message.Message)

			// Добавляем сообщение в resultMessage
			resultMessage.WriteString(messageText)
		}

		// Форматируем текст
		resultText = RemoveEmptyLines(resultMessage.String())

		return nil
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Не удалось запустить клиент Telega App. Ошибка: %v", err))
		log.Fatal("")
	}

	return resultText, nil
}

// ---------------------------------------------------------------------------------------------------------------------

// RemoveEmptyLines Удаляет все пустые строки из текста, и ненужные символы, проверяет оимит в 5000mb
func RemoveEmptyLines(input string) string {
	const lbl = "internal/app_telega/app_telega.go/RemoveEmptyLines()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	// 1. Удаляем пустые строки
	lines := strings.Split(input, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	text := strings.Join(nonEmptyLines, "\n")
	slog.Info("Пустые строки из текста успешно удалены")

	// 2. Удаляем все символы, кроме русских букв, пробелов и знаков препинания
	var cleanedText strings.Builder
	for _, char := range text {
		if (char >= 'а' && char <= 'я') ||
			(char >= 'А' && char <= 'Я') ||
			char == 'ё' || char == 'Ё' ||
			string(char) == " " ||
			string(char) == "." || string(char) == "," || string(char) == "!" ||
			string(char) == "?" || string(char) == ";" || string(char) == ":" ||
			string(char) == "—" || string(char) == "–" || string(char) == "-" ||
			string(char) == "(" || string(char) == ")" || string(char) == "«" ||
			string(char) == "»" || string(char) == "\"" || string(char) == "'" ||
			char == '\n' {
			cleanedText.WriteRune(char)
		}
	}

	// Получаем итоговый текст
	result := cleanedText.String()
	slog.Info("Разные символы успешно удалены")

	// 3. Ограничиваем длину текста до 4800 байт
	if len(result) > 4800 {
		slog.Warn("Текст превышает лимит 4800 байт, выполняется обрезка")
		// Обрезаем до 4500 байт, сохраняя целостность UTF-8 символов
		result = result[:4800]
		// Корректируем обрезку, если последний байт — часть многобайтового символа
		for len(result) > 0 && utf8.RuneStart(result[len(result)-1]) == false {
			result = result[:len(result)-1]
		}
	}

	slog.Info("Итоговый текст готов")
	return result
}
