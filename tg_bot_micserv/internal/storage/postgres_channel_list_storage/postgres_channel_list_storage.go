package postgres_channel_list_storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"tg_bot/internal/config"
	"tg_bot/tools/logger"
)

// PostgresStorage структура для работы с PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage создает новое подключение к PostgreSQL
func NewPostgresStorage(cfg *config.Config) (*PostgresStorage, error) {
	const lblConsume = "tg_bot_micserv/internal/storage/postgres_channel_list_storage/NewPostgresStorage()"
	myLogger := logger.NewColorLogger(lblConsume)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDBName, cfg.DBSSLMOD)
	myLogger.Info(fmt.Sprintf("Адрес DNS: %s", dsn))

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка подключения к базе Postgres: %s", err))
		return nil, fmt.Errorf("Ошибка подключения к базе Postgres: %w", err)
	}
	if err := db.Ping(); err != nil { // Проверка соединения
		myLogger.Error(fmt.Sprintf("Ошибка проверки пинга Postgres: %s", err))
		return nil, fmt.Errorf("Ошибка проверки пинга Postgres: %w", err)
	}

	storage := &PostgresStorage{db: db}

	return storage, nil
}

// Close закрывает соединение с БД
func (p *PostgresStorage) Close() error {
	return p.db.Close()
}

// AddChannel добавляет канал для чата
func (p *PostgresStorage) AddChannel(chatID int64, nameChannel, linkChannel string) {
	const lblConsume = "tg_bot_micserv/internal/storage/postgres_channel_list_storage/AddChannel()"
	myLogger := logger.NewColorLogger(lblConsume)

	query := `
        INSERT INTO channels_lists (chat_id, name_channel, link_channel, created_at) 
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (chat_id, name_channel) 
        DO UPDATE SET 
            link_channel = EXCLUDED.link_channel,
            updated_at = NOW()
    `

	_, err := p.db.Exec(query, chatID, nameChannel, linkChannel)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка сохранения канала: %s", err))
		return
	}

	myLogger.Info(fmt.Sprintf("Канал успешно сохранен. Имя канала: %s, Номер чата: %v", nameChannel, chatID))
}

// GetChannels получает все каналы для чата
func (p *PostgresStorage) GetChannels(chatID int64) map[string]string {
	const lblConsume = "tg_bot_micserv/internal/storage/postgres_channel_list_storage/GetChannels()"
	myLogger := logger.NewColorLogger(lblConsume)

	query := `
        SELECT name_channel, link_channel 
        FROM channels_lists
        WHERE chat_id = $1
        ORDER BY created_at DESC
    `

	rows, err := p.db.Query(query, chatID)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка получения канала из бд: %s", err))
		return make(map[string]string)
	}
	defer rows.Close()

	channels := make(map[string]string)

	for rows.Next() {
		var nameChannel, linkChannel string
		if err := rows.Scan(&nameChannel, &linkChannel); err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка сканирования строки ответа БД: %s", err))
			continue
		}
		channels[nameChannel] = linkChannel
	}

	if err := rows.Err(); err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка итерации строк: %s", err))
	}

	return channels
}

// DeleteChannels удаляет канал для чата
func (p *PostgresStorage) DeleteChannels(chatID int64, nameChannel string) {
	const lblConsume = "tg_bot_micserv/internal/storage/postgres_channel_list_storage/DeleteChannels()"
	myLogger := logger.NewColorLogger(lblConsume)

	query := `
        DELETE FROM channels_lists
        WHERE chat_id = $1 AND name_channel = $2
    `

	result, err := p.db.Exec(query, chatID, nameChannel)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка удаления канала из Postgres: %s", err))
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка при получении затронутых строк из Postgres: %s", err))
		return
	}

	if rowsAffected > 0 {
		myLogger.Error(fmt.Sprintf("Канал %s удален", nameChannel))
	} else {
		myLogger.Error(fmt.Sprintf("Канал %s не найден", nameChannel))
	}
}

// runMigrations запускает миграции базы данных
func (p *PostgresStorage) runMigrations() error {
	// Создание таблицы для отслеживания миграций
	createMigrationsTable := `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT NOW()
        )
    `

	if _, err := p.db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Список миграций
	migrations := []struct {
		version string
		sql     string
	}{
		{
			version: "001_create_channels_table",
			sql: `
                CREATE TABLE IF NOT EXISTS channels (
                    id SERIAL PRIMARY KEY,
                    chat_id BIGINT NOT NULL,
                    name_channel VARCHAR(255) NOT NULL,
                    link_channel TEXT NOT NULL,
                    created_at TIMESTAMP DEFAULT NOW(),
                    updated_at TIMESTAMP DEFAULT NOW(),
                    UNIQUE(chat_id, name_channel)
                );
                
                CREATE INDEX IF NOT EXISTS idx_channels_chat_id ON channels(chat_id);
                CREATE INDEX IF NOT EXISTS idx_channels_name ON channels(name_channel);
            `,
		},
	}

	// Применение миграций
	for _, migration := range migrations {
		// Проверяем, была ли уже применена миграция
		var count int
		err := p.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", migration.version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", migration.version, err)
		}

		if count > 0 {
			log.Printf("Migration %s already applied", migration.version)
			continue
		}

		// Применяем миграцию
		if _, err := p.db.Exec(migration.sql); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.version, err)
		}

		// Записываем информацию о применении миграции
		if _, err := p.db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.version, err)
		}

		log.Printf("Migration %s applied successfully", migration.version)
	}

	return nil
}
