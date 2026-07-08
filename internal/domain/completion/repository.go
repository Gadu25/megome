package completion

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetCompletion(userID int) (*CompletionResult, error) {
	var firstName, lastName, tagline, bio, title, phone, website, location, profileImage sql.NullString

	err := r.db.QueryRow(`
		SELECT firstName, lastName, tagline, bio, title, phone, website, location, profileImage
		FROM profiles
		WHERE userId = ?
	`, userID).Scan(
		&firstName, &lastName, &tagline, &bio, &title, &phone, &website, &location, &profileImage,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	fields := []ProfileField{
		{Name: "firstName", Filled: firstName.Valid && firstName.String != ""},
		{Name: "lastName", Filled: lastName.Valid && lastName.String != ""},
		{Name: "tagline", Filled: tagline.Valid && tagline.String != ""},
		{Name: "bio", Filled: bio.Valid && bio.String != ""},
		{Name: "title", Filled: title.Valid && title.String != ""},
		{Name: "phone", Filled: phone.Valid && phone.String != ""},
		{Name: "website", Filled: website.Valid && website.String != ""},
		{Name: "location", Filled: location.Valid && location.String != ""},
		{Name: "profileImage", Filled: profileImage.Valid && profileImage.String != ""},
	}

	var skillCount, educationCount, experienceCount, certCount, projectCount int
	r.db.QueryRow("SELECT COUNT(*) FROM skills WHERE userId = ?", userID).Scan(&skillCount)
	r.db.QueryRow("SELECT COUNT(*) FROM education WHERE userId = ?", userID).Scan(&educationCount)
	r.db.QueryRow("SELECT COUNT(*) FROM experiences WHERE userId = ?", userID).Scan(&experienceCount)
	r.db.QueryRow("SELECT COUNT(*) FROM certifications WHERE userId = ?", userID).Scan(&certCount)
	r.db.QueryRow("SELECT COUNT(*) FROM projects WHERE userId = ? AND isDraft = false", userID).Scan(&projectCount)

	sections := []Section{
		{Name: "skills", Filled: skillCount > 0},
		{Name: "education", Filled: educationCount > 0},
		{Name: "experience", Filled: experienceCount > 0},
		{Name: "certification", Filled: certCount > 0},
		{Name: "projects", Filled: projectCount > 0},
	}

	total := len(fields) + len(sections)
	filled := 0
	for _, f := range fields {
		if f.Filled {
			filled++
		}
	}
	for _, s := range sections {
		if s.Filled {
			filled++
		}
	}

	overall := 0
	if total > 0 {
		overall = filled * 100 / total
	}

	return &CompletionResult{
		Overall:  overall,
		Profile:  fields,
		Sections: sections,
	}, nil
}
