# Contributing to hclalign

Thank you for considering a contribution to `hclalign`!

## Coding Standards

- Go code should be formatted with `gofmt` and organized with `goimports`.
- Follow idiomatic Go practices and keep functions small and well tested.
- New features should include unit tests and, when appropriate, fuzz tests.

## Required Checks

Before submitting a pull request, ensure all of the following commands succeed:

```sh
golangci-lint run
go test -race ./...
go test -run Fuzz ./...
```

These commands must be run from the repository root. Additional tests are welcome.

## Commit Guidelines

- Make commits concise and focused; avoid unrelated changes.
- Write meaningful commit messages using the imperative mood.
- Ensure the working tree is clean and all checks pass before committing.

We appreciate your help in improving `hclalign`!
