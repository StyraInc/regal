//go:build !regal_standalone

package mode

// Standalone lets us change the output of some commands when Regal
// is used as a binary, as opposed to when it's embedded via its
// Go module.
const Standalone = false
