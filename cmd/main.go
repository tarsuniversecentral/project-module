package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tarsuniversecentral/project-module/internal/config"
	handler "github.com/tarsuniversecentral/project-module/internal/handlers"
	models "github.com/tarsuniversecentral/project-module/internal/models"
	service "github.com/tarsuniversecentral/project-module/internal/services"

	"github.com/gorilla/mux"
)

func main() {

	db, err := config.InitDatabase()
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.Close()

	// Initialize services and handlers
	ProjectModel := models.NewProjectModel(db)
	projectService := service.NewProjectService(ProjectModel)
	projectHandler := handler.NewProjectHandler(projectService)

	// Set up router
	r := mux.NewRouter()

	// Project routes
	r.HandleFunc("/projects", projectHandler.CreateProject).Methods("POST")
	r.HandleFunc("/projects/{id}", projectHandler.GetProject).Methods("GET")
	// Add more routes as needed

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
