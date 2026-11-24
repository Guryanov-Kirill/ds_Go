package service

import (
	"context"
	"errors"
	"fmt"

	"moviesvc/movie"
	"moviesvc/stand"
)

var (
	ErrInvalidRange = errors.New("invalid range")
	ErrNoMovies     = errors.New("no movies on stand")
)

// если фильма с id=99999 нет, то с большим id их точно нет
const maxID = 99999

// фильм, который не удалось загрузить
type LoadError struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

type LoadResult struct {
	Movies []movie.Movie `json:"movies"`
	Errors []LoadError   `json:"errors"`
}

type Service struct {
	stand *stand.Client
}

func New(c *stand.Client) *Service {
	return &Service{stand: c}
}

// Load загружает фильмы с id от from до to
func (s *Service) Load(ctx context.Context, from, to int) (LoadResult, error) {
	if from < 1 || to < from {
		return LoadResult{}, fmt.Errorf("%w: need 1 <= from <= to", ErrInvalidRange)
	}

	// пустые слайсы, чтобы в JSON было [] а не null
	result := LoadResult{Movies: []movie.Movie{}, Errors: []LoadError{}}

	for id := from; id <= to; id++ {
		m, err := s.stand.GetMovie(ctx, id)
		if err != nil {
			// если запрос отменили, дальше не идем
			if ctx.Err() != nil {
				return LoadResult{}, ctx.Err()
			}
			result.Errors = append(result.Errors, LoadError{ID: id, Message: err.Error()})
			continue
		}
		result.Movies = append(result.Movies, m)
	}
	return result, nil
}

// Last ищет последний фильм бинарным поиском
func (s *Service) Last(ctx context.Context) (movie.Movie, error) {
	var last movie.Movie
	found := false

	lo, hi := 1, maxID
	for lo <= hi {
		mid := (lo + hi) / 2
		m, err := s.stand.GetMovie(ctx, mid)
		switch {
		case err == nil:
			// фильм есть, ищем правее
			last, found = m, true
			lo = mid + 1
		case errors.Is(err, stand.ErrNotFound):
			// фильма нет, ищем левее
			hi = mid - 1
		default:
			return movie.Movie{}, err
		}
	}

	if !found {
		return movie.Movie{}, ErrNoMovies
	}
	return last, nil
}
