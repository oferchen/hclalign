// internal/engine/engine_test.go
package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/internal/diff"
	internalfs "github.com/oferchen/hclalign/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestProcessMissingTarget(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Target:      "nonexistent.hcl",
		Include:     []string{"**/*.hcl"},
		Concurrency: 1,
	}

	changed, err := Process(context.Background(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
	require.False(t, changed)
}

func TestProcessContextCancelled(t *testing.T) {
	dir := t.TempDir()
	data, err := os.ReadFile(filepath.Join("testdata", "idempotent_input.hcl"))
	require.NoError(t, err)
	filePath := filepath.Join(dir, "example.hcl")
	require.NoError(t, os.WriteFile(filePath, data, 0o644))

	cfg := &config.Config{
		Target:      dir,
		Include:     []string{"**/*.hcl"},
		Concurrency: 1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	testHookAfterParse = func() { cancel() }
	defer func() { testHookAfterParse = nil }()

	changed, err := Process(ctx, cfg)
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, changed)
}

func TestProcessContextCancelledAfterReorder(t *testing.T) {
	dir := t.TempDir()
	data, err := os.ReadFile(filepath.Join("testdata", "idempotent_input.hcl"))
	require.NoError(t, err)
	filePath := filepath.Join(dir, "example.hcl")
	require.NoError(t, os.WriteFile(filePath, data, 0o644))

	cfg := &config.Config{
		Target:      dir,
		Include:     []string{"**/*.hcl"},
		Concurrency: 1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	testHookAfterReorder = func() { cancel() }
	defer func() { testHookAfterReorder = nil }()

	changed, err := Process(ctx, cfg)
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, changed)
}

func TestProcessScenarios(t *testing.T) {
	casesDir := filepath.Join("..", "..", "tests", "cases")
	tests := []struct {
		name  string
		setup func(t *testing.T) (*config.Config, string, bool, map[string]string)
	}{
		{
			name: "stdout multiple files",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				dir := t.TempDir()
				in1, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
				require.NoError(t, err)
				out1, err := os.ReadFile(filepath.Join(casesDir, "simple", "out.tf"))
				require.NoError(t, err)
				f1 := filepath.Join(dir, "a.tf")
				require.NoError(t, os.WriteFile(f1, in1, 0o644))

				in2, err := os.ReadFile(filepath.Join(casesDir, "trailing_commas", "in.tf"))
				require.NoError(t, err)
				out2, err := os.ReadFile(filepath.Join(casesDir, "trailing_commas", "out.tf"))
				require.NoError(t, err)
				f2 := filepath.Join(dir, "b.tf")
				require.NoError(t, os.WriteFile(f2, in2, 0o644))

				cfg := &config.Config{
					Target:      dir,
					Include:     []string{"**/*.tf"},
					Mode:        config.ModeWrite,
					Stdout:      true,
					Concurrency: 1,
				}

				expOut := fmt.Sprintf("\n--- %s ---\n%s\n--- %s ---\n%s", f1, out1, f2, out2)
				files := map[string]string{f1: string(out1), f2: string(out2)}
				return cfg, expOut, true, files
			},
		},
		{
			name: "mode diff",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				dir := t.TempDir()
				inPath := filepath.Join(casesDir, "simple", "in.tf")
				outPath := filepath.Join(casesDir, "simple", "out.tf")
				inb, err := os.ReadFile(inPath)
				require.NoError(t, err)
				outb, err := os.ReadFile(outPath)
				require.NoError(t, err)
				f := filepath.Join(dir, "diff.tf")
				require.NoError(t, os.WriteFile(f, inb, 0o644))

				hints := internalfs.DetectHintsFromBytes(inb)
				diffText, err := diff.Unified(diff.UnifiedOpts{FromFile: f, ToFile: f, Original: inb, Styled: outb, Hints: hints})
				require.NoError(t, err)

				cfg := &config.Config{
					Target:      f,
					Include:     []string{"**/*.tf"},
					Mode:        config.ModeDiff,
					Stdout:      true,
					Concurrency: 1,
				}
				files := map[string]string{f: string(inb)}
				return cfg, fmt.Sprintf("\n--- %s ---\n%s", f, diffText), true, files
			},
		},
		{
			name: "mode check",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				dir := t.TempDir()
				inb, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
				require.NoError(t, err)
				outb, err := os.ReadFile(filepath.Join(casesDir, "simple", "out.tf"))
				require.NoError(t, err)
				f := filepath.Join(dir, "check.tf")
				require.NoError(t, os.WriteFile(f, inb, 0o644))

				cfg := &config.Config{
					Target:      f,
					Include:     []string{"**/*.tf"},
					Mode:        config.ModeCheck,
					Stdout:      true,
					Concurrency: 1,
				}
				files := map[string]string{f: string(inb)}
				return cfg, fmt.Sprintf("\n--- %s ---\n%s", f, outb), true, files
			},
		},
		{
			name: "symlink follow",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				base := t.TempDir()
				target := t.TempDir()
				inb, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
				require.NoError(t, err)
				outb, err := os.ReadFile(filepath.Join(casesDir, "simple", "out.tf"))
				require.NoError(t, err)
				realFile := filepath.Join(target, "file.tf")
				require.NoError(t, os.WriteFile(realFile, inb, 0o644))
				link := filepath.Join(base, "link")
				require.NoError(t, os.Symlink(target, link))

				cfg := &config.Config{
					Target:         base,
					Include:        []string{"**/*.tf"},
					Mode:           config.ModeWrite,
					Stdout:         true,
					Concurrency:    1,
					FollowSymlinks: true,
				}
				files := map[string]string{realFile: string(outb)}
				return cfg, fmt.Sprintf("\n--- %s ---\n%s", filepath.Join(link, "file.tf"), outb), true, files
			},
		},
		{
			name: "symlink no follow",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				base := t.TempDir()
				target := t.TempDir()
				inb, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
				require.NoError(t, err)
				realFile := filepath.Join(target, "file.tf")
				require.NoError(t, os.WriteFile(realFile, inb, 0o644))
				link := filepath.Join(base, "link")
				require.NoError(t, os.Symlink(target, link))

				cfg := &config.Config{
					Target:         base,
					Include:        []string{"**/*.tf"},
					Mode:           config.ModeWrite,
					Stdout:         true,
					Concurrency:    1,
					FollowSymlinks: false,
				}
				files := map[string]string{realFile: string(inb)}
				return cfg, "", false, files
			},
		},
		{
			name: "symlink file follow",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				base := t.TempDir()
				realDir := t.TempDir()
				inb, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
				require.NoError(t, err)
				outb, err := os.ReadFile(filepath.Join(casesDir, "simple", "out.tf"))
				require.NoError(t, err)
				realFile := filepath.Join(realDir, "file.tf")
				require.NoError(t, os.WriteFile(realFile, inb, 0o644))
				link := filepath.Join(base, "link.tf")
				require.NoError(t, os.Symlink(realFile, link))

				cfg := &config.Config{
					Target:         base,
					Include:        []string{"**/*.tf"},
					Mode:           config.ModeWrite,
					Stdout:         true,
					Concurrency:    1,
					FollowSymlinks: true,
				}
				files := map[string]string{link: string(outb), realFile: string(inb)}
				return cfg, fmt.Sprintf("\n--- %s ---\n%s", link, outb), true, files
			},
		},
		{
			name: "symlink file no follow",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				base := t.TempDir()
				realDir := t.TempDir()
				inb, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
				require.NoError(t, err)
				realFile := filepath.Join(realDir, "file.tf")
				require.NoError(t, os.WriteFile(realFile, inb, 0o644))
				link := filepath.Join(base, "link.tf")
				require.NoError(t, os.Symlink(realFile, link))

				cfg := &config.Config{
					Target:         base,
					Include:        []string{"**/*.tf"},
					Mode:           config.ModeWrite,
					Stdout:         true,
					Concurrency:    1,
					FollowSymlinks: false,
				}
				files := map[string]string{link: string(inb), realFile: string(inb)}
				return cfg, "", false, files
			},
		},
		{
			name: "target symlink dir follow",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				target := t.TempDir()
				inb, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
				require.NoError(t, err)
				outb, err := os.ReadFile(filepath.Join(casesDir, "simple", "out.tf"))
				require.NoError(t, err)
				realFile := filepath.Join(target, "file.tf")
				require.NoError(t, os.WriteFile(realFile, inb, 0o644))
				linkDir := t.TempDir()
				link := filepath.Join(linkDir, "link")
				require.NoError(t, os.Symlink(target, link))

				cfg := &config.Config{
					Target:         link,
					Include:        []string{"**/*.tf"},
					Mode:           config.ModeWrite,
					Stdout:         true,
					Concurrency:    1,
					FollowSymlinks: true,
				}
				files := map[string]string{realFile: string(outb)}
				return cfg, fmt.Sprintf("\n--- %s ---\n%s", filepath.Join(link, "file.tf"), outb), true, files
			},
		},
		{
			name: "target symlink dir no follow",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				target := t.TempDir()
				inb, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
				require.NoError(t, err)
				realFile := filepath.Join(target, "file.tf")
				require.NoError(t, os.WriteFile(realFile, inb, 0o644))
				linkDir := t.TempDir()
				link := filepath.Join(linkDir, "link")
				require.NoError(t, os.Symlink(target, link))

				cfg := &config.Config{
					Target:         link,
					Include:        []string{"**/*.tf"},
					Mode:           config.ModeWrite,
					Stdout:         true,
					Concurrency:    1,
					FollowSymlinks: false,
				}
				files := map[string]string{realFile: string(inb)}
				return cfg, "", false, files
			},
		},
		{
			name: "target symlink file follow",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				dir := t.TempDir()
				inb, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
				require.NoError(t, err)
				outb, err := os.ReadFile(filepath.Join(casesDir, "simple", "out.tf"))
				require.NoError(t, err)
				realFile := filepath.Join(dir, "real.tf")
				require.NoError(t, os.WriteFile(realFile, inb, 0o644))
				link := filepath.Join(dir, "link.tf")
				require.NoError(t, os.Symlink(realFile, link))

				cfg := &config.Config{
					Target:         link,
					Include:        []string{"**/*.tf"},
					Mode:           config.ModeWrite,
					Stdout:         true,
					Concurrency:    1,
					FollowSymlinks: true,
				}
				files := map[string]string{link: string(outb)}
				return cfg, fmt.Sprintf("\n--- %s ---\n%s", link, outb), true, files
			},
		},
		{
			name: "target symlink file no follow",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				dir := t.TempDir()
				inb, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
				require.NoError(t, err)
				realFile := filepath.Join(dir, "real.tf")
				require.NoError(t, os.WriteFile(realFile, inb, 0o644))
				link := filepath.Join(dir, "link.tf")
				require.NoError(t, os.Symlink(realFile, link))

				cfg := &config.Config{
					Target:         link,
					Include:        []string{"**/*.tf"},
					Mode:           config.ModeWrite,
					Stdout:         true,
					Concurrency:    1,
					FollowSymlinks: false,
				}
				files := map[string]string{link: string(inb), realFile: string(inb)}
				return cfg, "", false, files
			},
		},
		{
			name: "concurrency",
			setup: func(t *testing.T) (*config.Config, string, bool, map[string]string) {
				dir := t.TempDir()
				cases := []string{"simple", "trailing_commas", "comments", "heredocs"}
				names := []string{"a.tf", "b.tf", "c.tf", "d.tf"}
				files := make(map[string]string)
				var expected string
				for i, c := range cases {
					inb, err := os.ReadFile(filepath.Join(casesDir, c, "in.tf"))
					require.NoError(t, err)
					outb, err := os.ReadFile(filepath.Join(casesDir, c, "out.tf"))
					require.NoError(t, err)
					p := filepath.Join(dir, names[i])
					require.NoError(t, os.WriteFile(p, inb, 0o644))
					files[p] = string(outb)
					expected += fmt.Sprintf("\n--- %s ---\n%s", p, outb)
				}
				cfg := &config.Config{
					Target:      dir,
					Include:     []string{"**/*.tf"},
					Mode:        config.ModeWrite,
					Stdout:      true,
					Concurrency: 2,
				}
				return cfg, expected, true, files
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, wantStdout, wantChanged, files := tt.setup(t)
			r, w, err := os.Pipe()
			require.NoError(t, err)
			stdout := os.Stdout
			os.Stdout = w
			t.Cleanup(func() { os.Stdout = stdout })
			changed, err := Process(context.Background(), cfg)
			require.NoError(t, err)
			require.Equal(t, wantChanged, changed)
			w.Close()
			out, err := io.ReadAll(r)
			require.NoError(t, err)
			require.Equal(t, wantStdout, string(out))
			for p, exp := range files {
				got, err := os.ReadFile(p)
				require.NoError(t, err)
				require.Equal(t, strings.TrimSuffix(exp, "\n"), strings.TrimSuffix(string(got), "\n"))
			}
		})
	}
}

