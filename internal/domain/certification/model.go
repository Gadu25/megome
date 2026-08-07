package certification

type CertificationStore interface {
	GetCertifications(userId int, limit int, offset int) ([]Certification, error)
	GetCertificationById(id int) (Certification, error)
	CreateCertification(Certification) (Certification, error)
	UpdateCertification(id int, certification Certification) (Certification, error)
	DeleteCertification(id int, deletedBy int) (Certification, error)
	CountByUserID(userId int) (int, error)
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
	DisplayOrder     int     `json:"displayOrder"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
	DeletedAt        *string `json:"deletedAt,omitempty"`
}

type ReorderItem struct {
	ID           int `json:"id"`
	DisplayOrder int `json:"displayOrder"`
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

type PaginatedCertificationsResponse struct {
	Data       []Certification `json:"data"`
	Pagination struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Total  int `json:"total"`
	} `json:"pagination"`
}
