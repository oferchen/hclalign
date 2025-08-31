# hclalign

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

`hclalign` normalizes Terraform and other HCL files with a two‑phase pipeline that mirrors `terraform fmt` before reordering attributes for consistent alignment.

## Pipeline

1. **fmt** – detects the Terraform CLI with `exec.LookPath`; if found it runs `terraform fmt`, otherwise a pure Go formatter is used. Newline and BOM hints are carried through and applied only when writing the result.
2. **align** – reorders attributes to match a configurable schema.

`terraform fmt` is run again after alignment to ensure canonical layout. This process is idempotent: running the tool multiple times yields the same result.

## Supported Blocks

`hclalign` aligns attributes inside Terraform blocks including `variable`, `output`, `locals`, `module`, `provider`, `terraform`, `resource`, `data`, `dynamic`, `lifecycle`, and `provisioner`.

## Schema Options

The default schema orders variable attributes as `description → type → default → sensitive → nullable → validation`. Additional block types have their own canonical order:

- **output:** `description`, `value`, `sensitive`, `depends_on`
- **module:** `source`, `version`, `providers`, `count`, `for_each`, `depends_on`, then input variables alphabetically
- **provider:** `alias` followed by other attributes alphabetically
- **terraform:** `required_version`, `required_providers` (entries sorted alphabetically), `backend`, `cloud`, then remaining attributes and blocks in their original order
- **resource/data:** `provider`, `count`, `for_each`, `depends_on`, `lifecycle`, `provisioner`, then provider schema attributes grouped as required → optional → computed (each alphabetical)

Validation blocks are placed immediately after canonical attributes. Attributes not in the canonical list or provider schema are appended in their original order.

Entries within `required_providers` are sorted alphabetically by provider name. Other `terraform` attributes follow `required_version`, `backend`, and `cloud` in that order.

### `required_providers` sorting

```hcl
# before
terraform {
  required_providers {
    azurerm = {
      source = "hashicorp/azurerm"
    }
    aws = {
      source = "hashicorp/aws"
    }
  }
}

# after
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
    azurerm = {
      source = "hashicorp/azurerm"
    }
  }
}
```

### Flag interactions

Use `--types` to select which block types to align. `--order` customizes the variable schema and has no effect on other block types:

```sh
# align module and output blocks using their default order
hclalign . --types module,output

# override variable attribute order while still aligning modules with defaults
hclalign . --types variable,module --order value,description,type

# --order is ignored when variable blocks are not selected
hclalign . --types module --order value,description,type
```

## Provider Schema Integration

Resource and data blocks can be ordered according to provider schemas. Supply a
schema file via `--providers-schema` or let `hclalign` invoke `terraform
providers schema -json` by passing `--use-terraform-schema`. Unknown attributes
keep their original order.

## CLI Flags

- `--write` (default): rewrite files in place
- `--check`: exit with non‑zero status if changes are required
- `--diff`: print unified diff instead of writing files
- `--stdin`, `--stdout`: read from stdin and/or write to stdout
- `--include`, `--exclude`: glob patterns controlling which files are processed (defaults: include `**/*.tf`, `**/*.tfvars`; exclude `.terraform/**`, `**/vendor/**`)
- `--follow-symlinks`: traverse symbolic links
- `--order`: control variable attribute order
- `--concurrency`: maximum parallel file processing
- `-v, --verbose`: enable verbose logging
- `--providers-schema`: path to a provider schema JSON file
- `--use-terraform-schema`: derive schema via `terraform providers schema -json`
- `--types`: comma-separated list of block types to align (defaults to `variable`)
- `--all`: align all supported block types (mutually exclusive with `--types`)


## Exit Codes

- `0`: success
- `1`: changes required when running with `--check` or `--diff`
- `2`: invalid CLI usage or configuration error
- `3`: processing error during formatting or alignment

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
| `make fmt` | run `gofumpt` (v0.6.0); regenerate golden files; run `terraform fmt` on test cases if available |
| `make strip` | remove comments and enforce the single-line comment policy |
| `make lint` | execute `golangci-lint` |
| `make vet` | run `go vet` |
| `make test` | run tests with coverage |
| `make test-race` | run tests with the race detector |
| `make fuzz` | run fuzz tests |
| `make cover` | verify coverage ≥95% |
| `make build` | build the `hclalign` binary into `.build/` |
| `make clean` | remove build artifacts |

Terraform CLI is optional. If installed, `make fmt` runs `terraform fmt` on `tests/cases` and regenerates golden test files.

`make fmt` uses `go run mvdan.cc/gofumpt@v0.6.0` so contributors do not need to install `gofumpt` manually.

## Continuous Integration
Use `hclalign . --check` in CI to fail builds when formatting is needed. The provided GitHub Actions workflow runs `make tidy`, `make fmt`, `make lint`, `make test-race`, and `make cover` on Linux and macOS with multiple Go versions.

## License

`hclalign` is released under the [Apache-2.0 License](LICENSE).
