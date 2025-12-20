package main

import (
	"log"

	udpserver "github.com/baochammm/mangahub/cmd/udp-server"
	"github.com/baochammm/mangahub/internal/api/middleware"
	"github.com/baochammm/mangahub/internal/api/routes"
	"github.com/baochammm/mangahub/internal/manga"
	"github.com/baochammm/mangahub/internal/udp"
	"github.com/baochammm/mangahub/internal/websocket"
	"github.com/baochammm/mangahub/package/database"
	"github.com/joho/godotenv"

	grpcserver "github.com/baochammm/mangahub/cmd/grpc-server"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	database.InitSQLite("./data/mangahub.db")
	defer database.Close()
	//REST API Router
	r := gin.Default()

	// //WebSocket Server
	hub := websocket.NewChatHub()
	go hub.Run()
	//UDP Server
	udpRepo := udp.NewUDPRepository(database.DB)
	udpHandler := udp.NewUDPHandler(udpRepo)

	go udpserver.StartUDPServer(udpHandler)
	//gRPC Server
	grpcRepo := *manga.NewRepository(database.DB)
	go func() {
		if err := grpcserver.StartGRPCServer(grpcRepo, 9092); err != nil {
			log.Fatal("gRPC server failed:", err)
		}
	}()
	// API Routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Welcome to MangaHub API! Available endpoints:  /manga [GET] - List all mangas,  /users [POST] - Create a new user,  /users/:user_id/reading-list [POST] - Add a reading entry for a user, /users/:user_id [GET] - Get user details with reading lists"})
	})
	r.GET("/ws/chat", middleware.AuthMiddleware(), websocket.ServeWS(hub))
	routes.RegisterUnProtectedRoutes(r)
	routes.RegisterProtectedRoutes(r)
	if err := r.Run(":8080"); err != nil {
		log.Fatal("server failed:", err)
	}

}
