package fixer

import (
	"fmt"
	"io"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"

	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/report"
)

func Fix(rep *report.Report, readers map[string]io.Reader, _ fixes.Options) (map[string][]byte, error) {
	fixableViolations := map[string]struct{}{
		"opa-fmt":     {},
		"use-rego-v1": {},
	}

	filesToFix, err := computeFilesToFix(rep, readers, fixableViolations)
	if err != nil {
		return nil, fmt.Errorf("failed to determine files to fix: %w", err)
	}

	fixResults := make(map[string][]byte)

	var fixedViolations []int

	for file, content := range filesToFix {
		for i, violation := range rep.Violations {
			_, ok := fixableViolations[violation.Title]
			if !ok {
				continue
			}

			fixed := true

			switch violation.Title {
			case "opa-fmt":
				fixed, fixedContent, err := fixes.Fmt(content, &fixes.FmtOptions{
					Filename: file,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to fix %s: %w", file, err)
				}

				if fixed {
					fixResults[file] = fixedContent
				}
			case "use-rego-v1":
				fixed, fixedContent, err := fixes.Fmt(content, &fixes.FmtOptions{
					Filename: file,
					OPAFmtOpts: format.Opts{
						RegoVersion: ast.RegoV0CompatV1,
					},
				})
				if err != nil {
					return nil, fmt.Errorf("failed to fix %s: %w", file, err)
				}

				if fixed {
					fixResults[file] = fixedContent
				}
			default:
				fixed = false
			}

			if fixed {
				fixedViolations = append(fixedViolations, i)
			}
		}
	}

	for i := len(fixedViolations) - 1; i >= 0; i-- {
		rep.Violations = append(rep.Violations[:fixedViolations[i]], rep.Violations[fixedViolations[i]+1:]...)
	}

	rep.Summary.NumViolations = len(rep.Violations)

	return fixResults, nil
}

func computeFilesToFix(
	rep *report.Report,
	readers map[string]io.Reader,
	fixableViolations map[string]struct{},
) (map[string][]byte, error) {
	filesToFix := make(map[string][]byte)

	// determine which files need to be fixed
	for _, violation := range rep.Violations {
		file := violation.Location.File

		// skip files already marked for fixing
		if _, ok := filesToFix[file]; ok {
			continue
		}

		// skip violations that are not fixable
		if _, ok := fixableViolations[violation.Title]; !ok {
			continue
		}

		if _, ok := readers[file]; !ok {
			return nil, fmt.Errorf("no reader for fixable file %s", file)
		}

		bs, err := io.ReadAll(readers[file])
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		filesToFix[violation.Location.File] = bs
	}

	return filesToFix, nil
}
