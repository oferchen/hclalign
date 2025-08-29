// internal/engine/engine.go
package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/internal/diff"
	internalfs "github.com/oferchen/hclalign/internal/fs"
	"github.com/oferchen/hclalign/internal/hclalign"
	"github.com/oferchen/hclalign/patternmatching"
)

var (
	testHookAfterParse   func()
	testHookAfterReorder func()
	reorderAttributes    = hclalign.ReorderAttributes
	WriteFileAtomic      = internalfs.WriteFileAtomic
)

func Process(ctx context.Context, cfg *config.Config) (bool, error) {
	if cfg.Stdin {
		return processReader(ctx, os.Stdin, os.Stdout, cfg)
	}
	return processFiles(ctx, cfg)
}

func processFiles(ctx context.Context, cfg *config.Config) (bool, error) {
	if _, err := os.Stat(cfg.Target); err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("target %q does not exist", cfg.Target)
		}
		return false, err
	}
	matcher, err := patternmatching.NewMatcher(cfg.Include, cfg.Exclude)
	if err != nil {
		return false, err
	}
	var files []string
	var walk func(context.Context, string) error
	walk = func(ctx context.Context, dir string) error {
		if !matcher.Matches(dir) {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := ctx.Err(); err != nil {
				return err
			}
			path := filepath.Join(dir, entry.Name())
			if entry.Type()&os.ModeSymlink != 0 {
				info, err := os.Stat(path)
				if err != nil {
					return err
				}
				if info.IsDir() {
					if cfg.FollowSymlinks {
						if err := walk(ctx, path); err != nil {
							return err
						}
					}
					continue
				}
			}
			if entry.IsDir() {
				if err := walk(ctx, path); err != nil {
					return err
				}
				continue
			}
			if matcher.Matches(path) {
				files = append(files, path)
			}
		}
		return nil
	}
	info, err := os.Lstat(cfg.Target)
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		resolved, err := os.Stat(cfg.Target)
		if err != nil {
			return false, err
		}
		if resolved.IsDir() {
			if cfg.FollowSymlinks {
				if err := walk(ctx, cfg.Target); err != nil {
					return false, err
				}
			}
		} else if matcher.Matches(cfg.Target) {
			files = append(files, cfg.Target)
		}
	} else if info.IsDir() {
		if err := walk(ctx, cfg.Target); err != nil {
			return false, err
		}
	} else {
		if matcher.Matches(cfg.Target) {
			files = append(files, cfg.Target)
		}
	}
	sort.Strings(files)

	type result struct {
		path string
		data []byte
	}

	var changed atomic.Bool
	outs := make(map[string][]byte, len(files))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fileCh := make(chan string)
	results := make(chan result, len(files))

	var wg sync.WaitGroup
	var errMu sync.Mutex
	var errs []error

	go func() {
		defer close(fileCh)
		for _, f := range files {
			select {
			case <-ctx.Done():
				return
			case fileCh <- f:
			}
		}
	}()

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case f, ok := <-fileCh:
					if !ok {
						return
					}
					ch, out, err := processSingleFile(ctx, f, cfg)
					if err != nil {
						if errors.Is(err, context.Canceled) {
							cancel()
							return
						}
						log.Printf("error processing file %s: %v", f, err)
						errMu.Lock()
						errs = append(errs, fmt.Errorf("%s: %w", f, err))
						errMu.Unlock()
						continue
					}
					if ch {
						changed.Store(true)
					}
					if len(out) > 0 {
						select {
						case results <- result{path: f, data: out}:
						case <-ctx.Done():
							return
						}
					}
					if cfg.Verbose {
						log.Printf("processed file: %s", f)
					}
				}
			}
		}()
	}

	wg.Wait()
	if err := ctx.Err(); err != nil {
		return changed.Load(), err
	}
	close(results)
	for r := range results {
		outs[r.path] = r.data
	}

	for _, f := range files {
		if out, ok := outs[f]; ok && len(out) > 0 {
			if err := ctx.Err(); err != nil {
				if !errors.Is(err, context.Canceled) {
					return changed.Load(), err
				}
				break
			}
			if _, err := os.Stdout.Write(out); err != nil {
				return changed.Load(), err
			}
		}
	}

	if len(errs) > 0 {
		messages := make([]string, len(errs))
		for i, err := range errs {
			messages[i] = err.Error()
		}
		return changed.Load(), errors.New(strings.Join(messages, "\n"))
	}

	return changed.Load(), nil
}

