package handlers

import "aayushsiwa/expense-tracker/services"

type Handler struct {
	Service services.Service
}

func NewHandler(s services.Service) *Handler {
	return &Handler{Service: s}
}
