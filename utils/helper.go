package utils

import "strings"

// Abs returns the absolute value of f.
func Abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// Normalize returns a canonical form of the input string, converted to lowercase with whitespace and punctuation removed.
func Normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}

// InferRecordType infers the normalized transaction category from raw input. It returns "expense" for debit and expense-related terms, "income" for credit and income-related terms, "transfer" for transfer inputs, and an empty string for unrecognized inputs.
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
