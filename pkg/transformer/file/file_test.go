package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

func TestFileTransformerWriteAndRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.bin")
	payload := []byte{0x00, 0x01, 0xfe, 0xff}

	write := decodeFileTransformer(t, "operation: write\npath: "+path)
	writeContext := &transformer.TransformationContext{
		Object: map[string]any{"body": payload},
	}
	require.NoError(t, write.Transform(writeContext))
	require.Equal(t, map[string]any{"body": payload}, writeContext.Result)

	written, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, payload, written)

	read := decodeFileTransformer(t, "operation: read\npath: "+path)
	readContext := &transformer.TransformationContext{Object: map[string]any{}}
	require.NoError(t, read.Transform(readContext))
	require.Equal(t, map[string]any{"body": payload}, readContext.Result)
}

func TestFileTransformerUsesRuntimeOptionsAndStringBody(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.txt")
	file := decodeFileTransformer(t, "operation: read\npath: unused")
	context := &transformer.TransformationContext{
		Object: map[string]any{
			"operation": "write",
			"path":      path,
			"body":      "hello",
		},
	}

	require.NoError(t, file.Transform(context))
	written, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), written)
}

func TestFileTransformerRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config string
		object any
		error  string
	}{
		{name: "missing operation", config: "path: file", object: map[string]any{}, error: "operation must be specified"},
		{name: "missing path", config: "operation: read", object: map[string]any{}, error: "path must be specified"},
		{name: "unknown operation", config: "operation: remove\npath: file", object: map[string]any{}, error: "unsupported file operation"},
		{name: "missing write body", config: "operation: write\npath: file", object: map[string]any{}, error: "file write requires a body"},
		{name: "invalid write body", config: "operation: write\npath: file", object: map[string]any{"body": 42}, error: "invalid body type int"},
		{name: "invalid runtime path", config: "operation: read\npath: file", object: map[string]any{"path": 42}, error: "invalid type for path"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := decodeFileTransformer(t, test.config)
			context := &transformer.TransformationContext{Object: test.object}
			require.ErrorContains(t, file.Transform(context), test.error)
		})
	}
}

func TestFileTransformerAcceptsDirectBytesForWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.bin")
	file := decodeFileTransformer(t, "operation: write\npath: "+path)
	context := &transformer.TransformationContext{Object: []byte("content")}

	require.NoError(t, file.Transform(context))
	written, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, []byte("content"), written)
}

func decodeFileTransformer(t *testing.T, config string) transformer.Transformer {
	t.Helper()

	var transformerConfig transformer.TransformerConfig
	require.NoError(t, yaml.Unmarshal([]byte("type: file\n"+config+"\n"), &transformerConfig))
	return transformerConfig.Transformer
}
