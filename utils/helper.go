package utils

import "strings"

func Abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func Normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}

func InferRecordType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "dr", "debit", "debitcard", "expense", "payment", "withdrawal", "sent", "debit( dr )":
		return "expense"
	case "cr", "credit", "creditcard", "income", "deposit", "refund", "received", "credit( cr )":
		return "income"
	case "transfer":
		return "transfer"
	}
	return ""
}
