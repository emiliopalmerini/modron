# ADR-002: Migrate from MCP Server to CLI (rename to `modron`)

**Status**: Accepted
**Date**: 2026-05-06

## Context

ADR-001 established a custom Notion MCP server to give agents deterministic, typed, single-call access to the user's Notion databases. The implementation works, but the MCP-only transport has become a constraint:

- The server is only callable from MCP clients (Claude Code). Cron jobs, shell scripts, scheduled remote agents, and direct terminal use cannot invoke it.
- Every MCP client session pays a context-window cost for the full tool catalogue (~12 tools), even when none are used in that session.
- The deterministic-call property that motivated ADR-001 is a function of the tool surface (typed parameters, single API call, validated inputs), not of the MCP transport. A CLI with one subcommand per current MCP tool preserves that property.
- A persistent stdio process offers no real benefit here: every operation is a stateless HTTP call to the Notion API.

A CLI binary, by contrast, is composable with shell pipes, callable from any scheduler, distributable as a single static binary, and invokable from MCP-capable agents via a shell tool when needed.

The repository name `notion-mcp` no longer fits a CLI tool, so this migration also renames the project.

## Decision

Replace the MCP server with a CLI binary named **`modron`**. Drop the MCP transport entirely; do not ship a shim. Rename the repo, Go module, Nix package, binary, and local working directory in the same migration.

### Naming

- Binary: `modron`
- Go module path: `github.com/emiliopalmerini/modron`
- GitHub repo: `emiliopalmerini/modron` (renamed from `emiliopalmerini/notion-mcp`)
- Local working directory: `~/src/tools/modron`
- Nix package `pname`: `modron`

Rationale: modrons are the lawful-neutral constructs of Mechanus, embodying perfect order; the metaphor matches structured Notion databases. Pairs with the existing `mimic` CLI in the same toolset (both D&D-themed, both five characters).

### Entry Point

- New: `cmd/modron/main.go`.
- Old `cmd/notion-mcp/main.go` is removed.

### CLI Framework

Use `github.com/spf13/cobra` for subcommand routing and flag parsing. Rationale: ubiquitous in Go, well-documented help output, supports nested subcommands cleanly, generates shell completions.

### Command Surface

One top-level command per slice; verbs as subcommands. Mirrors existing MCP tools 1:1.

```
modron projects query    [flags]
modron projects create   [flags]
modron tasks query       [flags]
modron tasks create      [flags]
modron events query      [flags]
modron events create     [flags]
modron blackhole query   [flags]
modron blackhole create  [flags]
modron page get          <page_id>
modron page update       <page_id> --properties <json>
```

Flag names match current MCP parameter names exactly. Snake_case MCP parameters become kebab-case flags (`due_before` → `--due-before`, `start_cursor` → `--start-cursor`, `sort_direction` → `--sort-direction`, `has_launch_date` → `--has-launch-date`, etc.).

### Output Format

Each command supports `--output <format>` with values:

- `text` (default): human-readable markdown, identical to current `formatQueryResult` output.
- `json`: machine-readable structured output. Query commands emit `{"results": [...], "has_more": bool, "next_cursor": "..."}`. Create/update commands emit `{"id": "...", "url": "...", "name": "..."}`.

Errors always go to stderr; non-zero exit codes signal failure.

### Slice Architecture

Hexagonal slice structure from ADR-001 is preserved. Each slice:

- `domain.go`, `ports.go`, `notion.go` are unchanged in behavior; only import paths update for the module rename.
- `mcp.go` is replaced by `cli.go`, which exports a `NewCommand(repo Repository) *cobra.Command` returning the subcommand tree for that slice.
- `mcp_test.go` is replaced by `cli_test.go`. Tests assert flag parsing, validation, and output formatting against a fake Repository.

`cmd/modron/main.go` wires the slices: constructs the Notion client, instantiates each slice's repository, attaches each slice's root command to the program root, then `Execute()`s.

### Inputs

- `NOTION_TOKEN`: env var, unchanged.
- All flag values: parsed and validated by cobra/pflag, then passed into the same `Filter` / `CreateParams` structs the slices already use.

### Outputs

- stdout: command result in the requested format.
- stderr: errors and validation messages.
- exit code: 0 on success, non-zero on any error (validation, API, network).

### Edge Cases

- **Empty filters**: same behavior as ADR-001 (return all pages up to `--page-size`, sorted by last edited descending).
- **Invalid select/status values**: validated before the API call; print error with valid options to stderr; exit 1.
- **Pagination**: `--start-cursor` flag accepts a cursor; `text` output appends a `_More results available. Use start_cursor: ..._` line; `json` output includes `has_more` and `next_cursor` fields.
- **Rate limiting**: unchanged; existing 429 retry/backoff in the shared client applies.
- **Relation filters**: `--project-id` / `--parent-task-id` accept page UUIDs as before.
- **Date parameters**: ISO-8601 dates or datetimes as before.
- **Missing required positional args**: cobra emits its standard usage error to stderr; exit 1.

### Error Conditions

- Missing `NOTION_TOKEN`: print error to stderr, exit 1 (was: `log.Fatal` at startup).
- Notion API errors (400, 401, 404, 429, 500): mapped to descriptive messages on stderr, exit 1.
- Invalid flag values: validated before API call; error to stderr listing valid options; exit 1.

### Migration Steps

1. Update `go.mod` module path to `github.com/emiliopalmerini/modron`; rewrite all internal imports.
2. Add `github.com/spf13/cobra` dependency.
3. For each slice, write `cli_test.go` first (TDD), confirm failures, then add `cli.go` exporting `NewCommand(repo Repository) *cobra.Command`.
4. Add `cmd/modron/main.go` with cobra root command and slice wiring.
5. Delete `cmd/notion-mcp/`, every slice's `mcp.go` and `mcp_test.go`, and the `mark3labs/mcp-go` dependency from `go.mod`. Run `go mod tidy`.
6. Update `flake.nix`: `description`, `pname`, package name, `subPackages` → `cmd/modron`, refresh `vendorHash`.
7. Verify `go vet`, `go test ./...`, and `nix build` (or `go build`) succeed.
8. Commit in atomic chunks per repo conventions.
9. Rename GitHub repo via `gh repo rename modron`; update local `origin` remote URL.
10. Rename local working directory: `mv ~/src/tools/notion-mcp ~/src/tools/modron`.

## Consequences

- Binary becomes invokable from any context (cron, shell, scripts, scheduled agents, MCP clients via a generic shell tool). Removes the MCP-client constraint.
- MCP-context overhead disappears for clients that don't need Notion access in a given session.
- Agents calling the CLI must construct flag strings rather than receive a typed tool schema. Determinism is preserved by the typed flag surface and validation, but per-call ergonomics shift from "tool form" to "shell command." Agents that previously consumed the MCP tool catalogue must be updated to invoke the CLI.
- The `mark3labs/mcp-go` dependency and its transitive deps are removed.
- All external references to `notion-mcp` (clones, bookmarks, CI configs, agent instructions outside this repo) must be updated to `modron`. The GitHub repo rename installs a redirect, but the Go module path change is hard-cutover.
- ADR-001's database schemas, tool surface, and validation rules remain authoritative for the CLI's behavior; this ADR only changes transport and project name.
