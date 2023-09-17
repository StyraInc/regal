// nolint:wrapcheck
package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

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

	// Call addToDataYAML for both custom and builtin rules
	if err := addToDataYAML(params); err != nil {
		return err
	}

	if params.type_ == "custom" {
		return scaffoldCustomRule(params)
	}

	if params.type_ == "builtin" {
		return scaffoldBuiltinRule(params)
	}

	return fmt.Errorf("unsupported type %v", params.type_)
}

func addToDataYAML(params newRuleCommandParams) error {
	// Open data.yaml for reading and writing
	dataFile, err := os.OpenFile("../bundle/regal/config/provided/data.yaml", os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer dataFile.Close()

	// Decode the existing data.yaml content into a Config struct
	var existingConfig config.Config

	decoder := yaml.NewDecoder(dataFile)
	if err := decoder.Decode(&existingConfig); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	// Add the new entry to the Rules map within the Config struct
	if existingConfig.Rules == nil {
		existingConfig.Rules = make(map[string]config.Category)
	}

	// Create a new Rule value and set the Level field
	vrule := config.Rule{
		Level: "error",
	}

	// Assign the new Rule value to the Category map
	existingConfig.Rules[params.category] = config.Category{
		params.name: vrule, // Assign the rule to a key within the Category map
	}

	// Sort the map keys alphabetically
	sortedKeys := make([]string, 0, len(existingConfig.Rules))
	for key := range existingConfig.Rules {
		sortedKeys = append(sortedKeys, key)
	}

	sort.Strings(sortedKeys)

	// Create a new map with sorted keys
	sortedRules := make(map[string]config.Category)
	for _, key := range sortedKeys {
		sortedRules[key] = existingConfig.Rules[key]
	}

	existingConfig.Rules = sortedRules

	// Write the updated Config struct back to data.yaml
	if _, err := dataFile.Seek(0, 0); err != nil {
		return err
	}

	if err := dataFile.Truncate(0); err != nil {
		return err
	}

	encoder := yaml.NewEncoder(dataFile)

	return encoder.Encode(existingConfig)
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
