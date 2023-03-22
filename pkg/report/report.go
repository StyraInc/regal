package report

// RelatedResource provides documentation on a violation.
type RelatedResource struct {
	Description string `json:"description"`
	Reference   string `json:"ref"`
}

// Location provides information on the location of a violation.
type Location struct {
	Column int    `json:"col"`
	Row    int    `json:"row"`
	Offset int    `json:"offset,omitempty"`
	File   string `json:"file"`
	Text   []byte `json:"text,omitempty"`
}

// Violation describes any violation found by Regal.
type Violation struct {
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Category         string            `json:"category"`
	RelatedResources []RelatedResource `json:"related_resources,omitempty"`
	Location         Location          `json:"location,omitempty"`
}

// Report aggregate of Violation as returned by a linter run.
type Report struct {
	Violations []Violation `json:"report"`
}
