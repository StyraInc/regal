package report

import "fmt"

// RelatedResource provides documentation on a violation.
type RelatedResource struct {
	Description string `json:"description"`
	Reference   string `json:"ref"`
}

// Location provides information on the location of a violation.
type Location struct {
	Column int     `json:"col"`
	Row    int     `json:"row"`
	Offset int     `json:"offset,omitempty"`
	File   string  `json:"file"`
	Text   *string `json:"text,omitempty"`
}

// Violation describes any violation found by Regal.
type Violation struct {
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Category         string            `json:"category"`
	RelatedResources []RelatedResource `json:"related_resources,omitempty"`
	Location         Location          `json:"location,omitempty"`
}

type Summary struct {
	FilesScanned  int `json:"files_scanned"`
	FilesFailed   int `json:"files_failed"`
	FilesSkipped  int `json:"files_skipped"`
	NumViolations int `json:"num_violations"`
}

// Report aggregate of Violation as returned by a linter run.
type Report struct {
	Violations []Violation `json:"violations"`
	Summary    Summary     `json:"summary"`
}

// ViolationsFileCount returns the number of files containing violations.
func (r Report) ViolationsFileCount() map[string]int {
	fc := map[string]int{}
	for _, violation := range r.Violations {
		fc[violation.Location.File]++
	}

	return fc
}

// String shorthand form for a Location.
func (l Location) String() string {
	if l.Row == 0 && l.Column == 0 {
		return l.File
	}

	return fmt.Sprintf("%s:%d:%d", l.File, l.Row, l.Column)
}
