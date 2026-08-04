package assist

const (
	TaskGenerateBio                = "generate_bio"
	TaskGenerateTagline            = "generate_tagline"
	TaskGenerateProjectDescription = "generate_project_description"
	TaskGenerateExperience         = "generate_experience"
	TaskGenerateEducation          = "generate_education"
)

type Request struct {
	Task    string            `json:"task"`
	Context map[string]string `json:"context"`
	Extra   string            `json:"extra"`
}
