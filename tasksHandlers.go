package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func errResp(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	http.Error(w, fmt.Sprintf(`{"error":"%s"}`, message), statusCode)
	log.Printf("Error: %s | Status code: %d", message, statusCode)
}

// костылем из-за особенности тестов
func successResp(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if data == nil {
		w.Write([]byte(`{}`))
		log.Printf("Success response sent: {}")
		return
	}
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Success response sent: %+v", data)
}

func compareDates(t1, t2 time.Time) (time.Time, time.Time) {
	date1 := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	date2 := time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())
	return date1, date2
}

// Царь хэндлер
func taskH(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	switch r.Method {
	case http.MethodPost:
		addTaskH(w, r, db)
	case http.MethodGet:
		getTaskH(w, r, db)
	case http.MethodPut:
		editTaskH(w, r, db)
	case http.MethodDelete:
		deleteTaskH(w, r, db)
	default:
		errResp(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Следующая дата
func nextDateH(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now, err := time.Parse(dateFormat, nowStr)
	if err != nil {
		errResp(w, "Invalid 'now' date format", http.StatusBadRequest)
		return
	}
	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		errResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, nextDate)
}

// Добавление задачи
func addTaskH(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&task)
	if err != nil {
		errResp(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		errResp(w, "Task title is not specified", http.StatusBadRequest)
		return
	}
	if task.Date == "" {
		task.Date = time.Now().Format(dateFormat)
	} else {
		_, err := time.Parse(dateFormat, task.Date)
		if err != nil {
			errResp(w, "Invalid date format", http.StatusBadRequest)
			return
		}
	}

	now := time.Now()
	taskDate, _ := time.Parse(dateFormat, task.Date)
	taskDate_fix, now_fix := compareDates(taskDate, now)
	if taskDate_fix.Before(now_fix) {
		if task.Repeat == "" {
			task.Date = now.Format(dateFormat)
		} else {
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				errResp(w, err.Error(), http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	}
	id, err := AddTask(db, task)
	if err != nil {
		errResp(w, fmt.Sprintf("Error adding task to the database: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	task.ID = strconv.Itoa(int(id))
	successResp(w, task)
}

// Получение задачи по id
func getTaskH(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		errResp(w, "Task ID is not specified", http.StatusBadRequest)
		return
	}
	task, err := GetTaskByID(db, id)
	if err != nil {
		if err.Error() == "task not found" {
			errResp(w, "Task not found", http.StatusNotFound)
		} else {
			errResp(w, fmt.Sprintf("Error retrieving task: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}
	successResp(w, task)
}

// Редактирование задачи
func editTaskH(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&task)
	if err != nil {
		errResp(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	if task.ID == "" {
		errResp(w, "Task ID is not specified", http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		errResp(w, "Task title is not specified", http.StatusBadRequest)
		return
	}
	if task.Date == "" {
		task.Date = time.Now().Format(dateFormat)
	} else {
		_, err := time.Parse(dateFormat, task.Date)
		if err != nil {
			errResp(w, "Invalid date format", http.StatusBadRequest)
			return
		}
	}

	now := time.Now()
	taskDate, _ := time.Parse(dateFormat, task.Date)
	if taskDate.Before(now) {
		if task.Repeat == "" {
			task.Date = now.Format(dateFormat)
		} else {
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				errResp(w, err.Error(), http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	}
	err = UpdateTask(db, task)
	if err != nil {
		errResp(w, fmt.Sprintf("Error updating task in the database: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	successResp(w, nil)
}

// Удаление задачи
func deleteTaskH(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		errResp(w, "Task ID is not specified", http.StatusBadRequest)
		return
	}
	err := DeleteTask(db, id)
	if err != nil {
		if err.Error() == "task not found" {
			errResp(w, "Task with the specified ID does not exist", http.StatusNotFound)
			return
		}
		errResp(w, fmt.Sprintf("Error deleting task from the database: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	successResp(w, nil)
}

// Получаем задачи из БД с поиском
func tasksH(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	search := r.URL.Query().Get("search")
	tasks, err := GetTasks(db, search)
	if err != nil {
		errResp(w, fmt.Sprintf("Error fetching tasks from the database: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"tasks": tasks,
	}
	if len(tasks) == 0 {
		response["tasks"] = []Task{}
		log.Println("No tasks found in tasksHandler")
	}
	successResp(w, response)
}

// Выполнение задачи
func taskDoneH(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		errResp(w, "Task ID is not specified", http.StatusBadRequest)
		return
	}
	task, err := GetTaskByID(db, id)
	if err != nil {
		if err.Error() == "task not found" {
			errResp(w, "Task not found", http.StatusNotFound)
		} else {
			errResp(w, fmt.Sprintf("Error retrieving task: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	if task.Repeat == "" {
		err := DeleteTask(db, id)
		if err != nil {
			errResp(w, fmt.Sprintf("Error deleting task from the database: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		successResp(w, nil)
		return
	}
	now := time.Now()
	nextDate, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		errResp(w, fmt.Sprintf("Error calculating the next date: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	task.Date = nextDate
	err = UpdateTask(db, task)
	if err != nil {
		errResp(w, fmt.Sprintf("Error updating task in the database: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	successResp(w, nil)
}
