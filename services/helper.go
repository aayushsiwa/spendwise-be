package services

type recordRow struct {
	id          string
	date        any
	description string
	categoryID  any
	amount      float64
	recordType  string
	note        string
}

type importError string

func (e importError) Error() string { return string(e) }

func errMissingColumn(col string) error {
	return importError("missing required column: " + col)
}
