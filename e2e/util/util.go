package util

import (
	"os"
	"testing"

	"github.com/open-policy-agent/regal/internal/util"
	"github.com/open-policy-agent/regal/pkg/report"
)

func ViolationTitles(rep report.Report) *util.Set[string] {
	titles := make([]string, len(rep.Violations))
	for i := range rep.Violations {
		titles[i] = rep.Violations[i].Title
	}

	return util.NewSet(titles...)
}

func AssertOnlyViolations(tb testing.TB, rep report.Report, expected ...string) {
	tb.Helper()

	violationNames := ViolationTitles(rep)
	if violationNames.Size() != len(expected) {
		tb.Errorf("expected %d violations, got %d: %v", len(expected), violationNames.Size(), violationNames.Items())
	}

	for _, name := range expected {
		if !violationNames.Contains(name) {
			tb.Errorf("expected violation for rule %q, but it was not found", name)
		}
	}
}

func AssertContainsViolations(tb testing.TB, rep report.Report, expected ...string) {
	tb.Helper()

	violationNames := ViolationTitles(rep)
	for _, name := range expected {
		if !violationNames.Contains(name) {
			tb.Errorf("expected violation for rule %q, but it was not found", name)
		}
	}
}

func AssertNotContainsViolations(tb testing.TB, rep report.Report, unexpected ...string) {
	tb.Helper()

	violationNames := ViolationTitles(rep)
	if violationNames.Contains(unexpected...) {
		tb.Errorf("expected no violations for rules %v, but found: %v", unexpected, violationNames.Items())
	}
}

func RemoveIgnoreErr(paths ...string) func() {
	return func() {
		for _, path := range paths {
			_ = os.Remove(path)
		}
	}
}
