package clients

// Identifier represent different supported clients and can be used to toggle or change
// server behavior based on the client.
type Identifier int

const (
	IdentifierGeneric Identifier = iota
	IdentifierVSCode
	IdentifierGoTest
	IdentifierZed
)

func DetermineClientIdentifier(clientName string) Identifier {
	switch clientName {
	case "go test":
		return IdentifierGoTest
	case "Visual Studio Code":
		return IdentifierVSCode
	case "Zed":
		return IdentifierZed
	}

	return IdentifierGeneric
}
