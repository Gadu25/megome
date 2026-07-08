package completion

type ProfileField struct {
	Name  string `json:"name"`
	Filled bool  `json:"filled"`
}

type Section struct {
	Name    string `json:"name"`
	Filled  bool   `json:"filled"`
}

type CompletionResult struct {
	Overall int              `json:"overall"`
	Profile []ProfileField   `json:"profile"`
	Sections []Section       `json:"sections"`
}
