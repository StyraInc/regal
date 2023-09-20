// nolint:wrapcheck
package cmd

import (
	"bytes"
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

	if err := addRuleToREADME(params); err != nil {
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
	// Read the YAML content from the file
	yamlContent, err := os.ReadFile("bundle/regal/config/provided/data.yaml")
	if err != nil {
		return err
	}

	var existingConfig config.Config

	// Unmarshal the YAML content into a map
	if err := yaml.Unmarshal(yamlContent, &existingConfig); err != nil {
		return err
	}

	// Add the new entry to the Rules map within the Config struct
	if existingConfig.Rules == nil {
		existingConfig.Rules = make(map[string]config.Category)
	}

	// Create a new Rule value and set the Level field
	vrule := config.Rule{
		Level: "error",
	}

	// Check if the category already exists in rules
	if existingConfig.Rules[params.category] == nil {
		existingConfig.Rules[params.category] = make(map[string]config.Rule)
	}

	// Assign the new Rule value to the Category map
	existingConfig.Rules[params.category][params.name] = vrule

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
	return os.WriteFile("bundle/regal/config/provided/data.yaml", b.Bytes(), 0o600)
}

func addRuleToREADME(params newRuleCommandParams) error {
	var sortedRules string

	readmePath := "README.md"
	// Read the existing README.md content
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	// Define the start and end patterns for the table
	startPattern := "|-----------" +
		"|---------------------------------------------------------------------------------------------------" +
		"|-----------------------------------------------------------|"
	endPattern := "\n\n<!-- RULES_TABLE_END -->"

	// Find the position to insert the new rule entry
	startIndex := strings.Index(string(readmeContent), startPattern)
	endIndex := strings.Index(string(readmeContent), endPattern)

	if startIndex == -1 || endIndex == -1 {
		return fmt.Errorf("failed to locate start or end pattern in README.md")
	}

	// Extract the content before and after the table
	beforeTable := string(readmeContent[:startIndex+len(startPattern)])
	afterTable := string(readmeContent[endIndex:])

	// Extract the existing rule entries from the table
	existingRules := string(readmeContent[startIndex+len(startPattern) : endIndex])

	// Define the new rule entry
	newRule := fmt.Sprintf("| %-10s| [%-97s| %-58s|",
		params.category, fmt.Sprintf("%s](https://docs.styra.com/regal/rules/%s/%s)",
			params.name, params.category, params.name), "Place holder, description of the new rule")

	// Create a regular expression pattern to match lines starting with "| %-10s|"
	pattern := fmt.Sprintf(`\| %-10s\| \[%s\].*\n`, regexp.QuoteMeta(params.category), regexp.QuoteMeta(params.name))

	// Compile the regular expression
	re := regexp.MustCompile(pattern)
	// Check if there's a match
	if re.MatchString(existingRules) {
		// Replace matching lines with the new rule
		existingRules = re.ReplaceAllString(existingRules, newRule+"\n")
		sortedRules = existingRules
	} else {
		// Combine the existing rules with the new rule and sort them
		allRules := existingRules + "\n" + newRule
		sortedRules = sortRulesTable(allRules)
	}

	// Create the updated content
	newContent := beforeTable + sortedRules + afterTable

	// Write the updated content back to README.md
	err = os.WriteFile(readmePath, []byte(newContent), 0o600)
	if err != nil {
		return err
	}

	return nil
}

func sortRulesTable(rulesTable string) string {
	// Split the table into lines
	lines := strings.Split(rulesTable, "\n")

	// Sort the lines (excluding the header line)
	if len(lines) > 2 {
		sort.Strings(lines[2:])
	}

	// Join the sorted lines back together
	return strings.Join(lines, "\n")
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
