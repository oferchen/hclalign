// internal/engine/process_reader_test.go
package engine_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/internal/diff"
	"github.com/oferchen/hclalign/internal/engine"
	internalfs "github.com/oferchen/hclalign/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestProcessReaderPreservesHints(t *testing.T) {
	t.Parallel()

	bom := string([]byte{0xEF, 0xBB, 0xBF})
	input := bom + "variable \"simple\" {\r\n  default = 1\r\n  type = number\r\n}"
	expected := bom + "variable \"simple\" {\r\n  type    = number\r\n  default = 1\r\n}"

	var out bytes.Buffer
	cfg := &config.Config{Mode: config.ModeWrite, Stdout: true}
	changed, err := engine.ProcessReader(context.Background(), strings.NewReader(input), &out, cfg)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, expected, out.String())

	hints := internalfs.DetectHintsFromBytes(out.Bytes())
	require.True(t, hints.HasBOM)
	require.Equal(t, "\r\n", hints.Newline)
}

func TestProcessReaderModeDiff(t *testing.T) {
	t.Parallel()

	bom := string([]byte{0xEF, 0xBB, 0xBF})
	input := bom + "variable \"simple\" {\r\n  default = 1\r\n  type = number\r\n}"
	styled := bom + "variable \"simple\" {\r\n  type    = number\r\n  default = 1\r\n}"

	var out bytes.Buffer
	cfg := &config.Config{Mode: config.ModeDiff}
	changed, err := engine.ProcessReader(context.Background(), strings.NewReader(input), &out, cfg)
	require.NoError(t, err)
	require.True(t, changed)

	hints := internalfs.DetectHintsFromBytes([]byte(input))
	diffText, err := diff.Unified("stdin", "stdin", []byte(input), []byte(styled), hints.Newline)
	require.NoError(t, err)
	require.Equal(t, diffText, out.String())
}

func TestProcessReaderModeCheckNoChange(t *testing.T) {
	t.Parallel()

	input := "variable \"simple\" {\n  type    = number\n  default = 1\n}"

	var out bytes.Buffer
	cfg := &config.Config{Mode: config.ModeCheck, Stdout: true}
	changed, err := engine.ProcessReader(context.Background(), strings.NewReader(input), &out, cfg)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, input, out.String())

	hints := internalfs.DetectHintsFromBytes([]byte(input))
	diffText, err := diff.Unified("stdin", "stdin", []byte(input), out.Bytes(), hints.Newline)
	require.NoError(t, err)
	require.Empty(t, diffText)
}
