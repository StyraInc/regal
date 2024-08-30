package embedded

import "testing"

func TestEmbeddedEOPA(t *testing.T) {
	t.Parallel()

	// As of 2024-08-27, there are 47 capabilities files in the EOPA repo.
	// It follows that there should never be less than 47 valid
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

	if len(versions) < 47 {
		t.Errorf("Expected at least 47 EOPA capabilities in the embedded database (got %d)", len(versions))
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
