package api

import (
	handler "github.com/tarsuniversecentral/project-module/internal/handlers"
)

type API struct {
	ProjectHandler *handler.ProjectHandler
}

func NewAPI(projectHandler *handler.ProjectHandler) *API {
	return &API{
		ProjectHandler: projectHandler,
	}
}
