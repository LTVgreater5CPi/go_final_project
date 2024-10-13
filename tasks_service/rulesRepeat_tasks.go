package tasks_service

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now, err := time.Parse(dateFormat, nowStr)
	if err != nil {
		http.Error(w, "неверный формат даты", http.StatusBadRequest)
		return
	}
	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, nextDate)
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("правило повторения пустое")
	}
	taskDate, err := time.Parse(dateFormat, date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %v", err)
	}
	startDate := now
	if taskDate.After(now) {
		startDate = taskDate
	}

	switch {
	case strings.HasPrefix(repeat, "d "): //ежедневное повторение
		daysStr := strings.TrimSpace(repeat[2:])
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 || days > 400 {
			return "", fmt.Errorf("неверное правило повторения 'd': %v", err)
		}
		taskDate = taskDate.AddDate(0, 0, days)

		for taskDate.Before(now) {
			taskDate = taskDate.AddDate(0, 0, days)
		}
		return taskDate.Format(dateFormat), nil

	case repeat == "y": // ежегодное повторение
		for !taskDate.After(startDate) {
			year := taskDate.Year() + 1
			month := taskDate.Month()
			day := taskDate.Day()

			if month == time.February && day == 29 && !isLeapYear(year) {
				taskDate = time.Date(year, time.March, 1, 0, 0, 0, 0, taskDate.Location())
			} else {
				taskDate = time.Date(year, month, day, 0, 0, 0, 0, taskDate.Location())
			}
		}

		return taskDate.Format(dateFormat), nil

	case strings.HasPrefix(repeat, "w "): //еженедельное повторение
		daysStr := strings.TrimSpace(repeat[2:])
		days := strings.Split(daysStr, ",")
		if len(days) == 0 {
			return "", fmt.Errorf("дни не указаны")
		}

		var daysOfWeek []int
		for _, dayStr := range days {
			day, err := strconv.Atoi(dayStr)
			if err != nil || day < 1 || day > 7 {
				return "", fmt.Errorf("неверный день '%s'", dayStr)
			}
			if day == 7 {
				day = 0
			}
			daysOfWeek = append(daysOfWeek, day)
		}
		sort.Ints(daysOfWeek)

		startDate = taskDate
		if now.After(taskDate) {
			startDate = now
		}
		initialDate := taskDate

		for !containsInt(daysOfWeek, int(startDate.Weekday())) || !(startDate.YearDay() > initialDate.YearDay()) {
			startDate = startDate.AddDate(0, 0, 1)
		}
		return startDate.Format(dateFormat), nil

	case strings.HasPrefix(repeat, "m "): // ежемесячное повторение
		parts := strings.Split(strings.TrimSpace(repeat[2:]), " ")

		if len(parts) == 0 {
			return "", fmt.Errorf("дни не указаны")
		}
		dayParts := strings.Split(parts[0], ",")
		var daysOfMonth []int
		for _, dayStr := range dayParts {
			day, err := strconv.Atoi(dayStr)
			if err != nil || day == 0 || day < -2 || day > 31 {
				return "", fmt.Errorf("неверный день '%s'", dayStr)
			}
			daysOfMonth = append(daysOfMonth, day)
		}

		var months []int
		if len(parts) > 1 {
			monthParts := strings.Split(parts[1], ",")
			for _, monthStr := range monthParts {
				month, err := strconv.Atoi(monthStr)
				if err != nil || month < 1 || month > 12 {
					return "", fmt.Errorf("неверный месяц '%s'", monthStr)
				}
				months = append(months, month)
			}
		}
		sort.Ints(daysOfMonth)
		sort.Ints(months)

		for {
			curYear, curMonth := taskDate.Year(), taskDate.Month()

			if len(months) > 0 && !containsInt(months, int(curMonth)) {
				nextMonth := findNextMonth(int(curMonth), months)
				if nextMonth <= int(curMonth) {
					taskDate = time.Date(curYear+1, time.Month(nextMonth), 1, 0, 0, 0, 0, taskDate.Location())
				} else {
					taskDate = time.Date(curYear, time.Month(nextMonth), 1, 0, 0, 0, 0, taskDate.Location())
				}
				continue
			}
			// Находим ближайшую допустимую дату в текущем месяце
			var validDate time.Time
			for _, day := range daysOfMonth {
				var candDate time.Time
				lastDay := lastDayOfMonth(curYear, curMonth)

				if day > 0 {
					if day <= lastDay {
						candDate = time.Date(curYear, curMonth, day, 0, 0, 0, 0, taskDate.Location())
					} else {
						continue
					}
				} else {
					if -day <= lastDay {
						candDate = time.Date(curYear, curMonth, lastDay+day+1, 0, 0, 0, 0, taskDate.Location())
					} else {
						continue
					}
				}
				// Находим ближайшую дату после startDate
				if candDate.After(startDate) && (validDate.IsZero() || candDate.Before(validDate)) {
					validDate = candDate
				}
			}
			if !validDate.IsZero() {
				return validDate.Format(dateFormat), nil
			}
			// Если не найдено допустимой даты в текущем месяце, переходим к следующему месяцу
			taskDate = taskDate.AddDate(0, 1, 0)
			taskDate = time.Date(taskDate.Year(), taskDate.Month(), 1, 0, 0, 0, 0, taskDate.Location())
		}
	default:
		return "", fmt.Errorf("неподдерживаемое правило: '%s'", repeat)
	}
}
