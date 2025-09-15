package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"

	_ "github.com/vitrevance/api-exporter/pkg/transformer/array"
	_ "github.com/vitrevance/api-exporter/pkg/transformer/field"
)

const config = `
f2f:
  type: field
  source: keyname
  target: namekey
f2fm:
  type: field
  source: keyname
  target: namekey
  map:
    type: field
    source: subkey
arr:
  type: field
  source: items
  target: items
  map:
    type: array
    map:
      type: field
      target: url
`

func TestField(t *testing.T) {
	ts := make(map[string]transformer.TransformerConfig)
	require.NoError(t, yaml.Unmarshal([]byte(config), &ts))

	{
		ctx := &transformer.TransformationContext{
			Object: map[string]any{
				"keyname": "value",
			},
			Result:       make(map[string]any),
			Transformers: nil,
		}
		require.NoError(t, ts["f2f"].Transformer.Transform(ctx))
		require.NotNil(t, ctx.Result.(map[string]any)["namekey"])
		require.EqualValues(t, ctx.Result.(map[string]any)["namekey"], "value")
	}
	{
		ctx := &transformer.TransformationContext{
			Object: map[string]any{
				"keyname": map[string]any{
					"subkey": "subvalue",
				},
			},
			Result:       make(map[string]any),
			Transformers: nil,
		}
		require.NoError(t, ts["f2fm"].Transformer.Transform(ctx))
		require.NotNil(t, ctx.Result.(map[string]any)["namekey"])
		require.EqualValues(t, ctx.Result.(map[string]any)["namekey"], "subvalue")
	}
}

func TestArray(t *testing.T) {
	ts := make(map[string]transformer.TransformerConfig)
	require.NoError(t, yaml.Unmarshal([]byte(config), &ts))

	{
		ctx := &transformer.TransformationContext{
			Object: map[string]any{
				"items": []any{
					"1",
					"2",
					"3",
				},
			},
			Result:       make(map[string]any),
			Transformers: nil,
		}
		require.NoError(t, ts["arr"].Transformer.Transform(ctx))
		require.NotNil(t, ctx.Result.(map[string]any)["items"])
		require.EqualValues(t, []any{
			map[string]any{"url": "1"},
			map[string]any{"url": "2"},
			map[string]any{"url": "3"},
		}, ctx.Result.(map[string]any)["items"])
	}
}
