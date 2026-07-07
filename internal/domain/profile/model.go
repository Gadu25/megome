package profile

type ProfileStore interface {
	GetPublicProfile(int) (*Profile, error)
	GetProfile(userId int) (*Profile, error)
	MakeProfile(Profile) error
	UpsertOAuthProfile(Profile) error
}

type Profile struct {
	ID           int     `json:"id"`
	UserID       int     `json:"userId"`
	FirstName    string  `json:"firstName"`
	LastName     string  `json:"lastName"`
	Title        string  `json:"title"`
	Birthday     *string `json:"birthday"`
	Tagline      string  `json:"tagline"`
	Bio          string  `json:"bio"`
	Phone        string  `json:"phone"`
	Website      string  `json:"website"`
	Location     string  `json:"location"`
	ProfileImage string  `json:"profileImage"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type MakeProfilePayload struct {
	Bio       string  `json:"bio"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Tagline   string  `json:"tagline"`
	Birthday  *string `json:"birthday"`
	Title     string  `json:"title"`
	Phone     string  `json:"phone"`
	Website   string  `json:"website"`
	Location  string  `json:"location"`
}

type PublicProfileResponse struct {
	ID           int    `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Title        string `json:"title"`
	Birthday     string `json:"birthday"`
	Bio          string `json:"bio"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	Location     string `json:"location"`
	ProfileImage string `json:"profileImage"`
}

type ProfileResponse struct {
	Message string   `json:"message"`
	Profile *Profile `json:"profile"`
}
