package certification

type CertificationStore interface {
	GetPublicCertifications(userId int) ([]Certification, error)
	GetCertifications(userId int) ([]Certification, error)
	GetCertificationById(id int) (Certification, error)
	CreateCertification(Certification) (Certification, error)
	UpdateCertification(id int, certification Certification) (Certification, error)
	DeleteCertification(id int) (Certification, error)
}

type Certification struct {
	ID               int     `json:"id"`
	UserID           int     `json:"userId"`
	Title            string  `json:"title"`
	Issuer           string  `json:"issuer"`
	IssueDate        string  `json:"issueDate"`
	CertificateImage *string `json:"certificateImage"`
	ExpirationDate   *string `json:"expirationDate"`
	CredentialId     *string `json:"credentialId"`
	CredentialUrl    *string `json:"credentialUrl"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}

type CertificationPayload struct {
	Title            string  `json:"title" validate:"required"`
	Issuer           string  `json:"issuer" validate:"required"`
	IssueDate        string  `json:"issueDate" validate:"required"`
	CertificateImage *string `json:"certificateImage"`
	ExpirationDate   *string `json:"expirationDate"`
	CredentialId     *string `json:"credentialId"`
	CredentialUrl    *string `json:"credentialUrl"`
}
