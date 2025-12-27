package main

import (
	"log"
	"os"

	udpserver "github.com/baochammm/mangahub/cmd/udp-server"
	"github.com/baochammm/mangahub/internal/api/middleware"
	"github.com/baochammm/mangahub/internal/api/routes"
	"github.com/baochammm/mangahub/internal/manga"
	"github.com/baochammm/mangahub/internal/tcp"
	"github.com/baochammm/mangahub/internal/udp"
	"github.com/baochammm/mangahub/internal/user"
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

	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/mangahub.db"
	}

	database.InitSQLite(dbPath)
	defer database.Close()
	//REST API Router
	r := gin.Default()

	// CORS middleware - Allow frontend to access API from different origin
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	//UDP Server
	udpRepo := udp.NewUDPRepository(database.DB)
	udpHandler := udp.NewUDPHandler(udpRepo)

	go udpserver.StartUDPServer(udpHandler)

	// //WebSocket Server
	hub := websocket.NewChatHub()
	hub.ChatRepo = websocket.NewRepository(database.DB)
	go hub.Run()
	// API Routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Welcome to MangaHub API! Available endpoints:  /manga [GET] - List all mangas,  /users [POST] - Create a new user,  /users/:user_id/reading-list [POST] - Add a reading entry for a user, /users/:user_id [GET] - Get user details with reading lists"})
	})
	r.GET("/ws/chat", middleware.AuthMiddleware(), websocket.ServeWS(hub))
	routes.RegisterUnProtectedRoutes(r)
	routes.RegisterProtectedRoutes(r)
	routes.RegisterAdminRoutes(r)
	//TCP Server
	// TCP hub will be set in main.go
	tcpHub := tcp.NewHub()
	userRepo := user.NewRepository(database.DB)
	userHandler := user.NewHandler(userRepo, tcpHub)
	r.PATCH("/users/progress", middleware.AuthMiddleware(), userHandler.UpdateReadingProgress) // mangahub progress update --manga-id <id> --current-chapter <chapter>
	go func() {
		if err := tcp.StartTCPServer(tcpHub, 9090); err != nil {
			log.Fatal(err)
		}
	}()
	//gRPC Server
	grpcRepo := *manga.NewRepository(database.DB)
	go func() {
		if err := grpcserver.StartGRPCServer(grpcRepo, tcpHub, 9092); err != nil {
			log.Fatal("gRPC server failed:", err)
		}
	}()
	if err := r.Run(":8080"); err != nil {
		log.Fatal("server failed:", err)
	}

}
