# hclalign

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

`hclalign` is a Command Line Interface (CLI) tool for reordering attributes inside HCL variable blocks. It helps maintain a consistent style across Terraform and other HCL files.

## Features

- **Variable Attribute Reordering**: Rearranges variable block attributes into a predictable order.
- **Include/Exclude Globs**: Target or skip files using `--include` and `--exclude` glob patterns.
- **Strict Ordering**: Enforce that only the specified attributes appear in the given order with `--strict-order`.
- **Check and Diff Modes**: Use `--check` to ensure files are already formatted or `--diff` to preview required changes.
- **STDIN/STDOUT Support**: Process input from standard in and write to standard out with `--stdin` and `--stdout` for easy pipeline integration.
- **Idempotent & Atomic**: Running `hclalign` multiple times produces the same result and updates files using atomic writes to prevent partial edits.
- **Concurrent Processing**: Utilizes Go's concurrency features to process files in parallel and halts further dispatch when an error occurs.
- **Optional Symlink Traversal**: Follow symbolic links during directory walks with `--follow-symlinks`.
- **Verbose Logging**: Enable additional output with `-v` for debugging or development.

## Atomic Write Guarantees

`hclalign` writes changes to a temporary file and atomically renames it over the original. This prevents partial writes and ensures files are either fully updated or left untouched if an error occurs.

## Getting Started

### Prerequisites

You must have Go installed on your system to build and run `hclalign`.

### Installation

Clone the `hclalign` repository and use the Makefile to build the project:

```sh
git clone https://github.com/oferchen/hclalign.git
cd hclalign
make init
make tidy
make build
```

This compiles the project and creates an executable named `hclalign` in the current directory.

## Usage

```sh
./hclalign [target file or directory] [flags]
```

### Common Flags

- `--write`: Write changes back to files (default behavior).
- `--check`: Exit with a non-zero status if formatting changes are needed.
- `--diff`: Print a unified diff of required changes instead of modifying files.
- `--stdin`, `--stdout`: Read from STDIN and/or write results to STDOUT.
- `--include`: Glob patterns of files to include.
- `--exclude`: Glob patterns of files to exclude.
- `--follow-symlinks`: Follow symbolic links when traversing directories.
- `--order`: Attribute order for variable blocks.
- `--strict-order`: Fail if attributes appear outside the specified order.
- `--concurrency`: Maximum number of files processed in parallel (default: number of CPUs).
- `-v, --verbose`: Enable verbose logging.

## Examples

### Basic Formatting

Format all `.tf` files under the current directory and write the result back to disk:

```sh
./hclalign . --include "**/*.tf"
```

Check whether files are already formatted:

```sh
./hclalign . --check
```

Preview the diff of required changes:

```sh
./hclalign . --diff
```

Process a single file from STDIN and write the result to STDOUT:

```sh
cat variables.tf | ./hclalign --stdin --stdout
```

### CI Usage

Use `--check` in Continuous Integration to fail builds when formatting is needed:

```sh
hclalign . --check
```

### Editor Integration

Editors can format on save by piping file contents through `hclalign`:

```sh
hclalign --stdin --stdout <file.tf >formatted.tf && mv formatted.tf file.tf
```

## Default Include and Exclude Patterns

By default, `hclalign` processes Terraform files and skips common build artifacts:

| Type    | Patterns                                                                 |
|---------|---------------------------------------------------------------------------|
| Include | `**/*.tf`                                                                |
| Exclude | `**/.terraform/**`, `**/vendor/**`, `**/.git/**`, `**/node_modules/**`    |

## Following Symlinks

Symbolic links are ignored by default to avoid unexpected traversals. Use `--follow-symlinks` to include files reachable via symlinks.

## Strict Order Enforcement

Specify attribute order with `--order`. When `--strict-order` is set, all attributes must appear exactly once in the given order and no other attributes are allowed.
Without `--strict-order`, attributes not in the canonical list remain in their original positions relative to the reordered attributes.

## Exit Codes

`hclalign` returns the following exit codes:

| Code | Meaning                                              |
|------|------------------------------------------------------|
| 0    | Success; files were formatted or already aligned.    |
| 1    | Files need formatting in `--check` or `--diff` mode. |
| 2    | Invalid usage or configuration error.                |
| 3    | Processing error.                                    |

## Contributing

Contributions to `hclalign` are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

`hclalign` is released under the [Apache-2.0 License](LICENSE).
