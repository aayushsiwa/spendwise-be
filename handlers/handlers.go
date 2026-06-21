package handlers

import (
	"aayushsiwa/expense-tracker/services"
	"mime/multipart"

	"github.com/lithammer/shortuuid/v4"
)

type Handler struct {
	Service services.Service
}

func NewHandler(s services.Service) *Handler {
	return &Handler{
		Service: s,
	}
}

var (
	genIDFunc    = func(date string) (string, error) { return shortuuid.New(), nil }
	openFileFunc = func(fh *multipart.FileHeader) (multipart.File, error) { return fh.Open() }
)
