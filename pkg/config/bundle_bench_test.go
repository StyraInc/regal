package config_test

import (
	"testing"

	rbundle "github.com/open-policy-agent/regal/bundle"
	"github.com/open-policy-agent/regal/pkg/config"
)

func BenchmarkWithDefaultsFromBundle(b *testing.B) {
	embedded := rbundle.Embedded()

	b.Run("only embedded", func(b *testing.B) {
		for b.Loop() {
			_, _ = config.WithDefaultsFromBundle(embedded, nil)
		}
	})

	b.Run("with empty user config", func(b *testing.B) {
		userConfig := &config.Config{}
		for b.Loop() {
			_, _ = config.WithDefaultsFromBundle(embedded, userConfig)
		}
	})
}
