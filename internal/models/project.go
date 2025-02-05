package model

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tarsuniversecentral/project-module/internal/dto"
)

type ProjectModel struct {
	DB *sql.DB
}

func NewProjectModel(db *sql.DB) *ProjectModel {
	return &ProjectModel{DB: db}
}

// CreateProjectTx wraps the entire project creation process in a transaction.
// It inserts the main project record and, if provided, inserts the associated
// pitch deck and image file paths into their respective tables.
func (r *ProjectModel) CreateProjectTx(p *dto.Project, lookingForStr string) error {
	// Begin the transaction.
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	// In case of any error, roll back the transaction.
	rollback := func(tx *sql.Tx) {
		if rErr := tx.Rollback(); rErr != nil {
			log.Printf("Error rolling back transaction: %v", rErr)
		}
	}

	// Insert the main project record.
	projectQuery := `
		INSERT INTO projects (title, subtitle, industry, description, project_value, looking_for, github_link)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.Exec(projectQuery,
		p.Title,
		p.Subtitle,
		p.Industry,
		p.Description,
		p.ProjectValue,
		lookingForStr,
		p.GithubLink,
	)
	if err != nil {
		rollback(tx)
		log.Println("Error inserting project:", err)
		return err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		rollback(tx)
		log.Println("Error getting last insert ID:", err)
		return err
	}

	p.ID = int(lastInsertID)

	// Insert pitch deck file paths if provided.
	if len(p.PitchDecks) > 0 {
		if err = r.insertProjectPitchDecksTx(tx, p.ID, p.PitchDecks); err != nil {
			rollback(tx)
			return err
		}
	}

	// Insert image file paths if provided.
	if len(p.Images) > 0 {
		if err = r.insertProjectImagesTx(tx, p.ID, p.Images); err != nil {
			rollback(tx)
			return err
		}
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	return nil
}

func (r *ProjectModel) insertProjectPitchDecksTx(tx *sql.Tx, projectID int, paths []string) error {
	// Return early if there are no paths to insert.
	if len(paths) == 0 {
		return nil
	}

	// Build the INSERT query dynamically.
	// For each file, we need a placeholder group "(?, ?)".
	query := "INSERT INTO project_pitch_decks (project_id, file_path) VALUES "
	placeholders := make([]string, 0, len(paths))
	values := make([]interface{}, 0, len(paths)*2)

	for _, path := range paths {
		placeholders = append(placeholders, "(?, ?)")
		values = append(values, projectID, path)
	}
	query += strings.Join(placeholders, ",")

	// Execute the batch insert.
	if _, err := tx.Exec(query, values...); err != nil {
		log.Println("Error batch inserting pitch decks:", err)
		return err
	}
	return nil
}

func (r *ProjectModel) insertProjectImagesTx(tx *sql.Tx, projectID int, paths []string) error {
	// Return early if there are no paths to insert.
	if len(paths) == 0 {
		return nil
	}

	// Build the INSERT query dynamically.
	query := "INSERT INTO project_images (project_id, file_path) VALUES "
	placeholders := make([]string, 0, len(paths))
	values := make([]interface{}, 0, len(paths)*2)

	for _, path := range paths {
		placeholders = append(placeholders, "(?, ?)")
		values = append(values, projectID, path)
	}
	query += strings.Join(placeholders, ",")

	// Execute the batch insert.
	if _, err := tx.Exec(query, values...); err != nil {
		log.Println("Error batch inserting images:", err)
		return err
	}
	return nil
}

func (r *ProjectModel) GetProjects() ([]dto.Project, error) {
	rows, err := r.DB.Query(`SELECT id, title, subtitle, industry, description, project_value, looking_for FROM projects`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []dto.Project
	for rows.Next() {
		var p dto.Project
		if err := rows.Scan(&p.ID, &p.Title, &p.Subtitle, &p.Industry, &p.Description, &p.ProjectValue, &p.LookingFor); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *ProjectModel) GetProjectByID(id int) (*dto.Project, error) {
	var p dto.Project

	// Query to select the project by its ID
	row := r.DB.QueryRow(`
		SELECT id, title, subtitle, industry, description, project_value, looking_for
		FROM projects
		WHERE id = ?
	`, id)

	var lookingFor sql.NullString
	// Scan the row into the project struct
	err := row.Scan(
		&p.ID, &p.Title, &p.Subtitle, &p.Industry, &p.Description, &p.ProjectValue, &lookingFor,
	)

	if lookingFor.Valid {
		p.LookingFor = parseLookingFor(lookingFor.String)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}

	return &p, nil
}

func (m *ProjectModel) GetProjectFullDetails(id int) (*dto.Project, error) {
	query := `
		SELECT 
			p.id, 
			p.title, 
			p.subtitle, 
			p.industry, 
			p.description, 
			p.project_value, 
			p.looking_for, 
			p.github_link,
			tm.id, 
			tm.project_id, 
			tm.profile_url, 
			tm.title, 
			tm.role
		FROM projects p
		LEFT JOIN team_members tm ON p.id = tm.project_id
		WHERE p.id = ?
	`

	rows, err := m.DB.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var project *dto.Project
	for rows.Next() {
		// Project columns.
		var (
			pID          int
			title        string
			subtitle     sql.NullString
			industry     sql.NullString
			description  sql.NullString
			projectValue float64
			lookingFor   sql.NullString // Comma-separated list
			githubLink   sql.NullString
		)
		// Team member columns.
		var (
			tmID         sql.NullInt64
			tmProjectID  sql.NullInt64
			tmProfileURL sql.NullString
			tmTitle      sql.NullString
			tmRole       sql.NullString
		)

		err = rows.Scan(
			&pID,
			&title,
			&subtitle,
			&industry,
			&description,
			&projectValue,
			&lookingFor,
			&githubLink,
			&tmID,
			&tmProjectID,
			&tmProfileURL,
			&tmTitle,
			&tmRole,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// On the first row, initialize the project.
		if project == nil {
			project = &dto.Project{
				ID:           pID,
				Title:        title,
				Subtitle:     subtitle.String,
				Industry:     industry.String,
				Description:  description.String,
				ProjectValue: projectValue,
				LookingFor:   parseLookingFor(lookingFor.String),
				GithubLink:   githubLink.String,
				TeamMembers:  []dto.TeamMember{},
				PitchDecks:   []string{},
				Images:       []string{},
			}
		}

		// If team member data is present, add it.
		if tmID.Valid {
			teamMember := dto.TeamMember{
				ID:         int(tmID.Int64),
				ProjectID:  int(tmProjectID.Int64),
				ProfileURL: tmProfileURL.String,
				Title:      tmTitle.String,
				Role:       tmRole.String,
			}
			project.TeamMembers = append(project.TeamMembers, teamMember)
		}
	}

	if project == nil {
		return nil, sql.ErrNoRows
	}

	// Now, query for pitch deck file paths.
	pitchQuery := `SELECT file_path FROM project_pitch_decks WHERE project_id = ?`
	pitchRows, err := m.DB.Query(pitchQuery, id)
	if err != nil {
		return nil, fmt.Errorf("query pitch decks error: %w", err)
	}
	defer pitchRows.Close()

	var pitchDecks []string
	for pitchRows.Next() {
		var filePath string
		if err := pitchRows.Scan(&filePath); err != nil {
			return nil, fmt.Errorf("scan pitch deck error: %w", err)
		}
		pitchDecks = append(pitchDecks, filePath)
	}
	// Set the PitchDecks field on the project.
	project.PitchDecks = pitchDecks

	// Similarly, query for image file paths.
	imageQuery := `SELECT file_path FROM project_images WHERE project_id = ?`
	imageRows, err := m.DB.Query(imageQuery, id)
	if err != nil {
		return nil, fmt.Errorf("query images error: %w", err)
	}
	defer imageRows.Close()

	var images []string
	for imageRows.Next() {
		var filePath string
		if err := imageRows.Scan(&filePath); err != nil {
			return nil, fmt.Errorf("scan image error: %w", err)
		}
		images = append(images, filePath)
	}
	// Set the Images field on the project.
	project.Images = images

	return project, nil
}

// parseLookingFor converts a comma-separated string into a slice of strings.
func parseLookingFor(s string) []string {
	if s == "" {
		return []string{}
	}
	return splitAndTrim(s, ",")
}

// splitAndTrim splits a string by the given delimiter and trims spaces.
func splitAndTrim(s, delim string) []string {
	parts := strings.Split(s, delim)
	var result []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
