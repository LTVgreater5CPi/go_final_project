package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func slicecontains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func findNextMonth(currentMonth int, months []int) int {
	for _, month := range months {
		if month >= currentMonth {
			return month
		}
	}
	return months[0]
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// Обработка правил переноса
func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repetition rule is empty")
	}
	taskDate, err := time.Parse(DateFormat, date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}
	startDate := now
	if taskDate.After(now) {
		startDate = taskDate
	}

	switch {
	case strings.HasPrefix(repeat, "d "):
		daysStr := strings.TrimSpace(repeat[2:])
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 || days > 400 {
			return "", fmt.Errorf("invalid repetition rule 'd': %v", err)
		}
		taskDate = taskDate.AddDate(0, 0, days)
		for taskDate.Before(now) {
			taskDate = taskDate.AddDate(0, 0, days)
		}
		return taskDate.Format(DateFormat), nil

	case repeat == "y":
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
		return taskDate.Format(DateFormat), nil

	case strings.HasPrefix(repeat, "w "):
		daysStr := strings.TrimSpace(repeat[2:])
		days := strings.Split(daysStr, ",")
		if len(days) == 0 {
			return "", fmt.Errorf("invalid repetition rule 'w': days not specified")
		}

		var daysOfWeek []int
		for _, dayStr := range days {
			day, err := strconv.Atoi(dayStr)
			if err != nil || day < 1 || day > 7 {
				return "", fmt.Errorf("invalid repetition rule 'w': invalid day '%s'", dayStr)
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

		for !slicecontains(daysOfWeek, int(startDate.Weekday())) || !(startDate.YearDay() > initialDate.YearDay()) {
			startDate = startDate.AddDate(0, 0, 1)
		}
		return startDate.Format(DateFormat), nil

	case strings.HasPrefix(repeat, "m "):
		parts := strings.Split(strings.TrimSpace(repeat[2:]), " ")

		if len(parts) == 0 {
			return "", fmt.Errorf("invalid repetition rule 'm': days not specified")
		}
		dayParts := strings.Split(parts[0], ",")
		var daysOfMonth []int
		for _, dayStr := range dayParts {
			day, err := strconv.Atoi(dayStr)
			if err != nil || day == 0 || day < -2 || day > 31 {
				return "", fmt.Errorf("invalid repetition rule 'm': invalid day '%s'", dayStr)
			}
			daysOfMonth = append(daysOfMonth, day)
		}

		var months []int
		if len(parts) > 1 {
			monthParts := strings.Split(parts[1], ",")
			for _, monthStr := range monthParts {
				month, err := strconv.Atoi(monthStr)
				if err != nil || month < 1 || month > 12 {
					return "", fmt.Errorf("invalid repetition rule 'm': invalid month '%s'", monthStr)
				}
				months = append(months, month)
			}
		}
		sort.Ints(daysOfMonth)
		sort.Ints(months)

		for {
			currentYear, currentMonth := taskDate.Year(), taskDate.Month()

			if len(months) > 0 && !slicecontains(months, int(currentMonth)) {
				nextMonth := findNextMonth(int(currentMonth), months)
				if nextMonth <= int(currentMonth) {
					taskDate = time.Date(currentYear+1, time.Month(nextMonth), 1, 0, 0, 0, 0, taskDate.Location())
				} else {
					taskDate = time.Date(currentYear, time.Month(nextMonth), 1, 0, 0, 0, 0, taskDate.Location())
				}
				continue
			}

			var needDate time.Time
			for _, day := range daysOfMonth {
				var candDate time.Time
				lastDay := lastDayOfMonth(currentYear, currentMonth)

				if day > 0 {
					if day <= lastDay {
						candDate = time.Date(currentYear, currentMonth, day, 0, 0, 0, 0, taskDate.Location())
					} else {
						continue
					}
				} else {
					if -day <= lastDay {
						candDate = time.Date(currentYear, currentMonth, lastDay+day+1, 0, 0, 0, 0, taskDate.Location())
					} else {
						continue
					}
				}
				if candDate.After(startDate) && (needDate.IsZero() || candDate.Before(needDate)) {
					needDate = candDate
				}
			}
			if !needDate.IsZero() {
				return needDate.Format(DateFormat), nil
			}
			taskDate = taskDate.AddDate(0, 1, 0)
			taskDate = time.Date(taskDate.Year(), taskDate.Month(), 1, 0, 0, 0, 0, taskDate.Location())
		}
	default:
		return "", fmt.Errorf("unsupported repetition rule format: '%s'", repeat)
	}
}
