package grpcserver

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	sv_grpc "github.com/baochammm/mangahub/internal/grpc"
	pb "github.com/baochammm/mangahub/internal/grpc/manga"
	"github.com/baochammm/mangahub/internal/manga"
)

func StartGRPCServer(repo manga.Repository, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterMangaServiceServer(
		grpcServer,
		sv_grpc.NewServer(repo),
	)

	log.Printf("📡 gRPC server listening on :%d\n", port)

	// BLOCKING call — caller should run this in a goroutine
	return grpcServer.Serve(lis)
}
