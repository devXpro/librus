# Базовый образ
FROM golang:1.16-alpine

# Установка зависимостей для chromedp
RUN apk add --no-cache chromium \
    && apk add --no-cache xvfb \
    && apk add --no-cache wait4ports

# Установка зависимостей приложения
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходных файлов
COPY . .

# Сборка приложения
RUN go build -o /app
ENV TELEGRAM_TOKEN=token
ENV MONGO_HOST=mongodb
# Определение точки входа
CMD ["dumb-init", "/app"]