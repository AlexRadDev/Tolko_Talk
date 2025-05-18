# Используем официальный образ Go версии 1.23.6 как базовый
FROM golang:1.23.6

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем весь исходный код проекта
COPY . .

# Компилируем приложение
RUN go build -o /text-to-speech-microservice ./cmd

# Указываем команду для запуска приложения
CMD ["/text-to-speech-microservice"]