// Реализация клиента Telegram

package tg_app_init

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"

	"tg_app_micserv/internal/config"
	tg_post_model "tg_app_micserv/internal/model"
	"tg_app_micserv/tools/logger"
)

// TGAppClient реализует Сборщик сообщений для Telegram
//type TGAppClient struct {
//	client         *telegram.Client
//	tgAppClient    *telegram.Client
//	phone          string
//	twoFacPassword string
//	closer         chan struct{} // Для управления завершением
//	once           sync.Once     // Для идемпотентности Close
//}
//
// NewTGAppClient создает новый клиент Telegram
//func NewTGAppClient(cfg *config.Config, storage telegram.SessionStorage, path string) *TGAppClient {
//	if storage == nil {
//		storage = &session.FileStorage{Path: path} // session.json Сохраняем сессию в session.json (/app/session.json для докер)
//	}
//	tgAppClient := telegram.NewClient(cfg.API_ID, cfg.API_Hash, telegram.Options{
//		SessionStorage: storage,
//	})
//	return &TGAppClient{
//		client:         tgAppClient,
//		tgAppClient:    tgAppClient,
//		phone:          cfg.Phone,
//		twoFacPassword: cfg.Two_F_Password,
//		closer:         make(chan struct{}),
//	}
//}
//
//// PostParser извлекает сообщения из канала Telegram (парсит заданный канал)
//func (c *TGAppClient) PostParser(ctx context.Context, tgNameChannel string, timePeriod time.Duration) ([]tg_post_model.Message, error) {
//	const lbl = "tg_app_micserv/internal/tg_init_parser/tg_init_parser.go/PostParser()"
//	mylogger := logger.NewColorLogger(lbl)
//
//	// Создаем срез для хранения сообщений
//	var messages []tg_post_model.Message
//	// Запускаем клиента Telegram в асинхронном режиме
//	err := c.tgAppClient.Run(ctx, func(ctx context.Context) error {
//		status, err := c.tgAppClient.Auth().Status(ctx) // Проверяем статус аутентификации текущего клиента
//
//		if err != nil {
//			mylogger.Error("не удалось проверить статус аутентификации")
//			return fmt.Errorf("не удалось проверить статус аутентификации: %w", err)
//		}
//
//		// Если пользователь не авторизован — запускаем процесс авторизации
//		if !status.Authorized {
//			mylogger.Info("Пользователь не авторизован!!!")
//			authenticator := &customCodeAuthenticator{
//				phone:          c.phone,
//				twoFacPassword: c.twoFacPassword,
//			}
//			flow := auth.NewFlow(authenticator, auth.SendCodeOptions{})
//			if err := flow.Run(ctx, c.tgAppClient.Auth()); err != nil {
//				mylogger.Error("Авторизация не удалась")
//				return fmt.Errorf("authorization failed: %w", err)
//			}
//			mylogger.Info("Успешно авторизовано")
//		} else {
//			mylogger.Info("Уже авторизован, использует существующую сессию")
//		}
//
//		// Получаем API клиента Telegram
//		api := c.tgAppClient.API()
//		usernameRequest := &tg.ContactsResolveUsernameRequest{
//			Username: tgNameChannel,
//		}
//		// Отправляем запрос разрешения username
//		resolved, err := api.ContactsResolveUsername(ctx, usernameRequest)
//		if err != nil {
//			return fmt.Errorf("не удалось разрешить tgNameChannel: %w", err)
//		}
//
//		// Переменная для хранения входного представления канала или пользователя
//		var inputPeer tg.InputPeerClass
//		if len(resolved.Chats) > 0 {
//			chat, ok := resolved.Chats[0].(*tg.Channel)
//			if !ok {
//				mylogger.Info("Ожидаемый тип канала, получен")
//				return fmt.Errorf("Ожидаемый тип канала, получен %T", resolved.Chats[0])
//			}
//			inputPeer = &tg.InputPeerChannel{
//				ChannelID:  chat.ID,
//				AccessHash: chat.AccessHash,
//			}
//		} else if len(resolved.Users) > 0 {
//			user, ok := resolved.Users[0].(*tg.User)
//			if !ok {
//				return fmt.Errorf("Ожидаемый тип пользователя, получен %T", resolved.Users[0])
//			}
//			inputPeer = &tg.InputPeerUser{
//				UserID:     user.ID,
//				AccessHash: user.AccessHash,
//			}
//		} else {
//			return fmt.Errorf("Канал с таким именем не найден")
//		}
//
//		// Вычисляем временной порог, чтобы брать только последние сообщения
//		timeThreshold := time.Now().Add(-timePeriod)
//
//		// Запрашиваем историю сообщений из канала / пользователя
//		tgMessages, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
//			Peer:  inputPeer,
//			Limit: 100,
//		})
//		if err != nil {
//			mylogger.Error("Не удалось получить посты из канала")
//			return fmt.Errorf("Не удалось получить посты из канала: %w", err)
//		}
//
//		// Создаем срез для хранения сообщений Telegram API
//		var msgSlice []tg.MessageClass
//		switch m := tgMessages.(type) {
//		case *tg.MessagesMessages:
//			msgSlice = m.Messages
//		case *tg.MessagesMessagesSlice:
//			msgSlice = m.Messages
//		case *tg.MessagesChannelMessages:
//			msgSlice = m.Messages
//		default:
//			return fmt.Errorf("неожиданный тип сообщения: %T", m)
//		}
//
//		// Перебираем все сообщения из полученного среза
//		for _, msg := range msgSlice {
//			message, ok := msg.(*tg.Message)
//			if !ok {
//				continue
//			}
//
//			// Преобразуем дату сообщения в time.Time
//			msgTime := time.Unix(int64(message.Date), 0)
//			if msgTime.Before(timeThreshold) {
//				break
//			}
//
//			// Добавляем сообщение в результат с нужными полями
//			messages = append(messages, tg_post_model.Message{
//				Text:      message.Message,
//				Timestamp: msgTime,
//			})
//		}
//
//		return nil
//	})
//	if err != nil {
//		mylogger.Error("Не удалось запустить Telegram App TGAppClient")
//		return nil, fmt.Errorf("не удалось запустить Telegram tgAppClient: %w", err)
//	}
//
//	// Возвращаем срез сообщений
//	return messages, nil
//}

