package file

import (
	"fmt"
	"os"
	"strings"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

const (
	OperationRead  = "read"
	OperationWrite = "write"
)

// FileTargetConfig describes a filesystem operation.
type FileTargetConfig struct {
	Operation string `yaml:"operation"`
	Path      string `yaml:"path"`
}

// MergeMap applies operation and path overrides supplied by the current
// transformation object. Other fields, such as body, are intentionally
// ignored here and remain available as the operation's input.
func (c *FileTargetConfig) MergeMap(cfg map[string]any) error {
	if operation, exists := cfg["operation"]; exists {
		value, ok := operation.(string)
		if !ok {
			return fmt.Errorf("invalid type for operation, expected string")
		}
		c.Operation = value
	}

	if path, exists := cfg["path"]; exists {
		value, ok := path.(string)
		if !ok {
			return fmt.Errorf("invalid type for path, expected string")
		}
		c.Path = value
	}

	return nil
}

func init() {
	transformer.RegisterTransformerFactory("file", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		config := &FileTargetConfig{}
		if err := value.Decode(config); err != nil {
			return nil, err
		}
		return &fileTransformer{Config: config}, nil
	}))
}

type fileTransformer struct {
	Config *FileTargetConfig
}

func (t *fileTransformer) Transform(ctx *transformer.TransformationContext) error {
	// A transformer can be used concurrently by jobs and HTTP handlers. Copy
	// the decoded config before applying per-invocation overrides.
	config := *t.Config
	if source, ok := ctx.Object.(map[string]any); ok {
		if err := config.MergeMap(source); err != nil {
			return err
		}
	}

	config.Operation = strings.ToLower(strings.TrimSpace(config.Operation))
	if config.Path == "" {
		return fmt.Errorf("path must be specified")
	}

	switch config.Operation {
	case OperationRead:
		body, err := os.ReadFile(config.Path)
		if err != nil {
			return fmt.Errorf("read file %q: %w", config.Path, err)
		}
		ctx.Result = map[string]any{"body": body}
		return nil

	case OperationWrite:
		body, err := bodyFromObject(ctx.Object)
		if err != nil {
			return err
		}
		if err := os.WriteFile(config.Path, body, 0o644); err != nil {
			return fmt.Errorf("write file %q: %w", config.Path, err)
		}
		ctx.Result = map[string]any{"body": body}
		return nil

	case "":
		return fmt.Errorf("operation must be specified")
	default:
		return fmt.Errorf("unsupported file operation %q: expected %q or %q", config.Operation, OperationRead, OperationWrite)
	}
}

func bodyFromObject(object any) ([]byte, error) {
	if source, ok := object.(map[string]any); ok {
		body, exists := source["body"]
		if !exists {
			return nil, fmt.Errorf("file write requires a body")
		}
		object = body
	}

	switch body := object.(type) {
	case []byte:
		return body, nil
	case string:
		return []byte(body), nil
	default:
		return nil, fmt.Errorf("invalid body type %T, expected string or []byte", object)
	}
}
