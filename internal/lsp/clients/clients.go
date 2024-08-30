package clients

// Identifier represent different supported clients and can be used to toggle or change
// server behavior based on the client.
type Identifier int

const (
	IdentifierGeneric Identifier = iota
	IdentifierVSCode
	IdentifierGoTest
	IdentifierZed
	IdentifierNeovim
)

// DetermineClientIdentifier is used to determine the Regal client identifier
// based on the client name.
// Clients with identifiers here should be featured on the 'Editor Support'
// page in the documentation (https://docs.styra.com/regal/editor-support).
func DetermineClientIdentifier(clientName string) Identifier {
	switch clientName {
	case "go test":
		return IdentifierGoTest
	case "Visual Studio Code":
		return IdentifierVSCode
	case "Zed":
		return IdentifierZed
	case "Neovim":
		// 'Neovim' is sent as the client identifier when using the
		// nvim-lspconfig plugin.
		return IdentifierNeovim
	}

	return IdentifierGeneric
}
