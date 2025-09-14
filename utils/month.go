package utils

import "time"

// nextMonth increments a YYYY-MM string by one month
func NextMonth(m string) string {
	t, _ := time.Parse("2006-01", m)
	t = t.AddDate(0, 1, 0)
	return t.Format("2006-01")
}
