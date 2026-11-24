package grpcapi

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"moviesvc/movie"
	"moviesvc/pb"
	"moviesvc/service"
	"moviesvc/stand"
)

type Server struct {
	pb.UnimplementedMovieServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) LoadMovies(ctx context.Context, req *pb.LoadMoviesRequest) (*pb.LoadMoviesResponse, error) {
	result, err := s.svc.Load(ctx, int(req.GetFrom()), int(req.GetTo()))
	if err != nil {
		return nil, toStatus(err)
	}

	resp := &pb.LoadMoviesResponse{}
	for _, m := range result.Movies {
		resp.Movies = append(resp.Movies, toPB(m))
	}
	for _, e := range result.Errors {
		resp.Errors = append(resp.Errors, &pb.LoadError{Id: int32(e.ID), Message: e.Message})
	}
	return resp, nil
}

func (s *Server) GetLastMovie(ctx context.Context, _ *pb.GetLastMovieRequest) (*pb.Movie, error) {
	m, err := s.svc.Last(ctx)
	if err != nil {
		return nil, toStatus(err)
	}
	return toPB(m), nil
}

func toPB(m movie.Movie) *pb.Movie {
	return &pb.Movie{Id: int32(m.ID), Title: m.Title, Img: m.Img}
}

// toStatus переводит ошибку сервиса в gRPC код
func toStatus(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidRange):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrNoMovies), errors.Is(err, stand.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, stand.ErrTimeout), errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, stand.ErrUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
