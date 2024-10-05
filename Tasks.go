package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"strings"

	"github.com/go-chi/chi/v5"
)

type Task struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Note        string `json:"note,omitempty"`
	DateDead    string `json:"dateDead,omitempty"`
	IsComplete  bool   `json:"isComplete"`
}

var tasks = map[string]Task{}

func getTasks(w http.ResponseWriter, r *http.Request) {
	var taskErray []Task
	for _, task := range tasks {
		taskErray = append(taskErray, task)
	}
	resp, err := json.Marshal(taskErray)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// в заголовок записываем тип контента, у нас это данные в формате JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
func getUnusedIDs() []string {
	const maxID = 5
	unusedIDs := make([]string, 0, maxID)
	for i := 1; i <= maxID; i++ {
		id := fmt.Sprintf("%d", i)
		if _, exists := tasks[id]; !exists {
			unusedIDs = append(unusedIDs, id)
		}
	}
	return unusedIDs
}
func addTasks(w http.ResponseWriter, r *http.Request) {
	var task Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, idExist := tasks[task.ID]; idExist {
		errMsg := fmt.Sprintf("Задача с id %s уже существует.", task.ID)
		unusedIDs := getUnusedIDs()
		if len(unusedIDs) > 0 {
			errMsg += " Доступные ID: " + strings.Join(unusedIDs, ", ")
		}
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	tasks[task.ID] = task
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)

}
func getTasksID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, ok := tasks[id]
	if !ok {
		http.Error(w, "Такой задачи нет", http.StatusNoContent)
		return
	}
	resp, err := json.Marshal(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
func deleteTasksID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, ok := tasks[id]
	if !ok {
		http.Error(w, "Такой задачи нет", http.StatusNotFound)
		return
	}
	delete(tasks, id)
	w.WriteHeader(http.StatusOK)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, exists := tasks[id]
	if !exists {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	var updatedTask Task
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Обновление задачи
	tasks[id] = updatedTask
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTask)
}

func completeTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, exists := tasks[id]
	if !exists {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	// Отмечаем задачу как выполненную
	task.IsComplete = true
	tasks[id] = task

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func main() {
	r := chi.NewRouter()

	r.Get("/tasks", getTasks)
	r.Post("/tasks", addTasks)
	r.Get("/tasks/{id}", getTasksID)
	r.Delete("/tasks/{id}", deleteTasksID)
	r.Put("/tasks/{id}", updateTask)
	r.Patch("/tasks/{id}/complete", completeTask)
	Server()
}
