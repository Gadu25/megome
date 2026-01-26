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
		&certification.ID,
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
	rows, err := s.db.Query(
		"SELECT id, title, issuer, issueDate, expirationDate, credentialId, credentialUrl FROM certifications WHERE userId = ?",
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certifications []types.Certification
	for rows.Next() {
		cert, err := scanRowIntoCertification(rows)
		if err != nil {
			return nil, err
		}
		certifications = append(certifications, cert)
	}
	return certifications, nil
}

func (s *Store) CreateCertification(certification types.Certification) error {
	_, err := s.db.Exec("INSERT INTO certifications (title, issuer, issueDate, expirationDate, credentialId, credentialUrl, userId) VALUES (?, ?, ?, ?, ?, ?, ?)",
		certification.Title,
		certification.Issuer,
		certification.IssueDate,
		certification.ExpirationDate,
		certification.CredentialId,
		certification.CredentialUrl,
		certification.UserID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateCertification(id int, certification types.Certification) error {
	_, err := s.db.Exec("UPDATE certifications SET title = ?, issuer = ?, issueDate = ?, expirationDate = ?, credentialId = ?, credentialUrl = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
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

func scanRowIntoCertification(rows *sql.Rows) (types.Certification, error) {
	var certification types.Certification

	err := rows.Scan(
		&certification.ID,
		&certification.Title,
		&certification.Issuer,
		&certification.IssueDate,
		&certification.ExpirationDate,
		&certification.CredentialId,
		&certification.CredentialUrl,
	)
	if err != nil {
		return types.Certification{}, err
	}

	return certification, nil
}
