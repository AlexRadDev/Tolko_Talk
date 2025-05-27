# Используем официальный образ Go 1.23.6 для сборки
FROM golang:1.23.6 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем исполняемый файл с оптимизацией для размера
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w' -o main ./cmd/main.go

# Создаём финальный минимальный образ
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем собранный бинарный файл из стадии builder
COPY --from=builder /app/main .

# Копируем session.json в рабочую директорию контейнера
COPY session.json ./session.json

# Указываем команду для запуска
CMD ["./main"]