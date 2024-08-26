package embedded

import "testing"

func TestEmbeddedEOPA(t *testing.T) {
	// As of 2024-08-23, there are 57 capabilities files in the EOPA repo.
	// It follows that there should never be less than 54 valid
	// capabilities in the embedded database. This is really just a sanity
	// check to ensure the JSON files didn't get misplaced or something to
	// that effect.
	//
	// This also ensures that all of the embedded capabilities files are
	// valid JSON we can successfully marshal into *ast.Capabilities.

	versions, err := LoadCapabilitiesVersions("eopa")
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) < 54 {
		t.Errorf("Expected at least 54 EOPA capabilities in the embedded database")
	}

	for _, v := range versions {
		caps, err := LoadCapabilitiesVersion("eopa", v)

		if err != nil {
			t.Errorf("error with eopa capabilities version %s: %v", v, err)
		}

		if len(caps.Builtins) < 1 {
			t.Errorf("eopa capabilities version %s has no builtins", v)
		}
	}

}
