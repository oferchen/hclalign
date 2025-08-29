# hclalign

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

`hclalign` normalizes Terraform and other HCL files with a two‑phase pipeline that first formats input and then reorders attributes for consistent alignment.

## Pipeline

1. **fmt** – parses the file with `hclwrite` to produce canonical formatting.
2. **align** – reorders attributes to match a configurable schema.

This process is idempotent: running the tool multiple times yields the same result.

## Supported Blocks

`hclalign` currently aligns attributes inside Terraform `variable` blocks. Other block types are left untouched.

## Schema Options

The default schema orders variable attributes as `description`, `type`, `default`, `sensitive`, `nullable`. Override it with `--order` and enforce that only those attributes appear by adding `--strict-order`.

## CLI Flags

- `--write` (default): rewrite files in place
- `--check`: exit with non‑zero status if changes are required
- `--diff`: print unified diff instead of writing files
- `--stdin`, `--stdout`: read from stdin and/or write to stdout
- `--include`, `--exclude`: glob patterns controlling which files are processed
- `--follow-symlinks`: traverse symbolic links
- `--order`, `--strict-order`: control and enforce schema ordering
- `--concurrency`: maximum parallel file processing
- `-v, --verbose`: enable verbose logging

## Atomic Writes and BOM Preservation

Files are written atomically via a temporary file rename and the original newline style and optional UTF‑8 byte‑order mark (BOM) are preserved.

## Installation

```sh
git clone https://github.com/hashicorp/hclalign.git
cd hclalign
make init tidy build
```

The binary is created at `.build/hclalign`.

## Usage

```sh
hclalign [path] [flags]
```

### Examples

Format all `.tf` files under the current directory and write the result back:

```sh
hclalign . --include "**/*.tf"
```

Check whether files are already formatted:

```sh
hclalign . --check
```

Preview the diff of required changes:

```sh
hclalign . --diff
```

Process a single file from STDIN and write to STDOUT:

```sh
cat variables.tf | hclalign --stdin --stdout
```

## Make Targets

| Target | Description |
| --- | --- |
| `make init` | download and verify Go modules |
| `make tidy` | tidy module dependencies |
| `make fmt` | run gofumpt and gofmt on the codebase |
| `make lint` | execute `golangci-lint` |
| `make vet` | run `go vet` |
| `make test` | run tests with coverage |
| `make test-race` | run tests with the race detector |
| `make cover` | verify coverage ≥95% |
| `make build` | build the `hclalign` binary into `.build/` |
| `make commentcheck` | ensure source files include license comments |
| `make fix-comments` | automatically insert missing license comments |
| `make clean` | remove build artifacts |

## Continuous Integration

Use `hclalign . --check` in CI to fail builds when formatting is needed. The provided GitHub Actions workflow runs `make tidy`, `make fmt`, `make lint`, `make test-race`, `make cover`, and `make commentcheck` on Linux and macOS with multiple Go versions.

## License

`hclalign` is released under the [Apache-2.0 License](LICENSE).
