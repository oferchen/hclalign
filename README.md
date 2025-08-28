# hclalign

`hclalign` is a Command Line Interface (CLI) tool for reordering attributes inside HCL variable blocks. It helps maintain a consistent style across Terraform and other HCL files.

## Features

- **Variable Attribute Reordering**: Rearranges variable block attributes into a predictable order.
- **Include/Exclude Globs**: Target or skip files using `--include` and `--exclude` glob patterns.
- **Strict Ordering**: Enforce that only the specified attributes appear in the given order with `--strict-order`.
- **Check and Diff Modes**: Use `--check` to ensure files are already formatted or `--diff` to preview required changes.
- **STDIN/STDOUT Support**: Process input from standard in and write to standard out with `--stdin` and `--stdout` for easy pipeline integration.
- **Idempotent & Atomic**: Running `hclalign` multiple times produces the same result and updates files using atomic writes to prevent partial edits.
- **Concurrent Processing**: Utilizes Go's concurrency features to process files in parallel.
- **Verbose Logging**: Enable additional output with `-v` for debugging or development.

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
- `--include`: Glob patterns of files to include (default: `*.hcl`).
- `--exclude`: Glob patterns of files to exclude.
- `--order`: Attribute order for variable blocks.
- `--strict-order`: Fail if attributes appear outside the specified order.
- `--concurrency`: Maximum number of files processed in parallel (default: number of CPUs).
- `-v, --verbose`: Enable verbose logging.

## Examples

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

## Contributing

Contributions to `hclalign` are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

`hclalign` is released under the [MIT License](LICENSE).
