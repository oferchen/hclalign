# hclalign

`hclalign` is a Command Line Interface (CLI) application designed to organize and align HCL (HashiCorp Configuration Language) files. This tool is invaluable for projects utilizing HCL, including Terraform configurations, by improving readability and ensuring consistency across your codebase.

## Features

- **Flexible File Selection**: Use glob patterns to target specific files or groups of files for alignment.
- **Customizable Field Order**: Define your preferred order of variable block fields to maintain a consistent format across HCL files.
- **Concurrent Processing**: Leverages Go's concurrency capabilities to process files in parallel, optimizing performance on multicore systems.
- **Debugging Support**: Includes a debug mode for additional logging, useful for troubleshooting and development.

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

This will compile the project, creating an executable named `hclalign` in the current directory.

### Basic Commands

To get started with `hclalign`, here are some Makefile commands you might find useful:

- **Build**: `make build` compiles the project into an executable.
- **Run**: `make run` builds and executes `hclalign`.
- **Dependencies**: `make deps` downloads the required project dependencies.
- **Tidy**: `make tidy` cleans and verifies the module dependencies.
- **Test**: `make test` runs all configured tests.
- **Clean**: `make clean` removes temporary files and the executable to clean up the project space.
- **Init**: `make init` initializes a new Go module for the project.
- **Help**: `make help` displays available Makefile commands.

## Usage

Execute `hclalign` with your target directory or file and optional flags for criteria and order:

```sh
./hclalign [target file or directory] --criteria "*.tf,*.hcl" --order "description,type,default,sensitive,nullable,validation"
```

### Command Line Flags

- `--criteria, -c`: Glob patterns for selecting files (default: `*.tf`).
- `--order, -o`: Specify the order of variable block fields (default: `description,type,default,sensitive,nullable,validation`).

## Examples

**Before `hclalign`**, a Terraform variable file might look like this:

```hcl
variable "instance_type" {
  type        = string
  description = "EC2 instance type"
  default     = "t2.micro"
}

variable "instance_count" {
  description = "Number of instances to launch"
  type        = number
  default     = 1
}
```

**After running `hclalign`**, the file would be reorganized for improved clarity:

```hcl
variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t2.micro"
}

variable "instance_count" {
  description = "Number of instances to launch"
  type        = number
  default     = 1
}
```

## Contributing

Contributions to `hclalign` are welcome! Feel free to fork the repository, submit pull requests, or open issues for any bugs, feature suggestions, or improvements.
