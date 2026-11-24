package main

import (
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"moviesvc/config"
	"moviesvc/grpcapi"
	"moviesvc/httpapi"
	"moviesvc/pb"
	"moviesvc/service"
	"moviesvc/stand"
)

func main() {
	cfg := config.Load()
	svc := service.New(stand.New(cfg.MovieBaseURL))

	// gRPC сервер в отдельной горутине
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMovieServiceServer(grpcServer, grpcapi.New(svc))
	go func() {
		log.Fatal(grpcServer.Serve(lis))
	}()

	// HTTP сервер
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, httpapi.New(svc)))
}
