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
	row := s.db.QueryRow("SELECT id, title, issuer, issueDate, certificateImage, expirationDate, credentialId, credentialUrl FROM certifications WHERE id = ?", id)

	var certification Certification
	err := row.Scan(
		&certification.ID,
		&certification.Title,
		&certification.Issuer,
		&certification.IssueDate,
		&certification.CertificateImage,
		&certification.ExpirationDate,
		&certification.CredentialId,
		&certification.CredentialUrl,
	)

	if err != nil {
		return Certification{}, err
	}

	if certification.CertificateImage != nil && *certification.CertificateImage != "" {
		isFullURL := strings.HasPrefix(*certification.CertificateImage, "https://") || strings.HasPrefix(*certification.CertificateImage, "http://")

		if !isFullURL {
			*certification.CertificateImage = httputil.GetPublicFile(*certification.CertificateImage)
		}
	}

	return certification, nil
}

func (s *Repository) GetCertifications(userId int) ([]Certification, error) {
	rows, err := s.db.Query(
		"SELECT id, title, issuer, issueDate, certificateImage, expirationDate, credentialId, credentialUrl FROM certifications WHERE userId = ?",
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	certifications := make([]Certification, 0)

	for rows.Next() {
		cert, err := scanRowIntoCertification(rows)
		if err != nil {
			return nil, err
		}
		certifications = append(certifications, cert)
	}
	return certifications, nil
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

func (s *Repository) DeleteCertification(id int) (Certification, error) {
	cert, err := s.GetCertificationById(id)
	if err != nil {
		return Certification{}, err
	}

	_, err = s.db.Exec("DELETE FROM certifications WHERE id = ?", id)
	if err != nil {
		return Certification{}, err
	}

	return cert, nil
}

func scanRowIntoCertification(rows *sql.Rows) (Certification, error) {
	var certification Certification

	err := rows.Scan(
		&certification.ID,
		&certification.Title,
		&certification.Issuer,
		&certification.IssueDate,
		&certification.CertificateImage,
		&certification.ExpirationDate,
		&certification.CredentialId,
		&certification.CredentialUrl,
	)
	if err != nil {
		return Certification{}, err
	}

	if certification.CertificateImage != nil && *certification.CertificateImage != "" {
		isFullURL := strings.HasPrefix(*certification.CertificateImage, "https://") || strings.HasPrefix(*certification.CertificateImage, "http://")

		if !isFullURL {
			*certification.CertificateImage = httputil.GetPublicFile(*certification.CertificateImage)
		}
	}

	return certification, nil
}
