package main

import (
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"moviesvc/internal/adapter"
	"moviesvc/internal/config"
	"moviesvc/internal/handler"
	"moviesvc/internal/usecase"
	"moviesvc/pb"
)

func main() {
	cfg := config.Load()
	stand := adapter.NewStandClient(cfg.MovieBaseURL)
	uc := usecase.New(stand)
	// gRPC сервер в отдельной горутине
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMovieServiceServer(grpcServer, handler.NewGRPC(uc))
	go func() {
		log.Fatal(grpcServer.Serve(lis))
	}()

	// HTTP сервер
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, handler.NewHTTP(uc)))
}
