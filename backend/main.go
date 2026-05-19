package main

import (
	"log"
	"os"

	"github.com/Clivet-lug/todo-app/backend/databases"
	"github.com/Clivet-lug/todo-app/backend/handlers"
	"github.com/Clivet-lug/todo-app/backend/repositories"
	"github.com/Clivet-lug/todo-app/backend/server"
	"github.com/Clivet-lug/todo-app/backend/services"
	_ "github.com/lib/pq"
)

func main() {
	// Infrastructure
	db  := databases.ConnectDB()
	rdb := databases.ConnectRedis()
	defer db.Close()
	if rdb != nil {
		defer rdb.Close()
	}

	// Repositories (data layer)
	userRepo := repositories.NewUserRepository(db)
	todoRepo := repositories.NewTodoRepository(db)

	// Bootstrap tables 
	if err := userRepo.CreateUsersTable(); err != nil {
		log.Fatal("users table:", err)
	}
	if err := todoRepo.CreateTable(); err != nil {
		log.Fatal("todos table:", err)
	}

	// Services (business logic)
	authSvc := services.NewAuthService(userRepo)
	todoSvc := services.NewTodoService(todoRepo, userRepo)

	// Handlers (HTTP layer) 
	authHandler := handlers.NewAuthHandler(authSvc)
	todoHandler := handlers.NewTodoHandler(todoSvc)

	// Server 
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	srv := server.New(port, rdb, authHandler, todoHandler)
	if err := srv.Start(); err != nil {
		log.Fatal("server error:", err)
	}
}