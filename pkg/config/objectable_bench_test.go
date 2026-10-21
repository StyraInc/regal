package config_test

import (
	"encoding/json/v2"
	"os"
	"testing"

	"github.com/open-policy-agent/regal/internal/encoding/noast"
	"github.com/open-policy-agent/regal/internal/test/assert"
	"github.com/open-policy-agent/regal/internal/test/must"
	"github.com/open-policy-agent/regal/internal/testutil"
	"github.com/open-policy-agent/regal/pkg/config"
	"github.com/open-policy-agent/regal/pkg/roast/rast"
)

func TestToValueVsRastStructToValue(t *testing.T) {
	t.Parallel()

	raw := must.Return(os.ReadFile("../../bundle/regal/config/provided/data.yaml"))(t)
	cfg := testutil.MustUnmarshalYAML[*config.Config](t, raw)

	v1 := cfg.ToValue()
	v2 := rast.StructToValue(cfg)

	if v1.Compare(v2) != 0 {
		v1bs, _ := json.Marshal(v1, noast.Options)
		v2bs, _ := json.Marshal(v2, noast.Options)

		assert.JSONEqual(t, v1bs, v2bs)
	}
}

// ToValue-16         	   105201 ns/op	  234964 B/op	    5323 allocs/op
// RastStructToValue-16    180497 ns/op	  307350 B/op	    7216 allocs/op
func BenchmarkToValueVsRastStructToValue(b *testing.B) {
	raw := must.Return(os.ReadFile("../../bundle/regal/config/provided/data.yaml"))(b)
	cfg := testutil.MustUnmarshalYAML[*config.Config](b, raw)

	b.Run("ToValue", func(b *testing.B) {
		for b.Loop() {
			_ = cfg.ToValue()
		}
	})

	b.Run("RastStructToValue", func(b *testing.B) {
		for b.Loop() {
			_ = rast.StructToValue(cfg)
		}
	})
}
