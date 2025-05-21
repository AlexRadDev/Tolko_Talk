telegram-bot-service/
├── cmd/
│   └── app/
│       └── main.go              # Точка входа приложения, настройка и запуск микросервиса
├── internal/
│   ├── domain/
│   │   ├── model/
│   │   │   └── bot_request.go   # Доменная модель запроса бота
│   │   └── repository/
│   │       └── user_request.go  # Интерфейс репозитория для хранения состояния пользователей
│   ├── usecase/
│   │   └── bot_usecase.go       # Бизнес-логика обработки сообщений бота
│   ├── delivery/
│   │   └── http/
│   │       └── router.go        # HTTP-роутер для обработки входящих обновлений Telegram
│   └── infrastructure/
│       ├── bot/
│       │   └── telegram.go      # Инициализация Telegram-бота
│       ├── config/
│       │   └── config.go        # Загрузка конфигурации из .env
│       ├── logger/
│       │   └── logger.go        # Настройка логгера
│       └── repository/
│           └── in_memory.go     # Реализация in-memory репозитория
├── .env                         # Файл с переменными окружения
├── go.mod                       # Модуль Go
└── README.md                    # Описание проекта