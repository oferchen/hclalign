# Contributing to hclalign

Thank you for considering a contribution to `hclalign`!

## Coding Standards

- Go code should be formatted with `gofmt` and organized with `goimports`.
- Follow idiomatic Go practices and keep functions small and well tested.
- New features should include unit tests and, when appropriate, fuzz tests.

## Required Checks

Run the comment stripping tool to ensure each Go file begins with a path comment:

```sh
make strip-comments
```

Before submitting a pull request, ensure the continuous integration pipeline passes:

```sh
make ci
```

This command formats the code, vets, lints, runs the tests with the race detector,
performs a short fuzz run, and builds the project. If you prefer to run steps
individually, execute the following from the repository root:

```sh
make fmt
make vet
make lint
make test-race
make fuzz-short
make build
```

## Commit Guidelines

- Make commits concise and focused; avoid unrelated changes.
- Write meaningful commit messages using the imperative mood.
- Ensure the working tree is clean and all checks pass before committing.

We appreciate your help in improving `hclalign`!
