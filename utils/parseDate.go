package utils

import (
	"errors"
	"strings"
	"time"
)

var dateFormats = []string{
	"2006-01-02",      // 2023-07-25
	"02-01-2006",      // 25-07-2023
	"02/01/2006",      // 25/07/2023
	"01/02/2006",      // 07/25/2023
	"1/2/2006",        // 7/25/2023
	"2 Jan 2006",      // 25 Jul 2023
	"2 Jan, 2006",     // 25 Jul 2023
	"2 January 2006",  // 25 July 2023
	"January 2, 2006", // July 25, 2023
	"Jan 2, 2006",     // Jul 25, 2023
	"2006/01/02",      // 2023/07/25
	"2006.01.02",      // 2023.07.25
	"02.01.2006",      // 25.07.2023
}

func ParseDate(dateStr string) (string, error) {
	dateStr = strings.TrimSpace(dateStr)

	for _, format := range dateFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}

	return "", errors.New("unrecognized date format: " + dateStr)
}
