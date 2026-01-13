package certification

import (
	"database/sql"
	"megome/internal/services/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetCertificationById(id int) (*types.Certification, error) {
	row := s.db.QueryRow("SELECT id, title, issuer, issueDate, expirationDate, credentialId, credentialUrl FROM certifications WHERE id = ?", id)
	certification := new(types.Certification)
	err := row.Scan(
		&certification.Title,
		&certification.Issuer,
		&certification.IssueDate,
		&certification.ExpirationDate,
		&certification.CredentialId,
		&certification.CredentialUrl,
	)
	if err != nil {
		return nil, err
	}
	return certification, nil
}

func (s *Store) GetCertifications(userId int) ([]types.Certification, error) {
	return nil, nil
}

func (s *Store) CreateCertification(certification types.Certification) error {
	_, err := s.db.Exec("INSERT INTO certifications (title, issuer, issueDate, expirationDate, credentialId, credentialUrl) VALUES (?, ?, ?, ?, ?, ?)",
		certification.Title,
		certification.Issuer,
		certification.IssueDate,
		certification.ExpirationDate,
		certification.CredentialId,
		certification.CredentialUrl,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateCertification(id int, certification types.Certification) error {
	_, err := s.db.Exec("UPDATE FROM certifications SET title = ?, issuer = ?, issueDate = ?, expirationDate = ?, credentialId = ?, credentialUrl = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		certification.Title,
		certification.Issuer,
		certification.IssueDate,
		certification.ExpirationDate,
		certification.CredentialId,
		certification.CredentialUrl,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteCertification(id int) error {
	_, err := s.GetCertificationById(id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("DELETE FROM certifications WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}
