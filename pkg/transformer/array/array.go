package json

import (
	"fmt"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type itemsTransformer struct {
	Map transformer.TransformerConfig `yaml:"map"`
}

func init() {
	transformer.RegisterTransformerFactory("array", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		t := &itemsTransformer{}
		err := value.Decode(t)
		if err != nil {
			return nil, err
		}
		return t, err
	}))
}

func (this *itemsTransformer) Transform(ctx *transformer.TransformationContext) error {
	src, ok := ctx.Object.([]any)
	if !ok {
		return fmt.Errorf("invalid array object")
	}

	var result []any
	if arr, ok := ctx.Result.([]any); ok {
		result = arr
	} else {
		result = make([]any, 0)
	}

	for _, elem := range src {
		mapperCtx := &transformer.TransformationContext{
			Object:       elem,
			Result:       make(map[string]any),
			Transformers: ctx.Transformers,
		}
		err := this.Map.Transformer.Transform(mapperCtx)
		if err != nil {
			return err
		}
		result = append(result, mapperCtx.Result)
	}
	ctx.Result = result
	return nil
}