type TGAppClient struct {
	client     *telegram.Client
	closer     chan struct{}
	once       sync.Once
	api        *tg.Client // Исправлено: *tg.Client вместо *tg.API
	authStatus bool
	mu         sync.Mutex
}

// NewTGAppClient сщздает TG App клиента
func NewTGAppClient(cfg *config.Config, storage telegram.SessionStorage, path string) *TGAppClient {
	if storage == nil {
		storage = &session.FileStorage{Path: path}
	}
	tgAppClient := telegram.NewClient(cfg.API_ID, cfg.API_Hash, telegram.Options{
		SessionStorage: storage,
	})
	return &TGAppClient{
		client:     tgAppClient,
		closer:     make(chan struct{}),
		api:        tgAppClient.API(),
		authStatus: false,
	}
}

// Start стартует TG App клиента
func (c *TGAppClient) Start(ctx context.Context, phone, twoFacPassword string) error {
	const lbl = "tg_app_micserv/internal/tg_init_parser/tg_init_parser.go/Start()"
	mylogger := logger.NewColorLogger(lbl)

	return c.client.Run(ctx, func(ctx context.Context) error {
		status, err := c.client.Auth().Status(ctx)
		if err != nil {
			mylogger.Error("Не удалось проверить статус аутентификации", "error", err)
			return fmt.Errorf("не удалось проверить статус аутентификации: %w", err)
		}

		c.mu.Lock()
		if !status.Authorized {
			mylogger.Info("Пользователь не авторизован, начинаем авторизацию")
			authenticator := &customCodeAuthenticator{
				phone:          phone,
				twoFacPassword: twoFacPassword,
			}
			flow := auth.NewFlow(authenticator, auth.SendCodeOptions{})
			if err := flow.Run(ctx, c.client.Auth()); err != nil {
				mylogger.Error("Авторизация не удалась", "error", err)
				c.mu.Unlock()
				return fmt.Errorf("authorization failed: %w", err)
			}
			c.authStatus = true
			mylogger.Info("Успешно авторизовано")
		} else {
			c.authStatus = true
			mylogger.Info("Уже авторизован, использует существующую сессию")
		}
		c.mu.Unlock()

		<-c.closer // Ждём сигнала закрытия
		return nil
	})
}

