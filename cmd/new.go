// nolint:wrapcheck
package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/styrainc/regal/internal/embeds"
	"github.com/styrainc/regal/pkg/config"
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

		PreRunE: func(*cobra.Command, []string) error {
			if params.type_ != "custom" && params.type_ != "builtin" {
				return fmt.Errorf("type must be 'custom' or 'builtin', got %v", params.type_)
			}

			if params.category == "" {
				return errors.New("category is required for rule")
			}

			if !categoryRegex.MatchString(params.category) {
				return errors.New("category must be a single word, using lowercase letters only")
			}

			if params.name == "" {
				return errors.New("name is required for rule")
			}

			if !nameRegex.MatchString(params.name) {
				return errors.New("name must consist only of lowercase letters, numbers, underscores and dashes")
			}

			return nil
		},

		Run: func(*cobra.Command, []string) {
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
		if err := scaffoldBuiltinRule(params); err != nil {
			return err
		}

		if err := createBuiltinDocs(params); err != nil {
			return err
		}

		if err := addToDataYAML(params); err != nil {
			return err
		}

		r, err := createTable([]string{filepath.Join("bundle", "regal")})
		if err != nil {
			return err
		}

		newReadme := renderREADME(r)
		readmePath := filepath.Join(params.output, "README.md")

		return os.WriteFile(readmePath, []byte(newReadme), 0o600)
	}

	return fmt.Errorf("unsupported type %v", params.type_)
}

func addToDataYAML(params newRuleCommandParams) error {
	dataFilePath := filepath.Join("bundle", "regal", "config", "provided", "data.yaml")

	yamlContent, err := os.ReadFile(dataFilePath)
	if err != nil {
		// Check if the error is of type *os.PathError
		var pathErr *os.PathError
		if errors.As(err, &pathErr) && os.IsNotExist(pathErr.Err) {
			// Handle the case where the file does not exist
			return errors.New("data.yaml file not found. " +
				"Please run this command from the top-level directory of the Regal repository")
		}

		return err
	}

	var existingConfig config.Config

	// Unmarshal the YAML content into a map
	if err := yaml.Unmarshal(yamlContent, &existingConfig); err != nil {
		return err
	}

	existingConfig.Capabilities = nil

	// Check if the category already exists in rules
	if existingConfig.Rules[params.category] == nil {
		existingConfig.Rules[params.category] = make(map[string]config.Rule)
	}

	// Assign a new rule/level value to the Category map
	existingConfig.Rules[params.category][params.name] = config.Rule{Level: "error"}

	// Sort the map keys alphabetically (categories)
	sortedCategories := make([]string, 0, len(existingConfig.Rules))
	for cat := range existingConfig.Rules {
		sortedCategories = append(sortedCategories, cat)
	}

	sort.Strings(sortedCategories)

	// Sort rule names within each category alphabetically
	for _, cat := range sortedCategories {
		sortedRuleNames := make([]string, 0, len(existingConfig.Rules[cat]))
		for ruleName := range existingConfig.Rules[cat] {
			sortedRuleNames = append(sortedRuleNames, ruleName)
		}

		sort.Strings(sortedRuleNames)

		sortedCategory := make(config.Category)
		for _, ruleName := range sortedRuleNames {
			sortedCategory[ruleName] = existingConfig.Rules[cat][ruleName]
		}

		existingConfig.Rules[cat] = sortedCategory
	}

	var b bytes.Buffer

	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)

	err = yamlEncoder.Encode(existingConfig)
	if err != nil {
		return err
	}

	// Write the YAML content to the file
	dataYamlDir := filepath.Join(params.output, "bundle", "regal", "config", "provided")
	if err := os.MkdirAll(dataYamlDir, 0o770); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dataYamlDir, "data.yaml"), b.Bytes(), 0o600)
}

func scaffoldCustomRule(params newRuleCommandParams) error {
	rulesDir := filepath.Join(
		params.output, ".regal", "rules", "custom", "regal", "rules", params.category, params.name,
	)

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
	rulesDir := filepath.Join(params.output, "bundle", "regal", "rules", params.category, params.name)

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

func createBuiltinDocs(params newRuleCommandParams) error {
	docsDir := filepath.Join(params.output, "docs", "rules", params.category)

	docTmpl := template.New("builtin.md.tpl")

	docTmpl = docTmpl.Funcs(template.FuncMap{
		"ToUpper": strings.ToUpper,
	})

	docTmpl, err := docTmpl.ParseFS(embeds.EmbedTemplatesFS, "templates/builtin/builtin.md.tpl")
	if err != nil {
		return err
	}

	docFileName := strings.ToLower(params.name) + ".md"

	docFile, err := os.Create(filepath.Join(docsDir, docFileName))
	if err != nil {
		return err
	}

	err = docTmpl.Execute(docFile, templateValues(params))
	if err != nil {
		return err
	}

	log.Printf("Created doc template for builtin rule %q in %s\n", params.name, docsDir)

	return nil
}

func templateValues(params newRuleCommandParams) TemplateValues {
	var tmplNameValue string

	var tmplNameTestValue string

	dashedNameValue := strings.ReplaceAll(params.name, "_", "-")

	switch {
	case strings.Contains(params.name, "-"):
		tmplNameValue = `["` + dashedNameValue + `"]`
		tmplNameTestValue = `["` + dashedNameValue + `_test"]`
	case strings.Contains(params.name, "_"):
		tmplNameValue = `["` + dashedNameValue + `"]`
		tmplNameTestValue = `["` + dashedNameValue + `_test"]`
	default:
		tmplNameValue = "." + params.name
		tmplNameTestValue = "." + params.name + "_test"
	}

	return TemplateValues{
		Category:     params.category,
		NameOriginal: params.name,
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
