# Yunxiao CLI Live Test Checklist

Target repository: `https://codeup.aliyun.com/6a49f26608f52788b13355d2/test-repo/tree/master`

Target identifiers:

- Organization: `6a49f26608f52788b13355d2`
- Repository path: `6a49f26608f52788b13355d2/test-repo`
- Repository ID: `7123450`
- Default branch: `master`

Status legend:

- `[ ]` not started
- `[~]` in progress or blocked by a CLI/API defect under investigation
- `[x]` verified
- `[n/a]` intentionally skipped with reason

## Progress

- [x] Inventory supported top-level commands from `yunxiao --help`.
- [x] Inventory supported nested commands for repo/branch/commit/file/MR/pipeline/project/workitem/search/ssh-key.
- [x] Verify target repository access with `repo view 6a49f26608f52788b13355d2/test-repo`.
- [x] Verify config/auth/api/output commands. Fixed `config get endpoint` to return the effective default when `--scope` is not supplied; explicit scope still reads that scope. Fixed `auth login --endpoint` token verification to use the requested endpoint.
- [x] Verify repository commands. `repo set-default` verified against `7123450`; repository write paths were fixed to the current organization API. Disposable `repo create` now reaches the current API but is blocked by token permission, so update/archive/delete were not run against real repositories.
- [x] Verify branch/commit/file commands against `test-repo`. Fixed organization-aware code paths, `refName` commit listing, `/files/tree`, current response decoding, and base64 file content display.
- [x] Create or update a branch in `test-repo` for live MR testing; pushed `codex-live-test-20260705-210101` via SSH from `tmp/test-repo-live-20260705-210101`.
- [x] Create, inspect, comment/review, close/reopen, and merge a merge request in `test-repo`; created and merged live MR `!2` from `codex-live-test-20260705-210101-cli` to `master`.
- [x] Verify code review comments in `test-repo`; created live MR `!3` from `codex-live-review-20260705-212811` and verified direct comment, uploaded image reference, inline line review, resolving a review comment, and marking it unresolved again.
- [x] Verify search commands against `test-repo` and the sample Projex project. Repository/code/commit/MR/workitem search verified.
- [x] Verify pipeline/run commands where runnable pipeline data exists. Fixed organization-aware pipeline/run list/view/log/watch paths, all-organization fallback, current response decoding, job-log path, and uppercase terminal status handling. Pipeline/run write commands were skipped because the target organization has no disposable pipeline/run and visible runs belong to other projects.
- [x] Verify project/workitem commands. Fixed `project list/view/member/label`, workitem list/search, workitem create/view/edit/activity/delete, organization fallback, current response decoding, required workitem category fan-out, and workitem delete `--yes`. Iteration/milestone list now reaches the current API but is blocked by token permission.
- [x] Verify SSH key commands without deleting real user keys. Fixed organization-aware key list paths, current response decoding, all-organization de-duplication, and verified add/delete with a disposable generated key.
- [x] Verify local-only extension, alias, and completion commands.
- [x] Rebuild `dist/yunxiao-darwin-arm64` and run `go test ./...`.

## Detailed Checklist

### Config, Auth, API, Output

- [x] `config list`
- [x] `config get endpoint`
- [x] `config set` using an isolated config path or reversible value
- [x] `auth login`; prompt verified token creation link and permission text, `--with-token` verified with an isolated credential store so real credentials were not overwritten.
- [x] `auth status`
- [x] `auth logout`; verified with an isolated credential store so the real credential was not deleted.
- [x] `api GET /oapi/v1/platform/user`
- [x] `--json`
- [x] `--jq`
- [x] `--template`
- [x] `--plain`
- [x] `--verbose`
- [x] `--version`

### Repository

- [x] `repo list`
- [x] `repo view 6a49f26608f52788b13355d2/test-repo`
- [x] `repo view test-repo`
- [x] `repo set-default 7123450`
- [n/a] `repo create` on disposable test repo; current token receives 403 `Current token has no permission to api` after switching to the current organization path and documented body (`name` + `path`).
- [n/a] `repo update` on disposable test repo; skipped because disposable repository creation is blocked by token permission.
- [n/a] `repo archive` on disposable test repo; skipped because disposable repository creation is blocked by token permission.
- [x] `repo unarchive` unsupported-path behavior
- [n/a] `repo delete` on disposable test repo; skipped because disposable repository creation is blocked by token permission.

### Code Browsing

- [x] `branch list`
- [x] `branch view master`
- [x] `commit list`
- [x] `commit view c862c200d139d9e19469bfa769f237d51f06daed`
- [x] `file tree --ref master`
- [x] `file view README.md --ref master`
- [x] code review comments are covered through merge requests; verified direct comment, uploaded image reference, and inline review on live MR `!3`.

