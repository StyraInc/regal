package clients

// Identifier represent different supported clients and can be used to toggle or change
// server behavior based on the client.
type Identifier int

const (
	IdentifierGeneric Identifier = iota
	IdentifierVSCode
)

func DetermineClientIdentifier(clientName string) Identifier {
	if clientName == "Visual Studio Code" {
		return IdentifierVSCode
	}

	return IdentifierGeneric
}
