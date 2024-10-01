//go:build !race
// +build !race

package lsp

func isRaceEnabled() bool {
	return false
}
