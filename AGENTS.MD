# AGENTS.MD

Operational guide for automation and AI assistants working on **hclalign**.

---

## Purpose

Coordinate autonomous agents (and humans) to deliver a **production-grade, two-phase Terraform formatter**:

1. **Phase A – fmt:** replicate `terraform fmt` semantics byte-for-byte (stdin/stdout and files).
2. **Phase B – align:** granular, token-preserving re-ordering across Terraform block types according to HashiCorp specifications and (optionally) provider schemas.

Non-negotiables:

* **Apache-2.0** license for the repo.
* **Comment policy:** every `.go` file contains exactly **one** comment on line 1: `// <repo-relative-path>`.
* **Atomic writes:** same-dir temp → fsync → rename; preserve mode; preserve BOM/CRLF.
* **Idempotency:** running twice yields identical bytes.
* **Coverage:** ≥ **95%** functional coverage (enforced in CI).
* **No deprecated APIs,** no placeholders, no pseudo-code.

---

## Repository Context (short)

* Language: Go ≥ 1.21
* HCL engine: `hcl/v2/hclwrite` with **SetAttributeRaw** + **BuildTokens** only
* CLI: `--providers-schema`, `--use-terraform-schema`, `--check`, `--diff`
* Defaults: include `**/*.tf`, `**/*.tfvars`; exclude `.terraform/**`, `**/vendor/**`
* Structure (target):

  * `/cmd/hclalign/`
  * `/internal/fmt/`  (terraform fmt strategies)
  * `/internal/align/` (per-block strategies)
  * `/internal/engine/` (scanner, pipeline, workers)
  * `/internal/hcl/` (token/BOM/CRLF helpers)
  * `/internal/fs/`  (atomic writer)
  * `/internal/diff/`
  * `/tests/cases/**` (golden)

---

## Agents & Responsibilities

### 1) **Formatting Agent (FMT)**

**Goal:** Phase A parity with `terraform fmt`.

* **Tasks**

  * Implement `fmt.Run(strategy)` with `auto|binary|go` strategies.
  * Binary strategy: invoke `terraform fmt -` via stdin/stdout; validate exit.
  * Go strategy: reproduce layout rules via `hclwrite` and verify parity on fixtures.
* **Done when:** `fmt` output equals Terraform’s on the matrix; CRLF/BOM preserved.

### 2) **Alignment Agent (ALIGN)**

**Goal:** Phase B granular re-ordering with tokens preserved.

* **Tasks**

  * Implement strategy registry per block type:

    * `variable`: `description, type, default, sensitive, nullable, <other>, validation*`
    * `output`: `description, value, sensitive, depends_on, <other>`
    * `locals`: **no reordering** (fmt only)
    * `module`: `source, version, providers, count, for_each, depends_on, <inputs alpha>, <other>`
    * `provider`: `alias, <config attrs alpha>, <blocks>`
    * `terraform`: `required_version, experiments, required_providers(alpha), backend, cloud, <other>`
    * `resource/data`: meta-args first `provider, count, for_each, depends_on, lifecycle, provisioner*`, then **schema-driven** tiers: required → optional → computed (each alpha). Fallback to alpha when schema unknown.
  * Move **attributes** only; keep nested blocks & their order unless explicitly defined.
  * Use `BuildTokens(nil)` + `SetAttributeRaw`; never rebuild expressions.
* **Done when:** Golden tests pass; non-target bytes unchanged (aside from fmt).

### 3) **Schema Agent (SCHEMA)**

**Goal:** Provider schema integration.

* **Tasks**

  * Parse `terraform providers schema -json` (runtime or pre-generated `--providers-schema`).
  * Build resource/data attribute tiers per type.
  * Cache per-run; no network I/O.
* **Done when:** Resource ordering respects tiers in tests for ≥2 providers (e.g., `aws`, `null`).

### 4) **Filesystem Safety Agent (FS)**

**Goal:** Zero data-loss writes.

* **Tasks**

  * Implement `/internal/fs/atomic.go`: tmp in same dir → write → **fsync(fd)** → **fsync(dir)** → rename.
  * Preserve file mode; detect & preserve BOM and CRLF.
* **Done when:** Write tests verify atomicity, perms, BOM/CRLF round-trip.

### 5) **Test & Coverage Agent (TEST)**

**Goal:** ≥95% functional coverage.

* **Tasks**

  * Golden suite `tests/cases/*/{in.tf,fmt.tf,aligned.tf}` with auto-discovery.
  * Idempotency tests (fmt+align run twice).
  * Error-path tests (parse fail, missing terraform, malformed schema).
  * Fuzz entrypoints for align/token helpers.
* **Done when:** `make cover` ≥ 95% or CI fails.

### 6) **CI Agent (CI)**

**Goal:** Deterministic, cross-platform CI.

