package services

import "errors"

var ErrImportValidation = errors.New("import validation error")

// errMissingColumn returns an error indicating that the specified import column is missing.
func errMissingColumn(col string) error {
	return errors.Join(ErrImportValidation, errors.New("missing required column: "+col))
}
