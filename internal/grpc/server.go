package sv_grpc

import (
	"context"
	"fmt"

	pb "github.com/baochammm/mangahub/internal/grpc/manga"
	manga "github.com/baochammm/mangahub/internal/manga"
)

type Server struct {
	pb.UnimplementedMangaServiceServer
	repo manga.Repository
}

func NewServer(repo manga.Repository) *Server {
	return &Server{repo: repo}
}

/* ========== UC-014 ========== */

func (s *Server) GetManga(
	ctx context.Context,
	req *pb.GetMangaRequest,
) (*pb.GetMangaResponse, error) {

	manga, err := s.repo.GetByID(req.MangaId)
	if err != nil {
		return nil, fmt.Errorf("manga not found: %w", err)
	}

	return &pb.GetMangaResponse{
		Manga: &pb.Manga{
			Id:          manga.ID,
			Title:       manga.Title,
			Author:      manga.Author,
			Description: manga.Description,
		},
	}, nil
}

/* ========== UC-015 ========== */

func (s *Server) Search(
	ctx context.Context,
	req *pb.SearchMangaRequest,
) (*pb.SearchMangaResponse, error) {

	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize

	rows, err := s.repo.DB.Query(`
		SELECT id, title, author, description
		FROM mangas
		WHERE LOWER(title) LIKE '%' || LOWER(?) || '%'
		LIMIT ? OFFSET ?
	`, req.Keyword, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*pb.Manga, 0)

	for rows.Next() {
		m := &pb.Manga{}
		if err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Author,
			&m.Description,
		); err != nil {
			return nil, err
		}
		results = append(results, m)
	}

	var total int64
	_ = s.repo.DB.QueryRow(`
		SELECT COUNT(*)
		FROM mangas
		WHERE LOWER(title) LIKE '%' || LOWER(?) || '%'
	`, req.Keyword).Scan(&total)

	return &pb.SearchMangaResponse{
		Results: results,
		Total:   total,
	}, nil

}

/* ========== UC-016 ========== */
//ToDo after tcp
// func (s *Server) UpdateProgress(
// 	ctx context.Context,
// 	req *pb.UpdateProgressRequest,
// ) (*pb.UpdateProgressResponse, error) {

// 	if req.UserId == 0 || req.MangaId == "" {
// 		return nil, fmt.Errorf("invalid request")
// 	}

// 	err := s.repo.UpdateProgress(
// 		req.UserId,
// 		req.MangaId,
// 		req.Chapter,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Trigger TCP broadcast (real-time sync)
// 	// go BroadcastProgressUpdate(
// 	// 	req.UserId,
// 	// 	req.MangaId,
// 	// 	req.Chapter,
// 	// )

// 	return &pb.UpdateProgressResponse{
// 		Success: true,
// 	}, nil
// }
