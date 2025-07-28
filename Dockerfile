# 1. Используем официальный образ Go
FROM golang:1.23.2

WORKDIR /app

# Копируем только нужные файлы для сборки
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные файлы
COPY . .

# Собираем бинарник с помощью make
RUN make build

# Указываем исполняемый файл из Makefile
CMD ["./ExchangeRate"]