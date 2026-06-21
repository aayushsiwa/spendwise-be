package services

import "errors"

var ErrImportValidation = errors.New("import validation error")

type recordRow struct {
	id          string
	date        any
	description string
	categoryID  any
	amount      float64
	recordType  string
	note        string
}

func errMissingColumn(col string) error {
	return errors.Join(ErrImportValidation, errors.New("missing required column: "+col))
}
