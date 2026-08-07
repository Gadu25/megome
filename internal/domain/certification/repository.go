package certification

import (
	"database/sql"
	"megome/internal/pkg/httputil"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) GetCertificationById(id int) (Certification, error) {
	row := s.db.QueryRow("SELECT id, userId, title, issuer, issueDate, certificateImage, expirationDate, credentialId, credentialUrl, displayOrder, deletedAt, createdAt, updatedAt FROM certifications WHERE id = ? AND deletedAt IS NULL", id)
	return scanRowIntoCertification(row)
}

func (s *Repository) GetCertifications(userId int, limit int, offset int) ([]Certification, error) {
	rows, err := s.db.Query(
		"SELECT id, userId, title, issuer, issueDate, certificateImage, expirationDate, credentialId, credentialUrl, displayOrder, deletedAt, createdAt, updatedAt FROM certifications WHERE userId = ? AND deletedAt IS NULL ORDER BY displayOrder ASC, id ASC LIMIT ? OFFSET ?",
		userId,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRowsIntoCertification(rows)
}

func (s *Repository) CreateCertification(certification Certification) (Certification, error) {
	result, err := s.db.Exec(
		"INSERT INTO certifications (title, issuer, issueDate, certificateImage, expirationDate, credentialId, credentialUrl, userId) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		certification.Title,
		certification.Issuer,
		certification.IssueDate,
		httputil.NilIfEmpty(certification.CertificateImage),
		certification.ExpirationDate,
		certification.CredentialId,
		certification.CredentialUrl,
		certification.UserID,
	)
	if err != nil {
		return Certification{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Certification{}, err
	}

	return s.GetCertificationById(int(id))
}

func (s *Repository) UpdateCertification(id int, certification Certification) (Certification, error) {
	_, err := s.db.Exec(
		"UPDATE certifications SET title = ?, issuer = ?, issueDate = ?, certificateImage = ?, expirationDate = ?, credentialId = ?, credentialUrl = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		certification.Title,
		certification.Issuer,
		certification.IssueDate,
		httputil.NilIfEmpty(certification.CertificateImage),
		certification.ExpirationDate,
		certification.CredentialId,
		certification.CredentialUrl,
		id,
	)
	if err != nil {
		return Certification{}, err
	}

	return s.GetCertificationById(id)
}

func (s *Repository) DeleteCertification(id int, deletedBy int) (Certification, error) {
	cert, err := s.GetCertificationById(id)
	if err != nil {
		return Certification{}, err
	}

	_, err = s.db.Exec("UPDATE certifications SET deletedAt = NOW() WHERE id = ? AND deletedAt IS NULL", id)
	if err != nil {
		return Certification{}, err
	}

	return cert, nil
}

func scanRowIntoCertification(scanner interface{ Scan(dest ...interface{}) error }) (Certification, error) {
	var certification Certification
	err := scanner.Scan(
		&certification.ID,
		&certification.UserID,
		&certification.Title,
		&certification.Issuer,
		&certification.IssueDate,
		&certification.CertificateImage,
		&certification.ExpirationDate,
		&certification.CredentialId,
		&certification.CredentialUrl,
		&certification.DisplayOrder,
		&certification.DeletedAt,
		&certification.CreatedAt,
		&certification.UpdatedAt,
	)
	if err != nil {
		return Certification{}, err
	}

	if certification.CertificateImage != nil && *certification.CertificateImage != "" {
		if !strings.HasPrefix(*certification.CertificateImage, "https://") && !strings.HasPrefix(*certification.CertificateImage, "http://") {
			*certification.CertificateImage = httputil.GetPublicFile(*certification.CertificateImage)
		}
	}

	return certification, nil
}

func scanRowsIntoCertification(rows *sql.Rows) ([]Certification, error) {
	var certifications []Certification
	for rows.Next() {
		cert, err := scanRowIntoCertification(rows)
		if err != nil {
			return nil, err
		}
		certifications = append(certifications, cert)
	}
	return certifications, rows.Err()
}

func (s *Repository) ReorderCertifications(userID int, items []ReorderItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE certifications SET displayOrder = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ? AND userId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(item.DisplayOrder, item.ID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Repository) CountByUserID(userID int) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM certifications WHERE userId = ? AND deletedAt IS NULL", userID).Scan(&count)
	return count, err
}
