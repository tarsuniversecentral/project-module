package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/exp/rand"

	"github.com/tarsuniversecentral/project-module/internal/dto"
	service "github.com/tarsuniversecentral/project-module/internal/services"
)

type ProjectHandler struct {
	projectService *service.ProjectService
	fileService    *service.FileService
}

func NewProjectHandler(service *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: service}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {

	// Set a memory threshold of 10 MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Extracting form values
	project := dto.Project{
		Title:       r.FormValue("title"),
		Subtitle:    r.FormValue("subtitle"),
		Industry:    r.FormValue("industry"),
		Description: r.FormValue("description"),
		GithubLink:  r.FormValue("github_link"),
	}

	if val := r.FormValue("project_value"); val != "" {
		parsedValue, err := strconv.ParseFloat(val, 64)
		if err != nil {
			http.Error(w, "Invalid project_value format", http.StatusBadRequest)
			return
		}
		project.ProjectValue = parsedValue
	}

	project.LookingFor = r.Form["looking_for"]

	if err := dto.ValidateLookingFor(project.LookingFor); err != nil {
		http.Error(w, "Error validate looking_for: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve file headers for PDFs and images.
	pdfHeaders := r.MultipartForm.File["pdfs"]
	imageHeaders := r.MultipartForm.File["images"]

	// Process the file uploads concurrently in the service layer.
	fileResponse, err := h.fileService.ProcessUploads(pdfHeaders, imageHeaders)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	project.PitchDecks = fileResponse.PDFFiles
	project.Images = fileResponse.ImageFiles

	resProject, err := h.projectService.CreateProject(project)
	if err != nil {
		delErr := h.fileService.DeleteSavedFiles(dto.ConstructFileResults(fileResponse))
		if delErr != nil {
			combinedError := fmt.Errorf("project creation error: %v; file deletion error: %v", err, delErr)
			log.Printf("Internal server error: %v", combinedError)
			http.Error(w, combinedError.Error(), http.StatusInternalServerError)
			return

		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resProject)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.projectService.GetProject(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	project.LikeCount = rand.Intn(100)
	project.CommentCount = rand.Intn(45)
	project.ViewCount = rand.Intn(1000)
	project.Verified = rand.Intn(2) == 1

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) FileRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	file, err := h.fileService.RetrieveFile(filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving file: %v", err), http.StatusNotFound)
		return
	}
	defer file.Close()

	ext := filepath.Ext(filename)
	var contentType string
	switch strings.ToLower(ext) {
	case ".pdf":
		contentType = "application/pdf"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".svg":
		contentType = "image/svg+xml"
	default:
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, fmt.Sprintf("Error sending file: %v", err), http.StatusInternalServerError)
	}
}

func (h *ProjectHandler) AddTeamMemberToProject(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	projectIdStr := vars["projectId"]
	projectID, err := strconv.Atoi(projectIdStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var member dto.TeamMember
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	// Set the project ID from the URL, ensuring consistency.
	member.ProjectID = projectID

	// Insert the team member into the database.
	if err := h.projectService.AddTeamMember(&member); err != nil {
		http.Error(w, "Failed to insert team member", http.StatusInternalServerError)
		return
	}

	// Return the inserted team member as a JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(member); err != nil {
		log.Println("Failed to write response:", err)
	}
}

func (h *ProjectHandler) GetTeamMembersOfProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectIdStr := vars["projectId"]
	projectID, err := strconv.Atoi(projectIdStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Retrieve the team members from the database.
	members, err := h.projectService.GetTeamMembers(projectID)
	if err != nil {
		http.Error(w, "Failed to fetch team members", http.StatusInternalServerError)
		return
	}

	// Return the team members as a JSON response.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(members); err != nil {
		log.Println("Failed to write response:", err)
	}
}

func (h *ProjectHandler) UpdateTeamMemberRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	memberIDStr := vars["memberId"]
	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil {
		http.Error(w, "Invalid team member ID", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestBody.Role == "" {
		http.Error(w, "Role cannot be empty", http.StatusBadRequest)
		return
	}

	// Update the role of the team member in the database.
	err = h.projectService.UpdateTeamMemberRole(memberID, requestBody.Role)
	if err != nil {
		http.Error(w, "Failed to update team member role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // Respond with no content on success.
}
