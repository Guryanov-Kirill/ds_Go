package usecase

import (
	"context"
	"errors"
	"fmt"

	"moviesvc/internal/domain"
)

type MovieProvider interface {
	GetMovie(ctx context.Context, id int) (domain.Movie, error)
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, from, to int) (domain.LoadResult, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
}

// если фильма с id=99999 нет, то с большим id их точно нет
const maxID = 99999

type Service struct {
	provider MovieProvider
}

func New(provider MovieProvider) *Service {
	return &Service{provider: provider}
}

// LoadMovies загружает фильмы с id от from до to
func (s *Service) LoadMovies(ctx context.Context, from, to int) (domain.LoadResult, error) {
	if from < 1 || to < from {
		return domain.LoadResult{}, fmt.Errorf("%w: need 1 <= from <= to", domain.ErrInvalidRange)
	}

	// пустые слайсы, чтобы в JSON было [] а не null
	result := domain.LoadResult{Movies: []domain.Movie{}, Failed: []domain.FailedMovie{}}

	for id := from; id <= to; id++ {
		m, err := s.provider.GetMovie(ctx, id)
		if err != nil {
			// если запрос отменили, дальше не идем
			if ctx.Err() != nil {
				return domain.LoadResult{}, ctx.Err()
			}
			result.Failed = append(result.Failed, domain.FailedMovie{ID: id, Message: err.Error()})
			continue
		}
		result.Movies = append(result.Movies, m)
	}
	return result, nil
}

// GetLastMovie ищет последний фильм бинарным поиском
func (s *Service) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	var last domain.Movie
	found := false

	lo, hi := 1, maxID
	for lo <= hi {
		mid := (lo + hi) / 2
		m, err := s.provider.GetMovie(ctx, mid)
		switch {
		case err == nil:
			// фильм есть, ищем правее
			last, found = m, true
			lo = mid + 1
		case errors.Is(err, domain.ErrMovieNotFound):
			// фильма нет, ищем левее
			hi = mid - 1
		default:
			return domain.Movie{}, err
		}
	}

	if !found {
		return domain.Movie{}, domain.ErrNoMovies
	}
	return last, nil
}
