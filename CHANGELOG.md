# Changelog

<!--
WHAT: Summarize new features and safety improvements.
WHY: Keep users informed about notable changes and safeguards.
-->

All notable changes to this project will be documented in this file.

## Unreleased

- Changed project license to Apache-2.0.
- Added variable-attribute reordering tool with support for include/exclude globs.
- Introduced safety features such as check and diff modes, idempotent operation, and atomic file writes.
- Improved testing with race detector and fuzz test execution.
- Added optional symlink traversal via `--follow-symlinks` and clarified default include/exclude patterns.
- Documented exit codes and provided CI/editor usage examples to encourage safe automation.
- Enforced single-line SPDX comment rule.
- Achieved ≥95% line coverage across core packages.
