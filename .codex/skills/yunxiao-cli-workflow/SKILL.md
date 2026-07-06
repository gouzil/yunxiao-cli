---
name: yunxiao-cli-workflow
description: Use when working in the yunxiao-cli repository on CLI commands, typed Yunxiao OpenAPI services, extension behavior, auth/config handling, output contracts, Chinese docs, or live Yunxiao smoke checks.
---

# Yunxiao CLI Workflow

## Principle

Start from the real command surface, then make the smallest repo-shaped change. Keep command handlers typed, docs Chinese, and extension auth host-owned.

## Workflow

1. Inspect the live surface before editing:
   - Structural code: use CodeGraph for `internal/app/root.go`, the relevant `internal/app/*_commands.go`, and `internal/yunxiao/*`.
   - Literal docs/text: use `rg` in `README.md`, `docs/`, and `openspec/changes/*/{api-map.md,output-map.md}`.
2. Keep command ownership boring:
   - Root command registration lives in `internal/app/root.go`.
   - Cobra handlers live in `internal/app/*_commands.go`.
   - Cloud/API calls go through typed interfaces and structs in `internal/yunxiao`; do not promote loose `map[string]any` to a public command/service contract.
   - Human output goes through `internal/output`; update `docs/output-contracts.md` when output semantics change.
3. Preserve the extension boundary:
   - Extensions are local executables named `yunxiao-<name>`, not Go plugins.
   - Do not pass PATs into extensions. Extensions call `yunxiao api`; the host CLI reads credentials and injects auth into HTTP requests.
   - Extension env is non-sensitive context only: endpoint, organization, project, repo.
   - Keep `yunxiao api` raw body on stdout; diagnostics and verbose request info stay on stderr.
4. Keep docs aligned:
   - Public repo docs and README are Chinese by default.
   - Use `internal/app/root.go` plus command files as command truth, not stale prose.
   - Update `docs/api-services.md`, `docs/command-reference.md`, `docs/output-contracts.md`, or `docs/extension-authoring.md` only when the touched behavior changes.
   - Do not edit `openspec/changes/` planning docs unless the user asks for OpenSpec work.
5. Verify with the smallest relevant command:
   - Normal code change: `./scripts/test.sh`.
   - Static/lint-sensitive change: `./scripts/vet.sh`.
   - Build/release surface: `go build ./cmd/yunxiao` or `./scripts/build.sh`.
   - Live cloud smoke only when credentials/environment exist; prefer `YUNXIAO_ENDPOINT=https://openapi-rdc.aliyuncs.com` and read-only MR/repo checks.

## Guardrails

- Preserve unrelated dirty-tree changes.
- Keep auth on the host side and stdout/stderr boundaries strict.
- Do not mark real-cloud tasks complete without actual Yunxiao credentials and live evidence.
- Leave `repo unarchive` and legacy cross-repo search as pending/confirmed-by-live-env only until the API is verified.
