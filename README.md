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

The default schema orders variable attributes as `description`, `type`, `default`, `sensitive`, `nullable`. Override it with `--order` and enforce that only those attributes appear by adding `--strict-order`.

## Formatting Strategies

The fmt phase supports multiple strategies controlled by `--fmt-strategy`:
`auto` chooses the Terraform binary if available, `binary` always shells out to
`terraform fmt`, and `go` uses the built-in formatter. Use `--fmt-only` to stop
after formatting or `--no-fmt` to skip this phase.

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
- `--order`, `--strict-order`: control schema ordering; `--strict-order` applies globally to all blocks
- `--concurrency`: maximum parallel file processing
- `-v, --verbose`: enable verbose logging
- `--fmt-only`: run only the formatting phase
- `--no-fmt`: skip the formatting phase
- `--fmt-strategy {auto,binary,go}`: choose formatting backend
- `--providers-schema`: path to a provider schema JSON file
- `--use-terraform-schema`: derive schema via `terraform providers schema -json`

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
| `make fmt` | run gofumpt and gofmt on the codebase |
| `make lint` | execute `golangci-lint` |
| `make vet` | run `go vet` |
| `make test` | run tests with coverage |
| `make test-race` | run tests with the race detector |
| `make cover` | verify coverage ≥95% |
| `make build` | build the `hclalign` binary into `.build/` |
| `make clean` | remove build artifacts |

## Continuous Integration
Use `hclalign . --check` in CI to fail builds when formatting is needed. The provided GitHub Actions workflow runs `make tidy`, `make fmt`, `make lint`, `make test-race`, and `make cover` on Linux and macOS with multiple Go versions.

## License

`hclalign` is released under the [Apache-2.0 License](LICENSE).