func TestProcessManyFilesDeterministic(t *testing.T) {
	t.Skip("skipping until deterministic formatting is restored")

	casesDir := filepath.Join("..", "..", "tests", "cases")
	caseDirs := []string{"simple", "trailing_commas", "comments", "complex", "whitespace"}

	const numFiles = 100
	dir := t.TempDir()
	files := make(map[string]string, numFiles)
	var expectedChanged bool
	for i := 0; i < numFiles; i++ {
		c := caseDirs[i%len(caseDirs)]
		inPath := filepath.Join(casesDir, c, "in.tf")
		outPath := filepath.Join(casesDir, c, "out.tf")
		inb, err := os.ReadFile(inPath)
		require.NoError(t, err)
		outb, err := os.ReadFile(outPath)
		require.NoError(t, err)
		p := filepath.Join(dir, fmt.Sprintf("file%03d.tf", i))
		require.NoError(t, os.WriteFile(p, inb, 0o644))
		files[p] = string(outb)
		if !bytes.Equal(inb, outb) {
			expectedChanged = true
		}
	}

	cfg := &config.Config{
		Target:      dir,
		Include:     []string{"**/*.tf"},
		Mode:        config.ModeWrite,
		Stdout:      true,
		Concurrency: 4,
	}

	r, w, err := os.Pipe()
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = stdout })

	changed, err := Process(context.Background(), cfg)
	require.NoError(t, err)
	require.Equal(t, expectedChanged, changed)

	w.Close()
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var wantStdout strings.Builder
	for _, p := range paths {
		b, err := os.ReadFile(p)
		require.NoError(t, err)
		fmt.Fprintf(&wantStdout, "\n--- %s ---\n", p)
		wantStdout.Write(b)
	}
	require.Equal(t, wantStdout.String(), string(out))

	for p, exp := range files {
		got, err := os.ReadFile(p)
		require.NoError(t, err)
		require.Equal(t, exp, string(got))
	}
}
