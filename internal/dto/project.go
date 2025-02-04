package dto

import "fmt"

type LookingFor string

// Valid values for LookingFor.
const (
	Investment LookingFor = "Investment"
	Employees  LookingFor = "Employees"
	Partners   LookingFor = "Partners"
	Buyers     LookingFor = "Buyers"
)

type Project struct {
	ID           int          `json:"id"`
	Title        string       `json:"title"`
	Subtitle     string       `json:"subtitle,omitempty"`
	Industry     string       `json:"industry,omitempty"`
	Description  string       `json:"description,omitempty"`
	PitchDecks   []string     `json:"pitch_decks,omitempty"`
	ProjectValue float64      `json:"project_value,omitempty"`
	LookingFor   []string     `json:"looking_for,omitempty"`
	Images       []string     `json:"images,omitempty"`
	GithubLink   string       `json:"github_link,omitempty"`
	TeamMembers  []TeamMember `json:"team_members,omitempty"`
}

type TeamMember struct {
	ID         int    `json:"id"`
	ProjectID  int    `json:"project_id"`
	ProfileURL string `json:"profile_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Role       string `json:"role,omitempty"`
}

var validLookingForValues = map[LookingFor]struct{}{
	Investment: {},
	Employees:  {},
	Partners:   {},
	Buyers:     {},
}

func ValidateLookingFor(values []string) error {
	for _, v := range values {
		lf := LookingFor(v)
		if _, ok := validLookingForValues[lf]; !ok {
			return fmt.Errorf("invalid looking_for value: %q", v)
		}
	}
	return nil
}
