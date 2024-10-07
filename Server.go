package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Инициализация базы данных
func initDB() {
	// Получаем путь к базе данных из переменной окружения TODO_DBFILE
	dbPath := os.Getenv("TODO_DBFILE")
	if dbPath == "" {
		// Если переменная не установлена, используем путь по умолчанию
		dbPath = "./scheduler.db"
	}

	// Открываем базу данных
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Ошибка при подключении к базе данных: %s\n", err)
		os.Exit(1)
	}

	// Создаем таблицу, если она не существует
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			title TEXT NOT NULL,
			comment TEXT,
			repeat TEXT CHECK(LENGTH(repeat) <= 128)
		);
		CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
	`)
	if err != nil {
		fmt.Printf("Ошибка при создании таблицы: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("База данных подключена: %s\n", dbPath)
}

func Server() {
	// Инициализация базы данных
	initDB()
	defer db.Close()

	// Директория, откуда будут браться статические файлы
	webDir := "./web"

	// Чтение переменной окружения TODO_PORT
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Создаем маршрутизатор Chi для обработки API
	r := chi.NewRouter()
	r.Post("/api/task", handleTask)

	// Регистрируем маршруты для API
	r.Post("/tasks", addTask)
	r.Get("/tasks", getTasks)
	r.Get("/tasks/{id}", getTaskByID)
	r.Delete("/tasks/{id}", deleteTask)
	r.Put("/tasks/{id}", updateTask)
	r.Patch("/tasks/{id}/complete", completeTask)
	r.Get("/api/nextdate", handleNextDate)
	// Обработчик для возврата статических файлов
	fs := http.FileServer(http.Dir(webDir))
	r.Handle("/*", fs)

	// Запуск сервера
	fmt.Printf("Сервер запущен на порту %s\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s\n", err)
	}
}

func main() {
	Server()
}
