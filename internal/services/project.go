package services

import (
	"fmt"
	"strings"

	"github.com/tarsuniversecentral/project-module/internal/dto"
	"github.com/tarsuniversecentral/project-module/internal/models"
)

type ProjectService struct {
	model *models.ProjectModel
}

func NewProjectService(model *models.ProjectModel) *ProjectService {
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

	if err := s.validateProjectExists(id); err != nil {
		return nil, err
	}

	project, err := s.model.GetProjectFullDetails(id)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) AddTeamMember(teamMember *dto.TeamMember) error {

	if err := s.validateProjectExists(teamMember.ProjectID); err != nil {
		return err
	}

	err := s.model.InsertTeamMember(teamMember)
	if err != nil {
		return err
	}

	return nil
}

func (s *ProjectService) GetTeamMembers(id int) ([]*dto.TeamMember, error) {

	if err := s.validateProjectExists(id); err != nil {
		return nil, err
	}

	teamMembers, err := s.model.GetTeamMembers(id)
	if err != nil {
		return nil, err
	}

	return teamMembers, nil
}

func (s *ProjectService) UpdateTeamMemberRole(id int, role string) error {

	err := s.model.UpdateTeamMemberRole(id, role)
	if err != nil {
		return err
	}

	return nil
}

func (s *ProjectService) validateProjectExists(id int) error {
	exists, err := s.model.ProjectExists(id)
	if err != nil {
		return fmt.Errorf("failed to validate project: %w", err)
	}
	if !exists {
		return fmt.Errorf("project with ID %d does not exist", id)
	}
	return nil
}
