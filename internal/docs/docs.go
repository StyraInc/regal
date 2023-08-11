package docs

const docsBaseURL = "https://docs.styra.com/regal/rules"

// CreateDocsURL creates a complete URL to the documentation for a rule.
func CreateDocsURL(category, title string) string {
	return docsBaseURL + "/" + category + "/" + title
}
