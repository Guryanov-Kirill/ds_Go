package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"moviesvc/service"
	"moviesvc/stand"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) http.Handler {
	h := &Handler{svc: svc}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /load-movies", h.loadMovies)
	mux.HandleFunc("GET /last-movie", h.lastMovie)
	mux.HandleFunc("GET /swagger/index.html", serveSwagger)
	mux.HandleFunc("GET /swagger/openapi.json", serveOpenAPI)
	return mux
}

func (h *Handler) loadMovies(w http.ResponseWriter, r *http.Request) {
	from, err1 := strconv.Atoi(r.URL.Query().Get("from"))
	to, err2 := strconv.Atoi(r.URL.Query().Get("to"))
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "invalid_argument", "from and to must be integers")
		return
	}

	result, err := h.svc.Load(r.Context(), from, to)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) lastMovie(w http.ResponseWriter, r *http.Request) {
	m, err := h.svc.Last(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}

// writeServiceError переводит ошибку сервиса в HTTP статус
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidRange):
		writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
	case errors.Is(err, service.ErrNoMovies), errors.Is(err, stand.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, stand.ErrTimeout):
		writeError(w, http.StatusGatewayTimeout, "stand_timeout", err.Error())
	case errors.Is(err, stand.ErrUnavailable):
		writeError(w, http.StatusBadGateway, "stand_error", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}
