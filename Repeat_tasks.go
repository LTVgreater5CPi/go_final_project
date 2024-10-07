package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату на основе правила повторения
func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	// Парсим начальную дату
	startDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %v", err)
	}

	// Если правило повторения не указано, возвращаем ошибку
	if repeat == "" {
		return "", errors.New("правило повторения не указано")
	}

	// Разбираем базовые правила
	switch {
	case strings.HasPrefix(repeat, "d "): // Правило повторения через дни
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", errors.New("некорректное правило d, нужно указать количество дней")
		}

		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("некорректное количество дней для правила d")
		}

		// Добавляем дни к дате
		nextDate := startDate.AddDate(0, 0, days)
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}

		return nextDate.Format("20060102"), nil

	case repeat == "y": // Правило повторения ежегодно
		nextDate := startDate.AddDate(1, 0, 0)
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		return nextDate.Format("20060102"), nil

	case strings.HasPrefix(repeat, "w "): // Правило повторения по дням недели
		parts := strings.Split(repeat[2:], ",")
		weekDays := make(map[int]bool)
		for _, part := range parts {
			day, err := strconv.Atoi(part)
			if err != nil || day < 1 || day > 7 {
				return "", errors.New("некорректное значение дня недели для правила w")
			}
			weekDays[day] = true
		}

		// Идем день за днем, пока не найдем подходящий день недели
		nextDate := startDate.AddDate(0, 0, 1)
		for !nextDate.After(now) || !weekDays[int(nextDate.Weekday())+1] {
			nextDate = nextDate.AddDate(0, 0, 1)
		}

		return nextDate.Format("20060102"), nil

	case strings.HasPrefix(repeat, "m "): // Правило повторения по дням месяца
		parts := strings.Split(repeat[2:], " ")
		if len(parts) == 0 {
			return "", errors.New("некорректное правило m")
		}

		daysStr := strings.Split(parts[0], ",")
		monthDays := make(map[int]bool)
		for _, dayStr := range daysStr {
			day, err := strconv.Atoi(dayStr)
			if err != nil || day < -31 || day > 31 || day == 0 {
				return "", errors.New("некорректное значение дня месяца для правила m")
			}
			monthDays[day] = true
		}

		months := map[int]bool{}
		if len(parts) > 1 {
			monthStrs := strings.Split(parts[1], ",")
			for _, monthStr := range monthStrs {
				month, err := strconv.Atoi(monthStr)
				if err != nil || month < 1 || month > 12 {
					return "", errors.New("некорректное значение месяца для правила m")
				}
				months[month] = true
			}
		}

		nextDate := startDate.AddDate(0, 0, 1)
		for !nextDate.After(now) {
			day := nextDate.Day()
			month := int(nextDate.Month())

			// Проверяем соответствие дня и месяца правилам
			if monthDays[day] || (day < 0 && checkNegativeDay(nextDate, day)) {
				if len(months) == 0 || months[month] {
					return nextDate.Format("20060102"), nil
				}
			}

			nextDate = nextDate.AddDate(0, 0, 1)
		}

		return nextDate.Format("20060102"), nil

	default:
		return "", errors.New("неподдерживаемое или некорректное правило повторения")
	}
}

// checkNegativeDay проверяет, соответствует ли день конца месяца правилам
func checkNegativeDay(date time.Time, day int) bool {
	lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location()).Day()
	return lastDay+day+1 == date.Day()
}

// Обработчик для маршрута /api/nextdate
func handleNextDate(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "некорректная дата в параметре 'now'", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем просто строку даты без JSON-обертки
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}
