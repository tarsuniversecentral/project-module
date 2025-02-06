package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tarsuniversecentral/project-module/internal/dto"
)

type ProjectModel struct {
	db *sql.DB
}

func NewProjectModel(db *sql.DB) *ProjectModel {
	return &ProjectModel{db: db}
}

// CreateProjectTx wraps the entire project creation process in a transaction.
// It inserts the main project record and, if provided, inserts the associated
// pitch deck and image file paths into their respective tables.
func (m *ProjectModel) CreateProjectTx(p *dto.Project, lookingForStr string) error {
	// Begin the transaction.
	tx, err := m.db.Begin()
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
		if err = m.insertProjectPitchDecksTx(tx, p.ID, p.PitchDecks); err != nil {
			rollback(tx)
			return err
		}
	}

	// Insert image file paths if provided.
	if len(p.Images) > 0 {
		if err = m.insertProjectImagesTx(tx, p.ID, p.Images); err != nil {
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

func (m *ProjectModel) insertProjectPitchDecksTx(tx *sql.Tx, projectID int, paths []string) error {
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

func (m *ProjectModel) insertProjectImagesTx(tx *sql.Tx, projectID int, paths []string) error {
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

func (m *ProjectModel) GetProjects() ([]dto.Project, error) {
	rows, err := m.db.Query(`SELECT id, title, subtitle, industry, description, project_value, looking_for FROM projects`)
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

func (m *ProjectModel) GetProjectByID(id int) (*dto.Project, error) {
	var p dto.Project

	// Query to select the project by its ID
	row := m.db.QueryRow(`
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

	rows, err := m.db.Query(query, id)
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
	pitchRows, err := m.db.Query(pitchQuery, id)
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
	imageRows, err := m.db.Query(imageQuery, id)
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

func (m *ProjectModel) InsertTeamMember(member *dto.TeamMember) error {
	query := `
		INSERT INTO team_members (
			project_id, profile_url, title, role
		)
		VALUES (?, ?, ?, ?)`
	result, err := m.db.Exec(query, member.ProjectID, member.ProfileURL, member.Title, member.Role)
	if err != nil {
		log.Println("Error inserting team member:", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	member.ID = int(id)
	return nil
}

func (m *ProjectModel) GetTeamMembers(projectID int) ([]*dto.TeamMember, error) {
	query := `
		SELECT 
			id, 
			project_id, 
			profile_url, 
			title, 
			role
		FROM team_members
		WHERE project_id = ?`

	// Execute the query
	rows, err := m.db.Query(query, projectID)
	if err != nil {
		log.Println("Error querying team members:", err)
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Println("Error closing rows:", err)
		}
	}()

	var members []*dto.TeamMember

	// Iterate through the rows
	for rows.Next() {
		member := &dto.TeamMember{}
		if err := rows.Scan(
			&member.ID,
			&member.ProjectID,
			&member.ProfileURL,
			&member.Title,
			&member.Role,
		); err != nil {
			log.Println("Error scanning row:", err)
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		members = append(members, member)
	}

	// Check for errors after iteration
	if err := rows.Err(); err != nil {
		log.Println("Error after iterating rows:", err)
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return members, nil
}

func (m *ProjectModel) ProjectExists(projectID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = ?)`

	var exists bool
	err := m.db.QueryRow(query, projectID).Scan(&exists)
	if err != nil {
		log.Println("Error checking if project exists:", err)
		return false, fmt.Errorf("failed to check if project exists: %w", err)
	}

	return exists, nil
}

func (m *ProjectModel) UpdateTeamMemberRole(id int, role string) error {
	query := `
        UPDATE team_members
        SET role = ?, updated_at = CURRENT_TIMESTAMP
        WHERE id = ?`

	result, err := m.db.Exec(query, role, id)
	if err != nil {
		log.Println("Error updating team member role:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected, possibly invalid team member ID")
	}

	return nil
}
