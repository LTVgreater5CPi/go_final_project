package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
func MakeHandler(fn func(http.ResponseWriter, *http.Request, *sql.DB), db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db)
	}
}

func main() {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	appPassword = os.Getenv("TODO_PASSWORD")
	if appPassword == "" {
		log.Println("The TODO_PASSWORD variable is not set. Authentication is disabled")
	}
	db, err := setupDB()
	if err != nil {
		log.Fatalf("DB configuration error: %v", err)
	}
	defer db.Close()

	webDir := "./web"
	fileServer := http.FileServer(http.Dir(webDir))
	http.Handle("/", fileServer)

	// API рабы с аутентификацией
	http.HandleFunc("/api/nextdate", NextDateHandler)
	http.HandleFunc("/api/task", authMidW(MakeHandler(taskHandler, db)))
	http.HandleFunc("/api/tasks", authMidW(MakeHandler(GetTaskHandler, db)))
	http.HandleFunc("/api/task/done", authMidW(MakeHandler(TaskDoneHandler, db)))
	http.HandleFunc("/api/signin", MakeHandler(signInHandler, db))

	log.Printf("Starting the server on the port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
