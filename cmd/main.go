package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/tarsuniversecentral/project-module/internal/api"
	"github.com/tarsuniversecentral/project-module/internal/handlers"
	"github.com/tarsuniversecentral/project-module/internal/models"
	"github.com/tarsuniversecentral/project-module/internal/router"
	"github.com/tarsuniversecentral/project-module/internal/services"
	"github.com/tarsuniversecentral/project-module/pkg/database"
)

// Server wraps an http.Server instance.
type Server struct {
	httpServer *http.Server
}

// NewServer creates a new Server instance with the provided router.
func NewServer(router *mux.Router) *Server {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return &Server{httpServer: srv}
}

// Start runs the server and handles graceful shutdown on SIGINT/SIGTERM.
func (s *Server) Start() {
	// Start the server in a goroutine.
	go func() {
		log.Printf("Server running on %s\n", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not listen on %s: %v\n", s.httpServer.Addr, err)
		}
	}()

	// Listen for termination signals.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for the shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt graceful shutdown.
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}

func main() {
	// Initialize the database.
	db, err := database.InitDatabase()
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.Close()

	// Initialize models.
	projectModel := models.NewProjectModel(db)

	// Initialize services.
	projectService := services.NewProjectService(projectModel)

	// Initialize handlers.
	projectHandler := handlers.NewProjectHandler(projectService)

	// Create the composite API struct.
	apiComposite := api.NewAPI(projectHandler)

	// Set up the router with all routes.
	router := router.NewRouter(apiComposite)

	// Create and start the server.
	server := NewServer(router)
	server.Start()
}
