package handler

import "net/http"

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	print("HealthCheck")
	w.WriteHeader(http.StatusOK)
}
