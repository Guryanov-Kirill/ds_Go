package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"moviesvc/internal/domain"
)

// таймаут на один запрос к стенду
const requestTimeout = 5 * time.Second

type StandClient struct {
	baseURL string
	http    *http.Client
}

func NewStandClient(baseURL string) *StandClient {
	return &StandClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{},
	}
}

// формат фильма на стенде
type standMovie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

// GetMovie берет один фильм со стенда и превращает его в domain.Movie
func (c *StandClient) GetMovie(ctx context.Context, id int) (domain.Movie, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.Movie{}, fmt.Errorf("%w (id %d)", domain.ErrStandUnavailable, id)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return domain.Movie{}, fmt.Errorf("%w (id %d)", domain.ErrStandTimeout, id)
		}
		if errors.Is(err, context.Canceled) {
			return domain.Movie{}, context.Canceled
		}
		return domain.Movie{}, fmt.Errorf("%w (id %d)", domain.ErrStandUnavailable, id)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.Movie{}, fmt.Errorf("%w (id %d)", domain.ErrMovieNotFound, id)
	}
	if resp.StatusCode != http.StatusOK {
		return domain.Movie{}, fmt.Errorf("%w: status %d (id %d)", domain.ErrStandUnavailable, resp.StatusCode, id)
	}

	var raw standMovie
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return domain.Movie{}, fmt.Errorf("%w: bad response (id %d)", domain.ErrStandUnavailable, id)
	}

	return domain.Movie{ID: raw.ID, Title: raw.Title, Year: raw.Year, Director: raw.Director}, nil
}
