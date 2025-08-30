# hclalign

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

`hclalign` normalizes Terraform and other HCL files with a two‑phase pipeline that first formats input and then reorders attributes for consistent alignment.

## Pipeline

1. **fmt** – parses the file with `hclwrite` to produce canonical formatting.
2. **align** – reorders attributes to match a configurable schema.

This process is idempotent: running the tool multiple times yields the same result.

## Supported Blocks

`hclalign` aligns attributes inside Terraform blocks including `variable`, `output`, `locals`, `module`, `provider`, `terraform`, `resource`, `data`, `dynamic`, `lifecycle`, and `provisioner`.

## Schema Options

The default schema orders variable attributes as `description → type → default → sensitive → nullable → validation`. Override it with `--order`.

## Provider Schema Integration

Resource and data blocks can be ordered according to provider schemas. Supply a
schema file via `--providers-schema` or let `hclalign` invoke `terraform
providers schema -json` by passing `--use-terraform-schema`. Unknown attributes
fall back to alphabetical order.

## CLI Flags

- `--write` (default): rewrite files in place
- `--check`: exit with non‑zero status if changes are required
- `--diff`: print unified diff instead of writing files
- `--stdin`, `--stdout`: read from stdin and/or write to stdout
- `--include`, `--exclude`: glob patterns controlling which files are processed
- `--follow-symlinks`: traverse symbolic links
- `--order`: control schema ordering
- `--concurrency`: maximum parallel file processing
- `-v, --verbose`: enable verbose logging
- `--providers-schema`: path to a provider schema JSON file
- `--use-terraform-schema`: derive schema via `terraform providers schema -json`
- `--types`: comma-separated list of block types to align (defaults to `variable`)
- `--all`: align all supported block types (mutually exclusive with `--types`)
- `--prefix-order`: lexicographically sort attributes not covered by the schema

## Atomic Writes and BOM Preservation

Files are written atomically via a temporary file rename and the original newline style and optional UTF‑8 byte‑order mark (BOM) are preserved.

## Installation

```sh
git clone https://github.com/oferchen/hclalign.git
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
| `make fmt` | run gofumpt and gofmt on the codebase; `terraform fmt` then `hclalign` on test cases |
| `make lint` | execute `golangci-lint` |
| `make vet` | run `go vet` |
| `make test` | run tests with coverage |
| `make test-race` | run tests with the race detector |
| `make cover` | verify coverage ≥95% |
| `make build` | build the `hclalign` binary into `.build/` |
| `make clean` | remove build artifacts |

Terraform CLI is optional. If installed, `make fmt` runs `terraform fmt` on `tests/cases` before invoking `hclalign`; otherwise `terraform fmt` is skipped with a warning.

## Continuous Integration
Use `hclalign . --check` in CI to fail builds when formatting is needed. The provided GitHub Actions workflow runs `make tidy`, `make fmt`, `make lint`, `make test-race`, and `make cover` on Linux and macOS with multiple Go versions.

## License

`hclalign` is released under the [Apache-2.0 License](LICENSE).
