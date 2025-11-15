# Build stage
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /app/bin/server .

# Копируем .env файл
COPY --from=builder /app/environment/.env ./environment/.env

# Устанавливаем переменную окружения для пути к .env
ENV ENV_FILE=./environment/.env

# Экспортируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./server"]

