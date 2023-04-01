# Сборка приложения
FROM golang:1.17 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

# Запуск приложения
FROM zenika/alpine-chrome

WORKDIR /app

# Установка необходимых зависимостей
RUN apk add --no-cache ca-certificates

# Копирование собранного приложения из предыдущей стадии
COPY --from=builder /app/main /app/main
ENV TELEGRAM_TOKEN=token
ENV MONGO_HOST=mongodb
# Запуск приложения
CMD ["./main"]
