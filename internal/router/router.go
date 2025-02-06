package router

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tarsuniversecentral/project-module/internal/api"
)

func Routers(router *mux.Router) http.Handler {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/welcome", welcome)

	return r
}

func welcome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome")
}

// NewRouter registers routes for all domains and returns a configured router.
func NewRouter(api *api.API) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Project routes.
	projectRouter := router.PathPrefix("/projects").Subrouter()
	projectRouter.HandleFunc("", api.ProjectHandler.CreateProject).Methods("POST")
	projectRouter.HandleFunc("/{id:[0-9]+}", api.ProjectHandler.GetProject).Methods("GET")
	projectRouter.HandleFunc("/file/{filename}", api.ProjectHandler.FileRetrieveHandler).Methods("GET")

	projectRouter.HandleFunc("/{projectId:[0-9]+}/teammember", api.ProjectHandler.AddTeamMemberToProject).Methods("POST")
	projectRouter.HandleFunc("/{projectId:[0-9]+}/teammembers", api.ProjectHandler.GetTeamMembersOfProject).Methods("GET")
	projectRouter.HandleFunc("/teammember/role/{memberId}", api.ProjectHandler.UpdateTeamMemberRole).Methods("PUT")

	return router
}
