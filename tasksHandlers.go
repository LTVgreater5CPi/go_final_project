package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Общий царь хендлер
func taskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	switch r.Method {
	case http.MethodPost:
		AddTaskHandler(w, r, db)
	case http.MethodGet:
		GetTaskHandler(w, r, db)
	case http.MethodPut:
		UpdateTaskHandler(w, r, db)
	case http.MethodDelete:
		DeleteTaskHandler(w, r, db)
	default:
		errResp(w, "The method is not supported", http.StatusMethodNotAllowed)
	}
}

// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
func errResp(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	http.Error(w, fmt.Sprintf(`{"error":"%s"}`, message), statusCode)
}

func successResp(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(data)
}

// Вычисляем следующую дату задачи
func GetNextDate(now time.Time, date string, repeat string) (string, error) {
	nextDate := now.AddDate(0, 0, 1)
	return nextDate.Format("20060102"), nil
}

// ОСНОВНЫЕ ОБРАБОТЧИКИ
// Обновление задачи
func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		errResp(w, "JSON decoding error", http.StatusBadRequest)
		return
	}
	if task.ID == "" {
		errResp(w, "The task ID is not specified", http.StatusBadRequest)
		return
	}
	if err := UpdateTask(db, task); err != nil {
		if err.Error() == "task not found" {
			errResp(w, "task not found", http.StatusNotFound)
		} else {
			errResp(w, fmt.Sprintf("Issue update error: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}
	successResp(w, `{}`)
}

// Завершения задачи
func TaskDoneHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		errResp(w, "The task ID is not specified", http.StatusBadRequest)
		return
	}
	task, err := GetTaskByID(db, id)
	if err != nil {
		if err.Error() == "task not found" {
			errResp(w, "task not found", http.StatusNotFound)
		} else {
			errResp(w, fmt.Sprintf("Task search error: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}
	if task.Repeat == "" {
		if err := DeleteTask(db, id); err != nil {
			errResp(w, fmt.Sprintf("Deletion error: %s", err.Error()), http.StatusInternalServerError)
		} else {
			successResp(w, `{}`)
		}
		return
	}
	nextDate, err := GetNextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		errResp(w, fmt.Sprintf("Error in calculating the next date: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	task.Date = nextDate
	if err := UpdateTask(db, task); err != nil {
		errResp(w, fmt.Sprintf("Issue update error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	successResp(w, `{}`)
}

// Получение задачи
func GetTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		errResp(w, "ID not specified", http.StatusBadRequest)
		return
	}
	task, err := GetTaskByID(db, id)
	if err != nil {
		if err.Error() == "task not found" {
			errResp(w, "task not found", http.StatusNotFound)
		} else {
			errResp(w, fmt.Sprintf("Error receiving the task: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}
	successResp(w, task)
}

// Добавление задачи
func AddTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		errResp(w, "JSON decoding error", http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		errResp(w, "The title is not specified", http.StatusBadRequest)
		return
	}
	if task.Date == "" {
		task.Date = time.Now().Format(dateFormat)
	} else if _, err := time.Parse(dateFormat, task.Date); err != nil {
		errResp(w, "Invalid date format", http.StatusBadRequest)
		return
	}
	if id, err := AddTask(db, task); err != nil {
		errResp(w, fmt.Sprintf("Adding error: %s", err.Error()), http.StatusInternalServerError)
	} else {
		task.ID = strconv.Itoa(int(id))
		successResp(w, task)
	}
}

// Удаление задачи
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		errResp(w, "ID not specified", http.StatusBadRequest)
		return
	}
	if err := DeleteTask(db, id); err != nil {
		if err.Error() == "task not found" {
			errResp(w, "task not found", http.StatusNotFound)
		} else {
			errResp(w, fmt.Sprintf("Task deletion error: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}
	successResp(w, `{}`)
}
