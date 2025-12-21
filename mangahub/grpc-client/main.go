package grpcclient

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/baochammm/mangahub/internal/grpc/manga"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func NewMangaClient() (pb.MangaServiceClient, func(), error) {
	conn, err := grpc.NewClient(
		"localhost:9092",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		_ = conn.Close()
	}

	return pb.NewMangaServiceClient(conn), cleanup, nil
}

func newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func StartGRPCClientServer() {
	// Connect to server
	conn, err := grpc.NewClient("localhost:9092",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

}

func GetMangaByID(mangaID string) error {
	client, cleanup, err := NewMangaClient()
	if err != nil {
		return err
	}
	defer cleanup()

	ctx, cancel := newContext()
	defer cancel()

	resp, err := client.GetManga(ctx, &pb.GetMangaRequest{
		MangaId: mangaID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return fmt.Errorf(st.Message())
	}

	fmt.Printf(
		"Manga ID: %s\nTitle: %s\nAuthor: %s\nDescription: %s\n",
		resp.Manga.Id,
		resp.Manga.Title,
		resp.Manga.Author,
		resp.Manga.Description,
	)

	return nil
}
func SearchManga(keyword string, page, pageSize int32) error {
	client, cleanup, err := NewMangaClient()
	if err != nil {
		return err
	}
	defer cleanup()

	ctx, cancel := newContext()
	defer cancel()

	if page < 1 || pageSize < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	resp, err := client.Search(ctx, &pb.SearchMangaRequest{
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return fmt.Errorf(st.Message())
	}

	for _, manga := range resp.Results {
		fmt.Printf(
			"Manga ID: %s | Title: %s | Author: %s\n",
			manga.Id,
			manga.Title,
			manga.Author,
		)
	}

	return nil
}