* **Tasks**

  * GH Actions matrix: `os: [ubuntu-latest, macos-latest]`, Go: latest two minors.
  * Steps: tidy, lint, vuln (non-blocking), test (`-race -shuffle=on -coverprofile`), build, **coverage gate**.
  * Optional Terraform fmt parity job using a Terraform container.
* **Done when:** Pipeline green; parity job reports 100% match on FMT fixtures.

### 7) **Release Agent (REL)**

**Goal:** Repeatable releases.

* **Tasks**

  * Tag-driven build with `-trimpath`.
  * Attach coverage, checksums, SBOM (optional).
* **Done when:** Draft release from tag builds artifacts for linux/macOS amd64/arm64.

### 8) **Docs Agent (DOCS)**

**Goal:** Minimal, frictionless docs.

* **Tasks**

  * Update README (pipeline, flags, guarantees, limitations).
  * Write `CONTRIBUTING.md` (coding standards, comment policy).
  * Maintain this `AGENTS.MD`.
* **Done when:** Docs match actual behavior; quickstart succeeds.

### 9) **Comment Policy Agent (CMT)**

**Goal:** Enforce single-line banner comments.

* **Tasks**

  * Build `/cmd/commentcheck` that fails if any `.go` contains comments beyond line 1.
* **Done when:** CI fails on violation.

---

## Standard Prompts (for coding assistants)

### General Implementation Prompt

```
Act as a senior Go engineer for hclalign. Implement/modify code to:
1) Run Phase A (terraform fmt parity) then Phase B (token-preserving alignment).
2) Preserve BOM/CRLF; atomic write with fsync+rename; preserve file mode.
3) Support schema-driven ordering for resource/data; fallback to alpha.
4) Enforce single-line filename comment at line 1 in all .go files.
5) Raise coverage ≥95% with golden, unit, idempotency, and fuzz tests.
6) No deprecated APIs, no placeholders, no panics on user input.
Refactor only within /internal/* and /cmd/hclalign, keep CLI stable, add flags as documented in AGENTS.MD.
```

### Test Authoring Prompt

```
Create golden tests under tests/cases/** with {in.tf, fmt.tf, aligned.tf}.
Cover: variable (multi validation), output, locals, module, provider, terraform, resource/data (meta-args, nested, dynamic, lifecycle pre/post, provisioner).
Add CRLF+BOM, trailing commas, unicode, heredocs (<<EOF/<<-EOF), templates (${ } and %{ if }).
Add schema JSON fixtures for aws and null providers; verify tiering.
Add idempotency tests (second run no diff). Add fuzz for align/token helpers.
Target overall coverage ≥95%; fail if below.
```

### Safety/FS Prompt

```
Implement /internal/fs/atomic.go:
- Write to same-directory temp file, fsync file, fsync directory, then rename.
- Preserve mode bits.
- Detect and preserve BOM/CRLF.
Add unit tests that simulate crashes before/after rename and assert durability.
```

---

## Workflows

### Local

```
make tidy
make lint
make test         # -race -shuffle=on -cover
make cover        # enforces ≥95%
make build
```

### CI (summary)

* Lint → Vuln (warn) → Test (race, shuffle, cover) → Coverage gate → Build
* Optional job: Terraform binary parity for Phase A

---

## Acceptance Gates

* **FMT parity:** Phase A equals `terraform fmt` on fixtures.
* **Granular alignment:** Only expected reordering; values & nested blocks untouched.
* **Schema tiers:** Resource/data ordering respects required/optional/computed when schemas provided.
* **Atomicity & preservation:** BOM/CRLF + perms preserved; idempotent output.
* **Coverage:** ≥95% or fail.
* **Comment policy:** enforced; any violation fails CI.
* **License:** Apache-2.0 present and referenced in README.

---

## Runbooks

**Parity mismatch (FMT):**

1. Confirm Terraform binary is present.
2. If mismatched, update formatter and add failing fixture to `tests/cases/`.

**Unexpected reordering:**

1. Reproduce on minimal case; verify block strategy.
2. If schema-driven, inspect JSON tiers; add/adjust fixture.

**CRLF/BOM lost:**

1. Verify hints path in stdin/stdout and file write flows.
2. Add preservation test; fix `/internal/hcl` helpers.

**Coverage drop:**

1. Run `go tool cover -func=coverage.out`.
2. Add tests where missed; prefer golden for behavior.

---

## Do & Don’t

**Do**

* Keep edits surgical with `SetAttributeRaw`.
* Maintain stable CLI & exit codes.
* Keep commits small; include tests with features.

**Don’t**

* Rebuild expressions or normalize commas.
* Change nested block order/content unless specified.
* Introduce new comments beyond line-1 banner.

---

## Versioning & Releases

* SemVer; bump minor for new flags; patch for fixes.
* Release notes must list: behavior changes, new flags, schema support updates, test coverage.

---

*End of AGENTS.MD.*
