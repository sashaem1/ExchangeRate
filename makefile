# Путь к main.go
MAIN_PATH=cmd/ExchangeRate

# Имя исполняемого файла
BINARY_NAME=ExchangeRate

# Сборка
build:
	go build -o $(BINARY_NAME) ./$(MAIN_PATH)

# Запуск
run:
	go run ./$(MAIN_PATH)