package utils

import "testing"

func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		want float64
	}{
		{name: "positive number unchanged", f: 42.5, want: 42.5},
		{name: "negative number becomes positive", f: -42.5, want: 42.5},
		{name: "zero unchanged", f: 0.0, want: 0.0},
		{name: "small positive", f: 0.001, want: 0.001},
		{name: "small negative", f: -0.001, want: 0.001},
		{name: "large negative", f: -1_000_000.99, want: 1_000_000.99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Abs(tt.f)
			if got != tt.want {
				t.Errorf("Abs(%v) = %v, want %v", tt.f, got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{name: "already lowercase no special chars", s: "hello", want: "hello"},
		{name: "uppercase is lowercased", s: "Hello", want: "hello"},
		{name: "leading and trailing spaces trimmed", s: "  hello  ", want: "hello"},
		{name: "underscores removed", s: "hello_world", want: "helloworld"},
		{name: "spaces removed", s: "hello world", want: "helloworld"},
		{name: "slashes removed", s: "hello/world", want: "helloworld"},
		{name: "dashes removed", s: "hello-world", want: "helloworld"},
		{name: "mixed special characters", s: "  Hello_World/Foo-Bar  ", want: "helloworldfoobar"},
		{name: "empty string", s: "", want: ""},
		{name: "all spaces", s: "   ", want: ""},
		{name: "transaction date header", s: "Transaction Date", want: "transactiondate"},
		{name: "dr/cr header", s: "DR/CR", want: "drcr"},
		{name: "chq/refno header", s: "CHQ/RefNo", want: "chqrefno"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.s)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestInferRecordType(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		// Expense keywords
		{name: "dr -> expense", raw: "dr", want: "expense"},
		{name: "DR uppercase -> expense", raw: "DR", want: "expense"},
		{name: "debit -> expense", raw: "debit", want: "expense"},
		{name: "debitcard -> expense", raw: "debitcard", want: "expense"},
		{name: "expense -> expense", raw: "expense", want: "expense"},
		{name: "payment -> expense", raw: "payment", want: "expense"},
		{name: "withdrawal -> expense", raw: "withdrawal", want: "expense"},
		{name: "sent -> expense", raw: "sent", want: "expense"},
		{name: "debit( dr ) -> expense", raw: "debit( dr )", want: "expense"},

		// Income keywords
		{name: "cr -> income", raw: "cr", want: "income"},
		{name: "CR uppercase -> income", raw: "CR", want: "income"},
		{name: "credit -> income", raw: "credit", want: "income"},
		{name: "creditcard -> income", raw: "creditcard", want: "income"},
		{name: "income -> income", raw: "income", want: "income"},
		{name: "deposit -> income", raw: "deposit", want: "income"},
		{name: "refund -> income", raw: "refund", want: "income"},
		{name: "received -> income", raw: "received", want: "income"},
		{name: "credit( cr ) -> income", raw: "credit( cr )", want: "income"},

		// Transfer
		{name: "transfer -> transfer", raw: "transfer", want: "transfer"},
		{name: "TRANSFER uppercase -> transfer", raw: "TRANSFER", want: "transfer"},

		// Unknown / empty
		{name: "unknown value returns empty", raw: "unknown", want: ""},
		{name: "empty string returns empty", raw: "", want: ""},
		{name: "whitespace returns empty", raw: "   ", want: ""},

		// Leading/trailing whitespace is trimmed before comparison
		{name: "leading space dr", raw: " dr", want: "expense"},
		{name: "trailing space cr", raw: "cr ", want: "income"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferRecordType(tt.raw)
			if got != tt.want {
				t.Errorf("InferRecordType(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
