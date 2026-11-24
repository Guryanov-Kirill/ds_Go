package stand

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"moviesvc/movie"
)

var (
	ErrNotFound    = errors.New("movie not found on stand")
	ErrTimeout     = errors.New("stand did not answer in time")
	ErrUnavailable = errors.New("stand is unavailable")
)

// таймаут на один запрос к стенду
const requestTimeout = 5 * time.Second

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{},
	}
}

// формат фильма на стенде
type standMovie struct {
	Num   int    `json:"num"`
	Title string `json:"title"`
	Img   string `json:"img"`
}

// GetMovie берет один фильм со стенда по id
func (c *Client) GetMovie(ctx context.Context, id int) (movie.Movie, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return movie.Movie{}, fmt.Errorf("%w (id %d)", ErrUnavailable, id)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return movie.Movie{}, fmt.Errorf("%w (id %d)", ErrTimeout, id)
		}
		if errors.Is(err, context.Canceled) {
			return movie.Movie{}, context.Canceled
		}
		return movie.Movie{}, fmt.Errorf("%w (id %d)", ErrUnavailable, id)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return movie.Movie{}, fmt.Errorf("%w (id %d)", ErrNotFound, id)
	}
	if resp.StatusCode != http.StatusOK {
		return movie.Movie{}, fmt.Errorf("%w: status %d (id %d)", ErrUnavailable, resp.StatusCode, id)
	}

	var raw standMovie
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return movie.Movie{}, fmt.Errorf("%w: bad response (id %d)", ErrUnavailable, id)
	}

	return movie.Movie{ID: raw.Num, Title: raw.Title, Img: raw.Img}, nil
}
