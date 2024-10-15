package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

const dateFormat = "20060102"

// Инициализация и создание БД
func setupDB() (*sql.DB, error) {
	dbPath := os.Getenv("TODO_DBFILE")
	if dbPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("directory error: %v", err)
		}
		dbPath = filepath.Join(cwd, "scheduler.db")
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the DB: %w", err)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		createTableQuery := `
		CREATE TABLE scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			title TEXT NOT NULL,
			comment TEXT,
			repeat TEXT(128)
		);
		CREATE INDEX idx_date ON scheduler(date);
		`
		if _, err := db.Exec(createTableQuery); err != nil {
			db.Close()
			return nil, fmt.Errorf("error creating the table: %v", err)
		}
	}
	return db, nil
}

// Добавляем новую задачу в БД
func AddTask(db *sql.DB, task Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Обновляет данные задачи
func UpdateTask(db *sql.DB, task Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no rows were updated, task not found")
	}
	return nil
}

// Удаляем задачу по ID
func DeleteTask(db *sql.DB, id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affectedRows == 0 {
		return errors.New("task not found")
	}
	return nil
}

// Получаем задачу по ID
func GetTaskByID(db *sql.DB, id string) (Task, error) {
	var task Task
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err == sql.ErrNoRows {
		return task, errors.New("task not found")
	}
	return task, err
}

// Получаем список задач с возможностью поиска
func GetTasks(db *sql.DB, search string) ([]Task, error) {
	var tasks []Task
	var query string
	var args []interface{}

	if search != "" {
		if searchDate, err := time.Parse("02.01.2006", search); err == nil {
			query = "SELECT * FROM scheduler WHERE date = ? ORDER BY date LIMIT 50"
			args = append(args, searchDate.Format("20060102"))
		} else {
			query = "SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT 50"
			searchPattern := "%" + search + "%"
			args = append(args, searchPattern, searchPattern)
		}
	} else {
		query = "SELECT * FROM scheduler ORDER BY date LIMIT 50"
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var task Task
		var id int
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		task.ID = strconv.Itoa(id)
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}
