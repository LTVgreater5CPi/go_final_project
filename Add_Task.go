package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TaskInput — структура для десериализации JSON-запроса
type TaskInput struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat"`
}

// Обработчик для POST-запросов /api/task
func handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var taskInput TaskInput
	if err := json.NewDecoder(r.Body).Decode(&taskInput); err != nil {
		writeError(w, "Ошибка десериализации JSON")
		return
	}

	// Проверка на наличие обязательного поля title
	if taskInput.Title == "" {
		writeError(w, "Не указан заголовок задачи")
		return
	}

	// Получение текущей даты
	now := time.Now()
	nowFormatted := now.Format("20060102")
	var taskDate time.Time
	var taskDateStr string

	// Обработка даты "today" и других дат
	if taskInput.Date == "" || taskInput.Date == "today" {
		// Если дата не указана или указана как "today", используем текущую дату
		taskDate = now
		taskDateStr = nowFormatted
	} else {
		// Парсим дату
		var err error
		taskDate, err = time.Parse("20060102", taskInput.Date)
		if err != nil {
			writeError(w, "Дата указана в некорректном формате")
			return
		}

		// Если дата меньше текущей, корректируем в зависимости от правила повторения
		if taskDate.Before(now) {
			if taskInput.Repeat == "" {
				taskDateStr = nowFormatted // Без правила повторения дата должна быть сегодняшней
			} else if taskInput.Date == nowFormatted && taskInput.Repeat == "d 1" {
				// Если дата "today" и правило "d 1", оставляем сегодняшнюю дату
				taskDateStr = nowFormatted
			} else {
				// Если правило повторения указано, используем функцию NextDate
				nextDateStr, err := NextDate(now, taskInput.Date, taskInput.Repeat)
				if err != nil {
					writeError(w, "Некорректное правило повторения")
					return
				}
				taskDateStr = nextDateStr
			}
		} else {
			// Если дата больше или равна текущей, используем её
			taskDateStr = taskInput.Date
		}
	}

	// Вставляем задачу в базу данных
	res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		taskDateStr, taskInput.Title, taskInput.Comment, taskInput.Repeat)
	if err != nil {
		writeError(w, "Ошибка добавления задачи в базу данных")
		return
	}

	// Получаем ID созданной записи
	id, err := res.LastInsertId()
	if err != nil {
		writeError(w, "Ошибка получения идентификатора задачи")
		return
	}

	// Возвращаем успешный ответ с ID задачи
	response := map[string]string{
		"id": fmt.Sprintf("%d", id),
	}
	writeJSON(w, response)
}

// writeError отправляет сообщение об ошибке в формате JSON
func writeError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// writeJSON отправляет успешный ответ в формате JSON
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