func processSingleFile(ctx context.Context, filePath string, cfg *config.Config) (bool, []byte, error) {
	if err := ctx.Err(); err != nil {
		return false, nil, err
	}
	data, perm, hints, err := internalfs.ReadFileWithHints(ctx, filePath)
	if err != nil {
		return false, nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	if err := ctx.Err(); err != nil {
		return false, nil, err
	}

	file, diags := hclwrite.ParseConfig(data, filePath, hcl.InitialPos)
	if diags.HasErrors() {
		return false, nil, fmt.Errorf("parsing error in file %s: %v", filePath, diags.Errs())
	}
	if testHookAfterParse != nil {
		testHookAfterParse()
	}
	if err := ctx.Err(); err != nil {
		return false, nil, err
	}

	if err := reorderAttributes(file, cfg.Order, cfg.StrictOrder); err != nil {
		return false, nil, err
	}
	if testHookAfterReorder != nil {
		testHookAfterReorder()
	}
	if err := ctx.Err(); err != nil {
		return false, nil, err
	}

	formatted := bytes.ReplaceAll(file.Bytes(), []byte("\r\n"), []byte("\n"))
	styled := internalfs.ApplyHints(formatted, hints)
	original := data
	if bom := hints.BOM(); len(bom) > 0 {
		original = append(append([]byte{}, bom...), original...)
	}
	changed := !bytes.Equal(original, styled)

	var out []byte

	switch cfg.Mode {
	case config.ModeWrite:
		if !changed {
			if cfg.Stdout {
				out = styled
			}
			return false, out, nil
		}
		if err := WriteFileAtomic(ctx, filePath, formatted, perm, hints); err != nil {
			return false, nil, fmt.Errorf("error writing file %s with original permissions: %w", filePath, err)
		}
		if cfg.Stdout {
			out = styled
		}
	case config.ModeCheck:
		if cfg.Stdout {
			out = styled
		}
	case config.ModeDiff:
		if changed {
			text, err := diff.Unified(filePath, filePath, original, styled, hints.Newline)
			if err != nil {
				return false, nil, err
			}
			out = []byte(text)
		}
	}

	return changed, out, nil
}

func processReader(ctx context.Context, r io.Reader, w io.Writer, cfg *config.Config) (bool, error) {
	if w == nil {
		w = os.Stdout
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return false, err
	}

	hints := internalfs.DetectHintsFromBytes(data)
	if len(hints.BOM()) > 0 {
		data = data[len(hints.BOM()):]
	}

	file, diags := hclwrite.ParseConfig(data, "stdin", hcl.InitialPos)
	if diags.HasErrors() {
		return false, fmt.Errorf("parsing error: %v", diags.Errs())
	}
	if err := hclalign.ReorderAttributes(file, cfg.Order, cfg.StrictOrder); err != nil {
		return false, err
	}
	formatted := bytes.ReplaceAll(file.Bytes(), []byte("\r\n"), []byte("\n"))
	styled := internalfs.ApplyHints(formatted, hints)
	original := data
	if bom := hints.BOM(); len(bom) > 0 {
		original = append(append([]byte{}, bom...), original...)
	}
	changed := !bytes.Equal(original, styled)

	switch cfg.Mode {
	case config.ModeDiff:
		if changed {
			text, err := diff.Unified("stdin", "stdin", original, styled, hints.Newline)
			if err != nil {
				return false, err
			}
			if _, err := fmt.Fprint(w, text); err != nil {
				return false, err
			}
		}
	default:
		if cfg.Stdout {
			if _, err := w.Write(styled); err != nil {
				return changed, err
			}
		}
	}
	return changed, nil
}
