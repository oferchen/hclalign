# Contributing to hclalign

Thank you for considering a contribution to `hclalign`!

## Comment Policy

Every Go source file must begin with a single comment on line 1 containing its repository‑relative path:

```go
// path/to/file.go
```

Run `make strip` before committing to normalize file headers and verify the policy with `cmd/commentcheck`.

## Development Workflow

From the repository root, run the following targets before submitting a pull request:

```sh
make tidy   # download modules and tidy go.mod
make fmt    # format Go code and regenerate fixtures
make strip  # enforce comment policy
make lint   # static analysis via golangci-lint
make test   # execute tests with coverage
make cover  # enforce >=95% coverage
make build  # compile the hclalign binary
```

## Continuous Integration

GitHub Actions executes `make tidy`, `make fmt`, `make lint`, `make test`, `make cover`, and `make build` on Linux and macOS with multiple Go versions. Pull requests must pass this workflow.

## Commit Guidelines

- Keep commits focused and avoid unrelated changes.
- Use imperative, descriptive commit messages.
- Ensure the working tree is clean and all checks pass before committing.

We appreciate your help in improving `hclalign`!
