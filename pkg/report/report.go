package report

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

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
	Level            string            `json:"level"`
	RelatedResources []RelatedResource `json:"related_resources,omitempty"`
	Location         Location          `json:"location,omitempty"`
	IsAggregate      bool              `json:"-"`
}

// Notice describes any notice found by Regal.
type Notice struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Level       string `json:"level"`
	Severity    string `json:"severity"`
}

// An Aggregate is data collected by some rule while processing a file AST, to be used later by other rules needing a
// global context (i.e. broader than per-file)
// Rule authors are expected to collect the minimum needed data, to avoid performance problems
// while working with large Rego code repositories.
type Aggregate map[string]any

type Summary struct {
	FilesScanned  int `json:"files_scanned"`
	FilesFailed   int `json:"files_failed"`
	RulesSkipped  int `json:"rules_skipped"`
	NumViolations int `json:"num_violations"`
}

// Report aggregate of Violation as returned by a linter run.
type Report struct {
	Violations []Violation `json:"violations"`
	// We don't have aggregates when publishing the final report (see JSONReporter), so omitempty is needed here
	// to avoid surfacing a null/empty field.
	Aggregates       map[string][]Aggregate         `json:"aggregates,omitempty"`
	Notices          []Notice                       `json:"notices,omitempty"`
	Summary          Summary                        `json:"summary"`
	Metrics          map[string]any                 `json:"metrics,omitempty"`
	AggregateProfile map[string]ProfileEntry        `json:"-"`
	Profile          []ProfileEntry                 `json:"profile,omitempty"`
	IgnoreDirectives map[string]map[string][]string `json:"ignore_directives,omitempty"`
}

// ProfileEntry is a single entry of profiling information, keyed by location.
// This data may have been aggregated across multiple runs.
type ProfileEntry struct {
	Location    string `json:"location"`
	TotalTimeNs int64  `json:"total_time_ns"`
	NumEval     int    `json:"num_eval"`
	NumRedo     int    `json:"num_redo"`
	NumGenExpr  int    `json:"num_gen_expr"`
}

func (r *Report) AddProfileEntries(prof map[string]ProfileEntry) {
	if r.AggregateProfile == nil {
		r.AggregateProfile = map[string]ProfileEntry{}
	}

	for loc, entry := range prof {
		if _, ok := r.AggregateProfile[loc]; !ok {
			r.AggregateProfile[loc] = entry
		} else {
			profCopy := r.AggregateProfile[loc]
			profCopy.NumEval += entry.NumEval
			profCopy.NumRedo += entry.NumRedo
			profCopy.NumGenExpr += entry.NumGenExpr
			profCopy.TotalTimeNs += entry.TotalTimeNs
			r.AggregateProfile[loc] = profCopy
		}
	}
}

func (r *Report) AggregateProfileToSortedProfile(numResults int) {
	r.Profile = make([]ProfileEntry, 0, len(r.AggregateProfile))

	for loc, rs := range r.AggregateProfile {
		rs.Location = loc

		r.Profile = append(r.Profile, r.AggregateProfile[loc])
	}

	sort.Slice(r.Profile, func(i, j int) bool {
		return r.Profile[i].TotalTimeNs > r.Profile[j].TotalTimeNs
	})

	if numResults <= 0 || numResults > len(r.Profile) {
		return
	}

	r.Profile = r.Profile[:numResults]
}

// TODO: This does not belong here and is only for internal testing purposes at this point in time. Profile reports are
// currently only publicly available for the JSON reporter. Some variation of this will eventually be moved to the table
// reporter. (this code borrowed from OPA).
func (r Report) printProfile(w io.Writer) { //nolint:unused
	tableProfile := generateTableProfile(w)

	for i, rs := range r.Profile {
		timeNs := time.Duration(rs.TotalTimeNs) * time.Nanosecond
		line := []string{
			timeNs.String(),
			strconv.Itoa(rs.NumEval),
			strconv.Itoa(rs.NumRedo),
			strconv.Itoa(rs.NumGenExpr),
			rs.Location,
		}
		tableProfile.Append(line)

		if i == 0 {
			tableProfile.SetFooter([]string{"", "", "", "", ""})
		}
	}

	if tableProfile.NumLines() > 0 {
		tableProfile.Render()
	}
}

func generateTableWithKeys(writer io.Writer, keys ...string) *tablewriter.Table { //nolint:unused
	table := tablewriter.NewWriter(writer)
	aligns := make([]int, 0, len(keys))
	hdrs := make([]string, 0, len(keys))

	for _, k := range keys {
		hdrs = append(hdrs, strings.Title(k)) //nolint:staticcheck // SA1019, no unicode here
		aligns = append(aligns, tablewriter.ALIGN_LEFT)
	}

	table.SetHeader(hdrs)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetColumnAlignment(aligns)

	return table
}

func generateTableProfile(writer io.Writer) *tablewriter.Table { //nolint:unused
	return generateTableWithKeys(writer, "Time", "Num Eval", "Num Redo", "Num Gen Expr", "Location")
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
