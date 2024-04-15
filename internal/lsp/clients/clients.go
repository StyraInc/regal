package clients

// Identifier represent different supported clients and can be used to toggle or change
// server behavior based on the client.
type Identifier int

const (
	IdentifierGeneric Identifier = iota
	IdentifierVSCode
	IdentifierGoTest
)

func DetermineClientIdentifier(clientName string) Identifier {
	if clientName == "go test" {
		return IdentifierGoTest
	}

	if clientName == "Visual Studio Code" {
		return IdentifierVSCode
	}

	return IdentifierGeneric
}
