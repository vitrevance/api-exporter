package regex

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type regexTransformer struct {
	Match       string `yaml:"match"`
	Replacement string `yaml:"replacement"`
}

func init() {
	transformer.RegisterTransformerFactory("regex", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		t := &regexTransformer{}
		err := value.Decode(t)
		if err != nil {
			return nil, err
		}
		_, err = regexp.Compile(t.Match)
		if err != nil {
			return nil, err
		}
		return t, nil
	}))
}

func (this *regexTransformer) Transform(ctx *transformer.TransformationContext) error {
	str, ok := ctx.Object.(string)
	if !ok {
		return fmt.Errorf("regex is only applicable to strings")
	}

	re, err := regexp.Compile(this.Match)
	if err != nil {
		return fmt.Errorf("failed to compile regex: %w", err)
	}

	matches := re.FindStringSubmatch(str)
	if matches == nil {
		ctx.Result = str
		return nil
	}

	result := this.Replacement
	for i := 1; i < len(matches); i++ {
		placeholder := fmt.Sprintf("$%d", i)
		result = strings.ReplaceAll(result, placeholder, matches[i])
	}

	ctx.Result = result
	return nil
}