func (c *TGAppClient) PostParser(ctx context.Context, tgNameChannel string, timePeriod time.Duration) ([]tg_post_model.Message, error) {
	const lbl = "tg_app_micserv/internal/tg_init_parser/tg_init_parser.go/PostParser()"
	mylogger := logger.NewColorLogger(lbl)

	c.mu.Lock()
	if !c.authStatus {
		mylogger.Error("Клиент не авторизован")
		c.mu.Unlock()
		return nil, fmt.Errorf("клиент не авторизован")
	}
	c.mu.Unlock()

	if ctx.Err() != nil {
		mylogger.Error("Контекст отменён", "error", ctx.Err())
		return nil, fmt.Errorf("контекст отменён: %w", ctx.Err())
	}

	var messages []tg_post_model.Message
	usernameRequest := &tg.ContactsResolveUsernameRequest{
		Username: tgNameChannel,
	}
	resolved, err := c.api.ContactsResolveUsername(ctx, usernameRequest)
	if err != nil {
		mylogger.Error("Не удалось разрешить tgNameChannel", "error", err)
		return nil, fmt.Errorf("не удалось разрешить tgNameChannel: %w", err)
	}

	var inputPeer tg.InputPeerClass
	if len(resolved.Chats) > 0 {
		chat, ok := resolved.Chats[0].(*tg.Channel)
		if !ok {
			mylogger.Info("Ожидаемый тип канала, получен")
			return nil, fmt.Errorf("Ожидаемый тип канала, получен %T", resolved.Chats[0])
		}
		inputPeer = &tg.InputPeerChannel{
			ChannelID:  chat.ID,
			AccessHash: chat.AccessHash,
		}
	} else if len(resolved.Users) > 0 {
		user, ok := resolved.Users[0].(*tg.User)
		if !ok {
			return nil, fmt.Errorf("Ожидаемый тип пользователя, получен %T", resolved.Users[0])
		}
		inputPeer = &tg.InputPeerUser{
			UserID:     user.ID,
			AccessHash: user.AccessHash,
		}
	} else {
		return nil, fmt.Errorf("Канал с таким именем не найден")
	}

	timeThreshold := time.Now().Add(-timePeriod)
	tgMessages, err := c.api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:  inputPeer,
		Limit: 100,
	})
	if err != nil {
		mylogger.Error("Не удалось получить посты из канала", "error", err)
		return nil, fmt.Errorf("Не удалось получить посты из канала: %w", err)
	}

	var msgSlice []tg.MessageClass
	switch m := tgMessages.(type) {
	case *tg.MessagesMessages:
		msgSlice = m.Messages
	case *tg.MessagesMessagesSlice:
		msgSlice = m.Messages
	case *tg.MessagesChannelMessages:
		msgSlice = m.Messages
	default:
		return nil, fmt.Errorf("неожиданный тип сообщения: %T", m)
	}

	for _, msg := range msgSlice {
		message, ok := msg.(*tg.Message)
		if !ok {
			continue
		}

		msgTime := time.Unix(int64(message.Date), 0)
		if msgTime.Before(timeThreshold) {
			break
		}

		messages = append(messages, tg_post_model.Message{
			Text:      message.Message,
			Timestamp: msgTime,
		})
	}

	return messages, nil
}

// Close закрывает клиент
func (c *TGAppClient) Close(ctx context.Context) error {
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
	//return "", fmt.Errorf("Ввод кода не поддерживается в режиме микросервиса")
	var code string
	fmt.Print("Введите код подтверждения: ")
	fmt.Scanln(&code)
	return code, nil
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
