//nolint:lll
package hints

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/go-viper/mapstructure/v2"
)

// GetForError will return any matched, Styra-documented, errors for a given Go error.
func GetForError(e error) ([]string, error) {
	e0 := unwrapToFP(e)

	msgs, err := extractMessages(e0)
	if err != nil {
		return []string{}, fmt.Errorf("failed to extract messages: %w", err)
	}

	if len(msgs) < 1 {
		return []string{}, errors.New("no messages found")
	}

	msg := msgs[0]

	var hintKeys []string

	for u, r := range patterns {
		if r.MatchString(msg) {
			hintKeys = append(hintKeys, u)
		}
	}

	return hintKeys, nil
}

var patterns = map[string]*regexp.Regexp{
	`eval-conflict-error/complete-rules-must-not-produce-multiple-outputs`: regexp.MustCompile(`^eval_conflict_error: complete rules must not produce multiple outputs$`),
	`eval-conflict-error/object-keys-must-be-unique`:                       regexp.MustCompile(`^object insert conflict$|^eval_conflict_error: object keys must be unique$`),
	`rego-unsafe-var-error/var-name-is-unsafe`:                             regexp.MustCompile(`^rego_unsafe_var_error: var .* is unsafe$`),
	`rego-recursion-error/rule-name-is-recursive`:                          regexp.MustCompile(`^rego_recursion_error: rule .* is recursive:`),
	`rego-parse-error/var-cannot-be-used-for-rule-name`:                    regexp.MustCompile(`^rego_parse_error: var cannot be used for rule name$`),
	`rego-type-error/conflicting-rules-name-found`:                         regexp.MustCompile(`^rego_type_error: conflicting rules .* found$`),
	`rego-type-error/match-error`:                                          regexp.MustCompile(`^rego_type_error: match error`),
	`rego-type-error/arity-mismatch`:                                       regexp.MustCompile(`^rego_type_error: .*: arity mismatch`),
	`rego-type-error/function-has-arity-got-argument`:                      regexp.MustCompile(`^rego_type_error: function .* has arity [0-9]+, got [0-9]+ arguments?$`),
	`rego-compile-error/assigned-var-name-unused`:                          regexp.MustCompile(`^rego_compile_error: assigned var .* unused$`),
	`rego-parse-error/unexpected-assign-token`:                             regexp.MustCompile(`^rego_parse_error: unexpected assign token:`),
	`rego-parse-error/unexpected-identifier-token`:                         regexp.MustCompile(`^rego_parse_error: unexpected identifier token:`),
	`rego-parse-error/unexpected-left-curly-token`:                         regexp.MustCompile(`^rego_parse_error: unexpected { token:`),
	`rego-parse-error/unexpected-right-curly-token`:                        regexp.MustCompile(`^rego_parse_error: unexpected } token`),
	`rego-parse-error/unexpected-name-keyword`:                             regexp.MustCompile(`^rego_parse_error: unexpected .* keyword:`),
	`rego-parse-error/unexpected-string-token`:                             regexp.MustCompile(`^rego_parse_error: unexpected string token:`),
	`rego-type-error/multiple-default-rules-name-found`:                    regexp.MustCompile(`^rego_type_error: multiple default rules .* found`),
}

func unwrapToFP(e error) error {
	if w := errors.Unwrap(e); w != nil {
		return unwrapToFP(w)
	}

	return e
}

type message struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func extractMessages(e error) ([]string, error) {
	msgs := []message{}

	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			TagName: "json",
			Result:  &msgs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(e); err != nil {
		return nil, fmt.Errorf("failed to decode error: %w", err)
	}

	m := make([]string, len(msgs))

	for i := range msgs {
		m[i] = msgs[i].Code
		if m[i] != "" {
			m[i] += ": "
		}

		m[i] += msgs[i].Message
	}

	return m, nil
}
