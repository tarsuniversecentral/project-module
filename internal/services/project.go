package service

import (
	"strings"

	"github.com/tarsuniversecentral/project-module/internal/dto"
	model "github.com/tarsuniversecentral/project-module/internal/models"
)

type ProjectService struct {
	model *model.ProjectModel
}

func NewProjectService(model *model.ProjectModel) *ProjectService {
	return &ProjectService{model: model}
}

func (s *ProjectService) CreateProject(project dto.Project) (*dto.Project, error) {

	lookingForStr := strings.Join(project.LookingFor, ",")

	err := s.model.CreateProjectTx(&project, lookingForStr)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (s *ProjectService) GetProject(id int) (*dto.Project, error) {
	project, err := s.model.GetProjectByID(id)
	if err != nil {
		return nil, err
	}

	return project, nil
}
