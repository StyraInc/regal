// nolint:wrapcheck
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/styrainc/regal/internal/embeds"
)

// The revive check will warn about using underscore in struct names, but it's seemingly not aware of keywords.
//
//nolint:revive
type newRuleCommandParams struct {
	type_    string // 'type' is a keyword
	category string
	name     string
	output   string
}

type TemplateValues struct {
	Category     string
	NameOriginal string
	Name         string
	NameTest     string
}

var (
	categoryRegex = regexp.MustCompile(`^[a-z]+$`)
	nameRegex     = regexp.MustCompile(`^[a-z_]+[a-z0-9_\-]*$`)
)

//nolint:lll
func init() {
	newCommand := &cobra.Command{
		Hidden: true,
		Use:    "new <template>",
		Long: `Create a new resource according to the chosen template (currently only 'rule' available).

The new command is a development utility for scaffolding new resources for use by Regal.

An example of such a resource would be new linter rules, which could be created either for inclusion in Regal core, or custom rules for your organization or team.`,
	}

	params := newRuleCommandParams{}

	newRuleCommand := &cobra.Command{
		Use:   "rule [-t type] [-c category] [-n name]",
		Short: "Create new rule from template",
		Long: `Create a new linter rule, for inclusion in Regal or a custom rule for your organization or team.

Example:

regal new rule --type custom --category naming --name camel-case`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if params.type_ != "custom" && params.type_ != "builtin" {
				return fmt.Errorf("type must be 'custom' or 'builtin', got %v", params.type_)
			}

			if params.category == "" {
				return fmt.Errorf("category is required for rule")
			}

			if !categoryRegex.MatchString(params.category) {
				return fmt.Errorf("category must be a single word, using lowercase letters only")
			}

			if params.name == "" {
				return fmt.Errorf("name is required for rule")
			}

			if !nameRegex.MatchString(params.name) {
				return fmt.Errorf("name must consist only of lowercase letters, numbers, underscores and dashes")
			}

			return nil
		},

		Run: func(_ *cobra.Command, args []string) {
			if err := scaffoldRule(params); err != nil {
				log.SetOutput(os.Stderr)
				log.Println(err)
				os.Exit(1)
			}
		},
	}

	newRuleCommand.Flags().StringVarP(&params.type_, "type", "t", "custom", "type of rule (custom or builtin)")
	newRuleCommand.Flags().StringVarP(&params.category, "category", "c", "", "category for rule")
	newRuleCommand.Flags().StringVarP(&params.name, "name", "n", "", "name of rule")
	newRuleCommand.Flags().StringVarP(&params.output, "output", "o", "", "output directory")

	newCommand.AddCommand(newRuleCommand)
	RootCommand.AddCommand(newCommand)
}

func scaffoldRule(params newRuleCommandParams) error {
	if params.output == "" {
		params.output = mustGetWd()
	}

	if params.type_ == "custom" {
		return scaffoldCustomRule(params)
	}

	if params.type_ == "builtin" {
		return scaffoldBuiltinRule(params)
	}

	return fmt.Errorf("unsupported type %v", params.type_)
}

func scaffoldCustomRule(params newRuleCommandParams) error {
	rulesDir := filepath.Join(params.output, ".regal", "rules", params.category)

	if err := os.MkdirAll(rulesDir, 0o770); err != nil {
		return err
	}

	ruleTmpl, err := template.ParseFS(embeds.EmbedTemplatesFS, "templates/custom/custom.rego.tpl")
	if err != nil {
		return err
	}

	ruleFileName := strings.ToLower(strings.ReplaceAll(params.name, "-", "_")) + ".rego"

	ruleFile, err := os.Create(filepath.Join(rulesDir, ruleFileName))
	if err != nil {
		return err
	}

	err = ruleTmpl.Execute(ruleFile, templateValues(params))
	if err != nil {
		return err
	}

	testTmpl, err := template.ParseFS(embeds.EmbedTemplatesFS, "templates/custom/custom_test.rego.tpl")
	if err != nil {
		return err
	}

	testFileName := strings.ToLower(strings.ReplaceAll(params.name, "-", "_")) + "_test.rego"

	testFile, err := os.Create(filepath.Join(rulesDir, testFileName))
	if err != nil {
		return err
	}

	err = testTmpl.Execute(testFile, templateValues(params))
	if err != nil {
		return err
	}

	log.Printf("Created custom rule %q in %s\n", params.name, rulesDir)

	return nil
}

func scaffoldBuiltinRule(params newRuleCommandParams) error {
	rulesDir := filepath.Join(params.output, "bundle", "regal", "rules", params.category)

	if err := os.MkdirAll(rulesDir, 0o770); err != nil {
		return err
	}

	ruleTmpl, err := template.ParseFS(embeds.EmbedTemplatesFS, "templates/builtin/builtin.rego.tpl")
	if err != nil {
		return err
	}

	ruleFileName := strings.ToLower(strings.ReplaceAll(params.name, "-", "_")) + ".rego"

	ruleFile, err := os.Create(filepath.Join(rulesDir, ruleFileName))
	if err != nil {
		return err
	}

	err = ruleTmpl.Execute(ruleFile, templateValues(params))
	if err != nil {
		return err
	}

	testTmpl, err := template.ParseFS(embeds.EmbedTemplatesFS, "templates/builtin/builtin_test.rego.tpl")
	if err != nil {
		return err
	}

	testFileName := strings.ToLower(strings.ReplaceAll(params.name, "-", "_")) + "_test.rego"

	testFile, err := os.Create(filepath.Join(rulesDir, testFileName))
	if err != nil {
		return err
	}

	err = testTmpl.Execute(testFile, templateValues(params))
	if err != nil {
		return err
	}

	log.Printf("Created builtin rule %q in %s\n", params.name, rulesDir)

	return nil
}

func templateValues(params newRuleCommandParams) TemplateValues {
	var tmplNameValue string

	if strings.Contains(params.name, "-") {
		tmplNameValue = `["` + params.name + `"]`
	} else if strings.Contains(params.name, "_") {
		var tempName = strings.ReplaceAll(params.name, "_", "-")
		tmplNameValue = `["` + tempName + `"]`
	} else {
		tmplNameValue = "." + params.name
	}

	var tmplNameTestValue string

	if strings.Contains(params.name, "-") {
		tmplNameTestValue = `["` + params.name + `-test"]`
	} else if strings.Contains(params.name, "_") {
		var tempNameTest = strings.ReplaceAll(params.name, "_", "-")
		tmplNameTestValue = `["` + tempNameTest + `-test"]`
	} else {
		tmplNameTestValue = "." + params.name + "-test"
	}

	return TemplateValues{
		Category:     params.category,
		NameOriginal: strings.ReplaceAll(params.name, "_", "-"),
		Name:         tmplNameValue,
		NameTest:     tmplNameTestValue,
	}
}

func mustGetWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	return wd
}
