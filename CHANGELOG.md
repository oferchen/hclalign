# Changelog

<!--
WHAT: Summarize new features and safety improvements.
WHY: Keep users informed about notable changes and safeguards.
-->

All notable changes to this project will be documented in this file.

## Unreleased

- Aligned toolchain with Go 1.23 and gofumpt v0.8.0.
- Changed project license to Apache-2.0.
- Added variable-attribute reordering tool with support for include/exclude globs and the `--order` flag for custom schemas.
- Introduced safety features such as check and diff modes, idempotent operation, and atomic file writes.
- Improved testing with race detector and fuzz test execution.
- Documented exit codes and provided CI/editor usage examples to encourage safe automation.
- Enforced single-line SPDX comment rule.
- Achieved ≥95% line coverage across core packages.
- Introduced `--prefix-order` flag to alphabetize non-canonical attributes and sort module provider maps.
- Updated documentation for provider alphabetical ordering, new `ephemeral` terraform attribute, default excludes, and the `--prefix-order` flag with usage examples.
- Removed `experiments` from the canonical `terraform` block order.
- Sorted `provider` block attributes alphabetically after `alias`.
- Added `ephemeral` to canonical ordering for output blocks.
