// internal/engine/scan.go
package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/patternmatching"
)

func scan(ctx context.Context, cfg *config.Config) ([]string, error) {
	if _, err := os.Stat(cfg.Target); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("target %q does not exist", cfg.Target)
		}
		return nil, err
	}
	matcher, err := patternmatching.NewMatcher(cfg.Include, cfg.Exclude, cfg.Target)
	if err != nil {
		return nil, err
	}

	rootAbs, err := filepath.Abs(cfg.Target)
	if err != nil {
		return nil, err
	}
	rootEval, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		rootEval = rootAbs
	}

	var files []string
	visited := make(map[string]struct{})

	var walk func(context.Context, string) error
	walk = func(ctx context.Context, dir string) error {
		if !matcher.Matches(dir) {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		realDir, err := filepath.EvalSymlinks(absDir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		relDir, err := filepath.Rel(rootEval, realDir)
		if err != nil || relDir == ".." || strings.HasPrefix(relDir, ".."+string(os.PathSeparator)) {
			return nil
		}
		if _, ok := visited[realDir]; ok {
			return nil
		}
		visited[realDir] = struct{}{}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := ctx.Err(); err != nil {
				return err
			}
			path := filepath.Join(dir, entry.Name())
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			path = absPath
			info, err := os.Lstat(path)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				if !cfg.FollowSymlinks {
					continue
				}
				realPath, err := filepath.EvalSymlinks(path)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return err
				}
				rel, err := filepath.Rel(rootEval, realPath)
				if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
					continue
				}
				info, err = os.Stat(realPath)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return err
				}
				path = realPath
			}
			if info.IsDir() {
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

	var info os.FileInfo
	if cfg.FollowSymlinks {
		info, err = os.Stat(cfg.Target)
	} else {
		info, err = os.Lstat(cfg.Target)
	}
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 && !cfg.FollowSymlinks {
		return nil, nil
	}
	if info.IsDir() {
		if err := walk(ctx, cfg.Target); err != nil {
			return nil, err
		}
	} else {
		if matcher.Matches(cfg.Target) {
			files = append(files, cfg.Target)
		}
	}

	sort.Strings(files)
	return files, nil
}
