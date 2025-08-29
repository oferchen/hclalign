// internal/engine/scan.go
package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hclalign/config"
	"github.com/hashicorp/hclalign/patternmatching"
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
				if !cfg.FollowSymlinks {
					continue
				}
				if info.IsDir() {
					if err := walk(ctx, path); err != nil {
						return err
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
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		resolved, err := os.Stat(cfg.Target)
		if err != nil {
			return nil, err
		}
		if resolved.IsDir() {
			if cfg.FollowSymlinks {
				if err := walk(ctx, cfg.Target); err != nil {
					return nil, err
				}
			}
		} else if cfg.FollowSymlinks {
			if matcher.Matches(cfg.Target) {
				files = append(files, cfg.Target)
			}
		}
	} else if info.IsDir() {
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
