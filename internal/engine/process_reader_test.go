// /internal/engine/process_reader_test.go
package engine_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	diffText, err := diff.Unified("stdin", "stdin", []byte(input), []byte(styled), hints)
	require.NoError(t, err)
	require.Equal(t, diffText, out.String())
}

func TestProcessReaderModeDiffNoChange(t *testing.T) {
	t.Parallel()

	input := "variable \"simple\" {\n  type    = number\n  default = 1\n}"

	var out bytes.Buffer
	cfg := &config.Config{Mode: config.ModeDiff}
	changed, err := engine.ProcessReader(context.Background(), strings.NewReader(input), &out, cfg)
	require.NoError(t, err)
	require.False(t, changed)
	require.Empty(t, out.String())
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
	diffText, err := diff.Unified("stdin", "stdin", []byte(input), out.Bytes(), hints)
	require.NoError(t, err)
	require.Empty(t, diffText)
}

func TestProcessPrintsDelimiters(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.tf")
	f2 := filepath.Join(dir, "b.tf")
	require.NoError(t, os.WriteFile(f1, []byte("variable \"a\" {}\n"), 0o644))
	require.NoError(t, os.WriteFile(f2, []byte("variable \"b\" {}\n"), 0o644))

	cfg := &config.Config{
		Target:      dir,
		Include:     []string{"**/*.tf"},
		Mode:        config.ModeCheck,
		Stdout:      true,
		Concurrency: 1,
	}

	r, w, err := os.Pipe()
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = stdout })

	_, err = engine.Process(context.Background(), cfg)
	require.NoError(t, err)
	w.Close()
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	s := string(out)
	require.Contains(t, s, fmt.Sprintf("\n--- %s ---\n", f1))
	require.Contains(t, s, fmt.Sprintf("\n--- %s ---\n", f2))
}

func TestProcessStdinStdoutPreservesHints(t *testing.T) {
	t.Parallel()

	bom := string([]byte{0xEF, 0xBB, 0xBF})
	input := bom + "variable \"simple\" {\r\n  default = 1\r\n  type = number\r\n}"
	expected := bom + "variable \"simple\" {\r\n  type    = number\r\n  default = 1\r\n}"

	cfg := &config.Config{Mode: config.ModeWrite, Stdin: true, Stdout: true}

	run := func(in string) (string, bool) {
		inR, inW, err := os.Pipe()
		require.NoError(t, err)
		outR, outW, err := os.Pipe()
		require.NoError(t, err)
		defer inR.Close()
		defer outR.Close()

		oldIn, oldOut := os.Stdin, os.Stdout
		os.Stdin, os.Stdout = inR, outW
		defer func() { os.Stdin, os.Stdout = oldIn, oldOut }()

		go func() {
			_, _ = inW.Write([]byte(in))
			inW.Close()
		}()

		changed, err := engine.Process(context.Background(), cfg)
		require.NoError(t, err)

		outW.Close()
		b, err := io.ReadAll(outR)
		require.NoError(t, err)
		return string(b), changed
	}

	out1, changed1 := run(input)
	require.True(t, changed1)
	require.Equal(t, expected, out1)

	out2, changed2 := run(out1)
	require.False(t, changed2)
	require.Equal(t, out1, out2)

	hints := internalfs.DetectHintsFromBytes([]byte(out2))
	require.True(t, hints.HasBOM)
	require.Equal(t, "\r\n", hints.Newline)
}
