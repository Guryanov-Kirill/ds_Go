package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"moviesvc/internal/domain"
	"moviesvc/internal/usecase"
)

type HTTPHandler struct {
	uc usecase.MovieUseCase
}

func NewHTTP(uc usecase.MovieUseCase) http.Handler {
	h := &HTTPHandler{uc: uc}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /load-movies", h.loadMovies)
	mux.HandleFunc("GET /last-movie", h.lastMovie)
	mux.HandleFunc("GET /swagger/index.html", serveSwagger)
	mux.HandleFunc("GET /swagger/openapi.json", serveOpenAPI)
	return mux
}

type movieResponse struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

func toMovieResponse(m domain.Movie) movieResponse {
	return movieResponse{ID: m.ID, Title: m.Title, Year: m.Year, Director: m.Director}
}

func (h *HTTPHandler) loadMovies(w http.ResponseWriter, r *http.Request) {
	from, err1 := strconv.Atoi(r.URL.Query().Get("from"))
	to, err2 := strconv.Atoi(r.URL.Query().Get("to"))
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "invalid_argument", "from and to must be integers")
		return
	}

	result, err := h.uc.LoadMovies(r.Context(), from, to)
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	if len(result.Failed) > 0 {
		ids := make([]string, 0, len(result.Failed))
		for _, f := range result.Failed {
			ids = append(ids, strconv.Itoa(f.ID))
		}
		w.Header().Set("X-Failed-IDs", strings.Join(ids, ","))
	}

	movies := make([]movieResponse, 0, len(result.Movies))
	for _, m := range result.Movies {
		movies = append(movies, toMovieResponse(m))
	}
	writeJSON(w, http.StatusOK, movies)
}

func (h *HTTPHandler) lastMovie(w http.ResponseWriter, r *http.Request) {
	m, err := h.uc.GetLastMovie(r.Context())
	if err != nil {
		writeUseCaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toMovieResponse(m))
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

func writeUseCaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRange):
		writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
	case errors.Is(err, domain.ErrNoMovies), errors.Is(err, domain.ErrMovieNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, domain.ErrStandTimeout):
		writeError(w, http.StatusGatewayTimeout, "stand_timeout", err.Error())
	case errors.Is(err, domain.ErrStandUnavailable):
		writeError(w, http.StatusBadGateway, "stand_error", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}
