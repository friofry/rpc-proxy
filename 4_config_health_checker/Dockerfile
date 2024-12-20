# Используем официальный образ Go для сборки
FROM golang:1.20 as builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем только исходный код (без go.mod и go.sum)
COPY main.go .

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

# Используем минимальный образ для запуска приложения
FROM alpine:3.17

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарник из стадии сборки
COPY --from=builder /app/app /app/app

# Копируем необходимые статические файлы
COPY checker_config.json default_providers.json /app/

# Экспонируем порт
EXPOSE 8080

# Запуск приложения
ENTRYPOINT ["/app/app"]
