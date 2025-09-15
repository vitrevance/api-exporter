package parser

import (
	"encoding/json"
	"fmt"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type valueTransformer struct {
	Format any `default:"json" yaml:"format"`
}

func init() {
	transformer.RegisterTransformerFactory("parse", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		t := &valueTransformer{}
		err := value.Decode(t)
		return t, err
	}))
}

func (this *valueTransformer) Transform(ctx *transformer.TransformationContext) error {
	bytes, ok := ctx.Object.([]byte)
	if !ok {
		return fmt.Errorf("parser requires bytes")
	}

	switch this.Format {
	case "from_bytes":
		var obj map[string]any
		err := json.Unmarshal(bytes, &obj)
		if err != nil {
			var arr []any
			err = json.Unmarshal(bytes, &arr)
			if err != nil {
				return err
			}
			ctx.Result = arr
		}
		ctx.Result = obj
		break
	case "to_bytes":
		b, err := json.Marshal(ctx.Object)
		if err != nil {
			return err
		}
		ctx.Result = b
		break
	}
	return nil
}
