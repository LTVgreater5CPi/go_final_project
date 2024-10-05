package main

import (
	"fmt"
	"net/http"
	"os"
)

func Server() {
	// Директория, откуда будут браться статические файлы
	webDir := "./web"

	// Чтение переменной окружения TODO_PORT
	port := os.Getenv("TODO_PORT")
	if port == "" {
		// Если переменная окружения не задана, используем порт по умолчанию
		port = "7540"
	}

	// Создаем обработчик для возврата статических файлов из директории web
	fs := http.FileServer(http.Dir(webDir))

	// Регистрируем обработчик на все маршруты
	http.Handle("/", fs)

	// Запускаем сервер
	fmt.Printf("Сервер запущен на порту %s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s\n", err)
	}
}
