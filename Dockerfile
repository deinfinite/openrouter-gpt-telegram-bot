# Используем официальный образ Golang как базовый
FROM golang:1.22

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum и устанавливаем зависимости
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Копируем остальные файлы проекта
COPY . .

# Собираем приложение
RUN go build -o /openrouter-gpt-telegram-bot

# Указываем команду для запуска приложения
CMD ["/openrouter-gpt-telegram-bot"]
