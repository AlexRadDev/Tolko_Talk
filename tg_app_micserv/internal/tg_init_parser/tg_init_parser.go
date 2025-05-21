// Реализация клиента Telegram

package tg_parser

import (
	"context"

	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"tg_app_micserv/internal/model"
	"tg_app_micserv/tools/logger"
)

// Client реализует Сборщик сообщений для Telegram
type Client struct {
	client         *telegram.Client
	tgAppClient    *telegram.Client
	phone          string
	twoFacPassword string
	closer         chan struct{} // Для управления завершением
	once           sync.Once     // Для идемпотентности Close
}

// NewClient создает новый клиент Telegram
func NewClient(tgApiID int, tgApiHash, phone, twoFacPassword string, storage telegram.SessionStorage) *Client {
	tgAppClient := telegram.NewClient(tgApiID, tgApiHash, telegram.Options{
		SessionStorage: storage,
	})
	return &Client{
		client:         tgAppClient,
		tgAppClient:    tgAppClient,
		phone:          phone,
		twoFacPassword: twoFacPassword,
		closer:         make(chan struct{}),
	}
}

// PostParser извлекает сообщения из канала Telegram (парсит заданный канал)
func (c *Client) PostParser(ctx context.Context, tgNameChannel string, timePeriod time.Duration) ([]tg_post_model.Message, error) {
	const lbl = "tg_app_micserv/cmd/main.go/main()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	// Создаем срез для хранения сообщений
	var messages []tg_post_model.Message
	// Запускаем клиента Telegram в асинхронном режиме
	err := c.tgAppClient.Run(ctx, func(ctx context.Context) error {
		status, err := c.tgAppClient.Auth().Status(ctx) // Проверяем статус аутентификации текущего клиента
		if err != nil {
			slog.Error("не удалось проверить статус аутентификации")
			return fmt.Errorf("не удалось проверить статус аутентификации: %w", err)
		}

		// Если пользователь не авторизован — запускаем процесс авторизации
		if !status.Authorized {
			authenticator := &customCodeAuthenticator{
				phone:          c.phone,
				twoFacPassword: c.twoFacPassword,
			}
			flow := auth.NewFlow(authenticator, auth.SendCodeOptions{})
			if err := flow.Run(ctx, c.tgAppClient.Auth()); err != nil {
				slog.Error("Авторизация не удалась")
				return fmt.Errorf("authorization failed: %w", err)
			}
			slog.Info("Успешно авторизовано")
		} else {
			slog.Info("Уже авторизован, использует существующую сессию")
		}

		// Получаем API клиента Telegram
		api := c.tgAppClient.API()
		usernameRequest := &tg.ContactsResolveUsernameRequest{
			Username: tgNameChannel,
		}
		// Отправляем запрос разрешения username
		resolved, err := api.ContactsResolveUsername(ctx, usernameRequest)
		if err != nil {
			return fmt.Errorf("не удалось разрешить tgNameChannel: %w", err)
		}

		// Переменная для хранения входного представления канала или пользователя
		var inputPeer tg.InputPeerClass
		if len(resolved.Chats) > 0 {
			chat, ok := resolved.Chats[0].(*tg.Channel)
			if !ok {
				slog.Info("Ожидаемый тип канала, получен")
				return fmt.Errorf("Ожидаемый тип канала, получен %T", resolved.Chats[0])
			}
			inputPeer = &tg.InputPeerChannel{
				ChannelID:  chat.ID,
				AccessHash: chat.AccessHash,
			}
		} else if len(resolved.Users) > 0 {
			user, ok := resolved.Users[0].(*tg.User)
			if !ok {
				return fmt.Errorf("Ожидаемый тип пользователя, получен %T", resolved.Users[0])
			}
			inputPeer = &tg.InputPeerUser{
				UserID:     user.ID,
				AccessHash: user.AccessHash,
			}
		} else {
			return fmt.Errorf("Канал с таким именем не найден")
		}

		// Вычисляем временной порог, чтобы брать только последние сообщения
		timeThreshold := time.Now().Add(-timePeriod)

		// Запрашиваем историю сообщений из канала / пользователя
		tgMessages, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:  inputPeer,
			Limit: 100,
		})
		if err != nil {
			slog.Error("Не удалось получить посты из канала")
			return fmt.Errorf("Не удалось получить посты из канала: %w", err)
		}

		// Создаем срез для хранения сообщений Telegram API
		var msgSlice []tg.MessageClass
		switch m := tgMessages.(type) {
		case *tg.MessagesMessages:
			msgSlice = m.Messages
		case *tg.MessagesMessagesSlice:
			msgSlice = m.Messages
		case *tg.MessagesChannelMessages:
			msgSlice = m.Messages
		default:
			return fmt.Errorf("неожиданный тип сообщения: %T", m)
		}

		// Перебираем все сообщения из полученного среза
		for _, msg := range msgSlice {
			message, ok := msg.(*tg.Message)
			if !ok {
				continue
			}

			// Преобразуем дату сообщения в time.Time
			msgTime := time.Unix(int64(message.Date), 0)
			if msgTime.Before(timeThreshold) {
				break
			}

			// Добавляем сообщение в результат с нужными полями
			messages = append(messages, tg_post_model.Message{
				Text:      message.Message,
				Timestamp: msgTime,
			})
		}

		return nil
	})
	if err != nil {
		slog.Error("Не удалось запустить Telegram App Client")
		return nil, fmt.Errorf("не удалось запустить Telegram tgAppClient: %w", err)
	}

	// Возвращаем срез сообщений
	return messages, nil
}

// Close закрывает клиент
func (c *Client) Close(ctx context.Context) error {
	c.once.Do(func() {
		close(c.closer)

		// Создаем новый контекст с таймаутом для завершения
		closeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// Пытаемся выполнить завершение клиента через Run
		err := c.client.Run(closeCtx, func(ctx context.Context) error {
			// Клиент автоматически сохраняет сессию при завершении
			return nil
		})
		if err != nil {
			//c.logger.Error("Ошибка при завершении Telegram-клиента", "error", err)
			return
		}

		//c.logger.Info("Telegram-клиент успешно завершен")
	})
	return nil
}

// customCodeAuthenticator имплементирует auth.CodeAuthenticator
type customCodeAuthenticator struct {
	phone          string
	twoFacPassword string
}

func (c *customCodeAuthenticator) Phone(_ context.Context) (string, error) {
	return c.phone, nil
}

func (c *customCodeAuthenticator) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	return "", fmt.Errorf("Ввод кода не поддерживается в режиме микросервиса")
}

func (c *customCodeAuthenticator) Password(_ context.Context) (string, error) {
	return c.twoFacPassword, nil
}

func (c *customCodeAuthenticator) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	//c.logger.Info("Terms of service accepted", "text", tos.Text)
	return nil
}

func (c *customCodeAuthenticator) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("Регистрация не поддерживается в режиме микросервиса")
}
