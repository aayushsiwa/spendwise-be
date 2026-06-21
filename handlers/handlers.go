package handlers

import "aayushsiwa/expense-tracker/services"

type Handler struct {
	Service services.Service
}

// NewHandler returns a new Handler configured with the provided service.
func NewHandler(s services.Service) *Handler {
	return &Handler{Service: s}
}
