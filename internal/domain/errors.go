package domain

import "errors"

var (
	ErrInvalidRange  = errors.New("invalid range")
	ErrMovieNotFound = errors.New("movie not found")
	ErrNoMovies      = errors.New("no movies on stand")

	ErrStandTimeout     = errors.New("stand did not answer in time")
	ErrStandUnavailable = errors.New("stand is unavailable")
)
