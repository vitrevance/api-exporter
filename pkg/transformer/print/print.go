package print

import (
	"fmt"
	"log"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type printTransformer struct {
	Format string `yaml:"format"`
	Log    bool   `yaml:"log"`
}

func init() {
	transformer.RegisterTransformerFactory("print", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		t := &printTransformer{}
		err := value.Decode(t)
		if t.Format == "" {
			t.Format = "%v"
		}
		return t, err
	}))
}

func (this *printTransformer) Transform(ctx *transformer.TransformationContext) error {
	str := fmt.Sprintf(this.Format, ctx.Object)
	if this.Log {
		log.Println(str)
	}
	ctx.Result = str
	return nil
}