### Merge Requests

- [x] `mr list`
- [x] `mr create`; verified with live MR `!2` from `codex-live-test-20260705-210101-cli` to `master`.
- [x] `mr view`; verified against live MR `!2`.
- [x] `mr diff`; fixed current API patch-set handling and verified against live MR `!2`.
- [x] `mr files`; fixed current API patch-set handling and verified against live MR `!2`.
- [x] `mr comment`; fixed current `CreateChangeRequestComment` payload and verified direct global comment on live MR `!3` (`6d56cad3417a4c5e877282a2491b0182`).
- [x] `mr comment` with uploaded image reference; uploaded `codex-live-review-20260705-212811.svg` to branch `codex-live-review-20260705-212811` and verified image markdown comment on live MR `!3` (`6a7e1b37a4c0472e8d3feea15b1f3b1c`).
- [x] `mr comment --file --line`; verified inline review on `codex-live-review-20260705-212811.txt:3` in live MR `!3` (`a18a081b772c4dc8a53d170b279e9e1e`).
- [x] `mr resolve-comment`; fixed current comment update payload and verified inline review comment `a18a081b772c4dc8a53d170b279e9e1e` became `resolved: true`.
- [x] `mr unresolve-comment`; verified inline review comment `a18a081b772c4dc8a53d170b279e9e1e` became `resolved: false`.
- [x] `mr approve`; verified against live MR `!2`.
- [x] `mr changes-requested`; verified against live MR `!2`.
- [x] `mr status`; verified against live MR `!2`.
- [x] `mr edit`; verified against live MR `!2`.
- [x] `mr close`; verified against live MR `!2`.
- [x] `mr reopen`; verified against live MR `!2`.
- [x] `mr merge`; verified against live MR `!2`, merged into `master` as `00ef5d2544ff752883df6ae843329adaaf48f646`.

### Pipeline And Runs

- [x] `pipeline list`
- [x] `pipeline view 5047627`
- [n/a] `pipeline run <id>`; skipped because target organization `6a49f26608f52788b13355d2` has no pipelines and visible pipelines belong to other projects.
- [x] `run list --pipeline 5047627`
- [x] `run view 13 --pipeline 5047627`; `run view 13` now returns `--pipeline is required` because run IDs are not globally unique.
- [x] `run log 13 --pipeline 5047627`; verified current job-log API path, this historical run returned empty log content.
- [x] `run watch 13 --pipeline 5047627`
- [n/a] `run cancel/retry/task actions`; skipped because no disposable run is available and the visible runs are historical runs from other projects.

### Project And Work Items

- [x] `project list`
- [x] `project view 259be5fb2d81a291e0883d6b50`
- [x] `project member list 259be5fb2d81a291e0883d6b50`
- [n/a] `project iteration list/view`; current token receives 403 `Current token has no permission to api` after switching to the current organization path.
- [n/a] `project milestone list/view`; current token receives 403 `Current token has no permission to api` after switching to the current organization path.
- [x] `project label list 259be5fb2d81a291e0883d6b50`
- [x] `workitem list --project-id 259be5fb2d81a291e0883d6b50`
- [x] `workitem create --project-id 259be5fb2d81a291e0883d6b50 --type Task --assignee 684fb271e5683b38471884c1`
- [x] `workitem view fd95c1c10081250b1bdcd401d3`
- [x] `workitem edit fd95c1c10081250b1bdcd401d3`
- [x] `workitem activity fd95c1c10081250b1bdcd401d3`
- [x] `workitem delete fd95c1c10081250b1bdcd401d3 --yes`; verified raw `logicalStatus` became `RECYCLED`.

### Search

- [x] `search repo test-repo`
- [x] `search code 测试 --repo 7123450 --ref master`
- [x] `search commit README --repo 7123450`
- [x] `search mr test --repo 7123450`
- [x] `search workitem DEMO --project-id 259be5fb2d81a291e0883d6b50`

### SSH Key

- [x] `ssh-key list`
- [x] `ssh-key add` using disposable generated public key, title `yunxiao-cli-live-test-20260705-182633`, ID `1072258`
- [x] `ssh-key delete 1072258 --yes`; verified it disappeared from organization key list after a short delay

### Local-Only Commands

- [x] `alias list`
- [x] `alias set`
- [x] `alias delete`
- [x] `completion bash`
- [x] `completion zsh`
- [x] `completion fish`
- [x] `completion powershell`
- [x] `extension list`
- [x] `extension create`
- [x] `extension install` from local scaffold
- [x] `extension exec`
- [x] `extension upgrade` using a disposable Git-managed extension under a temporary `YUNXIAO_CONFIG_DIR`
- [x] `extension remove`
