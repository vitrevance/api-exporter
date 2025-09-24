package field

import (
	"fmt"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type jsonTransformer struct {
	Source *string                        `yaml:"source"`
	Target *string                        `yaml:"target"`
	Map    *transformer.TransformerConfig `yaml:"map"`
}

func init() {
	transformer.RegisterTransformerFactory("field", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		t := &jsonTransformer{}
		err := value.Decode(t)
		if err != nil {
			return nil, err
		}
		return t, err
	}))
}

func (this *jsonTransformer) Transform(ctx *transformer.TransformationContext) error {
	var src any = ctx.Object
	if this.Source != nil {
		json, ok := ctx.Object.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid json object")
		}
		src, ok = json[*this.Source]
		if !ok {
			return fmt.Errorf("json object has no field %v", *this.Source)
		}
	}

	var target any = ctx.Result
	if this.Target != nil {
		parent, ok := ctx.Result.(map[string]any)
		if !ok {
			return fmt.Errorf("parent object of field %v is present but is not a JSON map", *this.Target)
		}
		target, ok = parent[*this.Target]
		if !ok {
			target = make(map[string]any)
		}
	}

	if this.Map != nil {
		mapperCtx := &transformer.TransformationContext{
			Object:       src,
			Result:       target,
			Transformers: ctx.Transformers,
		}
		err := this.Map.Transformer.Transform(mapperCtx)
		if err != nil {
			return err
		}
		src = mapperCtx.Result
	}

	if this.Target == nil {
		ctx.Result = src
	} else {
		parent := ctx.Result.(map[string]any)
		parent[*this.Target] = src
	}

	return nil
}
