package handlers

import (
	"github.com/lithammer/shortuuid/v4"
)

func (h *Handler) GenerateCustomID(date string) (string, error) {
	return shortuuid.New(), nil
}
