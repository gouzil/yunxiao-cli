## 输出契约原则

本文档定义每个 CLI 操作的默认人类可读输出。实现阶段必须同时支持结构化输出：

- 默认输出面向终端阅读，风格参考 `gh`：列表用表格，详情用标题 + 字段，写操作用明确成功/失败提示。
- `--json <fields>` MUST 输出合法 JSON，字段名使用 camelCase。
- `--jq <expr>` MUST 只作用于 JSON 输出。
- `--template <tmpl>` MUST 作用于结构化模型，而不是解析人类可读文本。
- `--web` 成功打开浏览器时可只输出 URL 或静默；无图形环境 MUST 输出 URL。
- token、secret、credential 默认 MUST 脱敏，只显示前缀和固定星号。
- 当云效 API 不返回某字段时，默认输出 MUST 显示 `unknown` 或省略该字段；不得伪造数据。
- 删除、合并、关闭等破坏性操作在非 `--yes` 模式下 MUST 先确认。

通用空结果：

```text
No <resource> found.
```

通用错误：

```text
X <summary>
  - Request ID: <requestId>
  - HTTP status: <status>
  - Code: <code>
  - Message: <message>
```

## 1. auth / config / api

### `yunxiao auth login`

交互提示：

```text
Yunxiao endpoint [openapi-rdc.aliyuncs.com]:
Paste your Yunxiao personal access token:
```

输入 token 后的校验状态：

```text
- Verifying token with <endpoint>...
```

成功：

```text
✓ Logged in to <endpoint> account <displayName> (keyring)
- Active account: true
- Organization: <organizationName|organizationId|unknown>
- Git operations protocol: https
- Token: <prefix>********************************
- Token scopes: <scope1>, <scope2> | unknown
```

如果凭据保存到权限受限文件而不是系统 keyring：

```text
✓ Logged in to <endpoint> account <displayName> (config file)
- Active account: true
- Organization: <organizationName|organizationId|unknown>
- Git operations protocol: https
- Token: <prefix>********************************
- Token scopes: <scope1>, <scope2> | unknown
```

失败：

```text
X Authentication failed for <endpoint>
  - Message: <message>
```

非交互模式从 stdin 读取 token：

```bash
echo "$YUNXIAO_TOKEN" | yunxiao auth login --with-token
```

成功输出与交互模式一致。若 stdin 为空：

```text
X Authentication failed for <endpoint>
  - Message: token is required on stdin when using --with-token
```

不得支持将 token 作为普通命令行参数明文传入；帮助信息 MUST 引导用户使用交互输入、stdin 或环境变量。

### `yunxiao auth status`

参考 `gh auth status`，按 endpoint 分组：

```text
<endpoint>
  ✓ Logged in to <endpoint> account <displayName> (keyring)
  - Active account: true
  - User ID: <userId>
  - Organization: <organizationName|organizationId|unknown>
  - Git operations protocol: https
  - Token: <prefix>********************************
  - Token scopes: <scope1>, <scope2> | unknown
```

未登录：

```text
<endpoint>
  X Not logged in to <endpoint>
  - Run: yunxiao auth login
```

### `yunxiao auth logout`

```text
✓ Logged out of <endpoint> account <displayName>
```

### `yunxiao config list`

```text
KEY                   VALUE                         SOURCE
endpoint              openapi-rdc.aliyuncs.com      global
organization          66f0...                       repo
repo                  my-group/my-repo              repo
git_protocol          https                         global
```

### `yunxiao config get <key>`

```text
<value>
```

### `yunxiao config set <key> <value>`

```text
✓ Set <key> to <value> (<scope>)
```

### `yunxiao api <method> <path>`

默认输出原始响应体：

```json
{
  "requestId": "...",
  "success": true,
  "data": {}
}
```

`--verbose` 时在响应体前输出请求摘要：

```text
> GET https://<endpoint>/<path>
< HTTP/2 200
< x-acs-request-id: <requestId>

<response body>
```

## 2. repo / branch / commit / file / ssh-key

### `yunxiao repo list`

```text
ID        NAME        PATH                 DEFAULT BRANCH  VISIBILITY  ARCHIVED  UPDATED
12345     api         cro/api              master          private     false     2026-06-09 10:31
```

### `yunxiao repo view <repo>`

```text
<namespace>/<name>
  - ID: <repositoryId>
  - Name: <name>
  - Path: <path>
  - Default branch: <branch>
  - Visibility: <visibility>
  - Archived: <true|false>
  - Created: <time>
  - Updated: <time>
  - Web URL: <url>
  - SSH URL: <url|unknown>
  - HTTP URL: <url|unknown>
```

### `yunxiao repo create <name>`

```text
✓ Created repository <namespace>/<name>
- ID: <repositoryId>
- Default branch: <branch>
- Web URL: <url>
```

### `yunxiao repo update <repo>`

```text
✓ Updated repository <namespace>/<name>
```

### `yunxiao repo archive <repo>`

```text
✓ Archived repository <namespace>/<name>
```

### `yunxiao repo unarchive <repo>`

首期若 API 未确认：

```text
X repo unarchive is not supported by the current Yunxiao API mapping
  - Status: pending API confirmation
```

确认支持后：

```text
✓ Unarchived repository <namespace>/<name>
```

### `yunxiao repo delete <repo>`

```text
✓ Deleted repository <namespace>/<name>
```

### `yunxiao repo transfer <repo>`

```text
✓ Transferred repository <namespace>/<name> to <targetNamespace>
```

### `yunxiao repo set-default <repo>`

```text
✓ Set default repository to <namespace>/<name>
- Scope: repository
- Config: <path>
```

### `yunxiao branch list`

```text
NAME            DEFAULT  PROTECTED  COMMIT      UPDATED
master          true     true       a1b2c3d     2026-06-09 10:31
feature/demo    false    false      e4f5g6h     2026-06-08 18:20
```

### `yunxiao branch view <branch>`

```text
<branch>
  - Repository: <namespace>/<repo>
  - Default: <true|false>
  - Protected: <true|false|unknown>
  - Commit: <sha>
  - Author: <name>
  - Updated: <time>
```

### `yunxiao branch create <branch>`

```text
✓ Created branch <branch> from <ref>
```

### `yunxiao branch delete <branch>`

```text
✓ Deleted branch <branch>
```

### `yunxiao commit list`

```text
SHA        AUTHOR        DATE                 TITLE
a1b2c3d    gouzi         2026-06-09 10:31     fix: update cli output
```

### `yunxiao commit view <sha>`

```text
commit <sha>
Author: <name> <<email>>
Date:   <time>

    <title>

<body>

Statuses:
  - <context>: <state> (<targetUrl|none>)
```

### `yunxiao commit comment <sha>`

```text
✓ Commented on commit <shortSha>
- Comment ID: <commentId>
```

### `yunxiao file tree <path>`

```text
TYPE    PATH                         SIZE
dir     src                          -
file    src/main.go                  1932
```

### `yunxiao file view <path>`

默认直接输出文件内容：

```text
<file content>
```

`--metadata`：

```text
<path>
  - Ref: <ref>
  - Blob ID: <blobId>
  - Size: <bytes>
  - Encoding: <encoding>
```

### `yunxiao file blame <path>`

```text
LINE      SHA        AUTHOR       DATE          CONTENT
1         a1b2c3d    gouzi        2026-06-09    package main
```

### `yunxiao file create/update/delete`

```text
✓ <Created|Updated|Deleted> file <path>
- Branch: <branch>
- Commit: <sha>
```

### `yunxiao file commit-multiple`

```text
✓ Committed <count> file changes
- Branch: <branch>
- Commit: <sha>
```

### `yunxiao ssh-key list`

```text
ID        TITLE             FINGERPRINT                  CREATED
987       MacBook Pro       SHA256:abc123...             2026-06-01 09:12
```

### `yunxiao ssh-key view <key-id>`

```text
SSH key <key-id>
  - Title: <title>
  - Fingerprint: <fingerprint>
  - Created: <time>
  - Key: <publicKey>
```

### `yunxiao ssh-key add <public-key-file>`

```text
✓ Added SSH key <title>
- ID: <keyId>
- Fingerprint: <fingerprint>
```

### `yunxiao ssh-key delete <key-id>`

```text
✓ Deleted SSH key <key-id>
```

## 3. mr

### `yunxiao mr list`

```text
ID     STATE    TITLE                         SOURCE        TARGET    AUTHOR    UPDATED
42     open     Add CLI output contract       feature/x     master    gouzi     2026-06-09 10:31
```

### `yunxiao mr view <mr>`

```text
!<id> <title>
  - State: <state>
  - Author: <name>
  - Source: <sourceBranch>
  - Target: <targetBranch>
  - Review: <approved|changes_requested|pending|unknown>
  - Mergeable: <true|false|unknown>
  - Pipeline: <state|unknown>
  - Labels: <labels|none>
  - Web URL: <url>

Description:
<description>
```

### `yunxiao mr create`

```text
✓ Created merge request !<id>
- Title: <title>
- Source: <sourceBranch>
- Target: <targetBranch>
- Web URL: <url>
```

### `yunxiao mr edit`

```text
✓ Updated merge request !<id>
```

### `yunxiao mr close`

```text
✓ Closed merge request !<id>
```

### `yunxiao mr reopen`

```text
✓ Reopened merge request !<id>
```

### `yunxiao mr merge`

```text
✓ Merged merge request !<id>
- Merge commit: <sha|unknown>
```

### `yunxiao mr files`

```text
STATUS     ADDITIONS  DELETIONS  PATH
modified   24         3          internal/cmd/mr.go
added      80         0          internal/output/mr.go
```

### `yunxiao mr diff`

默认输出 unified diff：

```diff
diff --git a/internal/cmd/mr.go b/internal/cmd/mr.go
...
```

### `yunxiao mr comment`

```text
✓ Commented on merge request !<id>
- Comment ID: <commentId>
```

### `yunxiao mr comment edit`

```text
✓ Updated comment <commentId> on merge request !<id>
```

### `yunxiao mr comment delete`

```text
✓ Deleted comment <commentId> from merge request !<id>
```

### `yunxiao mr approve`

```text
✓ Approved merge request !<id>
```

### `yunxiao mr changes-requested`

```text
✓ Requested changes on merge request !<id>
```

### `yunxiao mr labels attach`

```text
✓ Attached labels to merge request !<id>
- Labels: <label1>, <label2>
```

### `yunxiao mr labels list`

```text
NAME        COLOR      DESCRIPTION
bug         #d73a4a    Something is not working
```

### `yunxiao mr status`

```text
Merge request !<id>
  - State: <state>
  - Mergeable: <true|false|unknown>
  - Review: <approved|changes_requested|pending|unknown>
  - Conflicts: <true|false|unknown>
  - Pipeline: <success|failed|running|pending|unknown>
  - Checks:
    ✓ <name> <status>
    X <name> <status>
```

## 4. pipeline / run

### `yunxiao pipeline list`

```text
ID        NAME                 STATUS    UPDATED
1001      build-and-test       enabled   2026-06-09 10:31
```

### `yunxiao pipeline view <pipeline-id>`

```text
Pipeline <pipeline-id>: <name>
  - Status: <enabled|disabled|unknown>
  - Creator: <name>
  - Updated: <time>
  - Web URL: <url>

Jobs:
  - <jobName> (<jobId>)
```

### `yunxiao pipeline create`

```text
✓ Created pipeline <name>
- ID: <pipelineId>
- Web URL: <url>
```

### `yunxiao pipeline update`

```text
✓ Updated pipeline <pipelineId>
```

### `yunxiao pipeline delete`

```text
✓ Deleted pipeline <pipelineId>
```

### `yunxiao pipeline run`

```text
✓ Started pipeline <pipelineId>
- Run ID: <runId>
- Branch: <branch>
- Web URL: <url>
```

### `yunxiao pipeline artifact-url`

```text
Artifact URL:
<url>
```

### `yunxiao run list`

```text
ID        PIPELINE             BRANCH      STATUS     STARTED              DURATION
9988      build-and-test       master      running    2026-06-09 10:31     2m14s
```

### `yunxiao run view <run-id>`

```text
Run <runId>
  - Pipeline: <pipelineName> (<pipelineId>)
  - Status: <status>
  - Trigger: <triggerMode>
  - Triggered by: <name>
  - Branch: <branch>
  - Started: <time>
  - Finished: <time|running>
  - Duration: <duration>
  - Web URL: <url>

Jobs:
  ✓ build     success   1m20s
  - test      running   54s
```

### `yunxiao run cancel`

```text
✓ Canceled run <runId>
```

### `yunxiao run log`

默认输出日志正文：

```text
<log lines>
```

带任务摘要：

```text
==> <jobName> (<jobId>)
<log lines>
```

### `yunxiao run watch`

运行中刷新：

```text
Refreshing run status every <interval>s. Press Ctrl+C to quit.

Run <runId> <status>
  ✓ build     success
  - test      running
```

终态：

```text
✓ Run <runId> completed with status <success>
```

失败终态：

```text
X Run <runId> completed with status <failed>
```

### `yunxiao run retry` / `yunxiao run retry-task`

```text
✓ Retried <runId|jobId>
- New status: <status>
```

### `yunxiao run stop-task`

```text
✓ Stopped job <jobId> in run <runId>
```

### `yunxiao run skip-task`

```text
✓ Skipped job <jobId> in run <runId>
```

### `yunxiao run execute-task`

```text
✓ Executed job <jobId> in run <runId>
```

### `yunxiao run validate pass`

```text
✓ Passed manual validation <jobId> in run <runId>
```

### `yunxiao run validate refuse`

```text
✓ Refused manual validation <jobId> in run <runId>
```

## 5. project / workitem

### `yunxiao project list`

```text
ID        NAME             STATUS    OWNER      UPDATED
p-100     CRO API          active    gouzi      2026-06-09 10:31
```

### `yunxiao project view <project>`

```text
Project <projectId>: <name>
  - Status: <status>
  - Owner: <name|unknown>
  - Members: <count|unknown>
  - Iterations: <count|unknown>
  - Web URL: <url>
```

### `yunxiao project create`

```text
✓ Created project <name>
- ID: <projectId>
- Web URL: <url>
```

### `yunxiao project update`

```text
✓ Updated project <projectId>
```

### `yunxiao project delete`

```text
✓ Deleted project <projectId>
```

### `yunxiao project member list`

```text
USER ID      NAME       ROLE        JOINED
u-100        gouzi      admin       2026-06-01
```

### `yunxiao project member add`

```text
✓ Added member <user> to project <projectId>
```

### `yunxiao project member delete`

```text
✓ Removed member <user> from project <projectId>
```

### `yunxiao project iteration list`

```text
ID        NAME          STATUS     START        END
s-1       2026-W24      active     2026-06-08   2026-06-14
```

### `yunxiao project iteration view`

```text
Iteration <iterationId>: <name>
  - Status: <status>
  - Start: <date>
  - End: <date>
  - Project: <projectId>
```

### `yunxiao project iteration create/update`

```text
✓ <Created|Updated> iteration <name>
- ID: <iterationId>
```

### `yunxiao project milestone list`

```text
ID        NAME              STATUS     DUE
m-1       Public beta       open       2026-07-01
```

### `yunxiao project milestone create/update/delete`

```text
✓ <Created|Updated|Deleted> milestone <name|milestoneId>
```

### `yunxiao project label list`

```text
ID        NAME        COLOR      DESCRIPTION
l-1       backend     #0366d6    Backend work
```

### `yunxiao project label create/update`

```text
✓ <Created|Updated> label <name>
```

### `yunxiao project version list`

```text
ID        NAME        STATUS     RELEASE DATE
v-1       1.0.0       planning   2026-07-01
```

### `yunxiao workitem list`

```text
ID        TYPE      STATE       TITLE                         ASSIGNEE   UPDATED
WI-123    Bug       Open        Login redirects incorrectly   gouzi      2026-06-09
```

### `yunxiao workitem view <workitem>`

```text
Work item <id>: <title>
  - Type: <type>
  - State: <state>
  - Assignee: <name|unassigned>
  - Reporter: <name|unknown>
  - Project: <projectId>
  - Iteration: <iteration|none>
  - Priority: <priority|unknown>
  - Updated: <time>
  - Web URL: <url>

Description:
<description>
```

### `yunxiao workitem create`

```text
✓ Created work item <id>
- Title: <title>
- Type: <type>
- Web URL: <url>
```

### `yunxiao workitem edit`

```text
✓ Updated work item <id>
```

### `yunxiao workitem delete`

```text
✓ Deleted work item <id>
```

### `yunxiao workitem activity`

```text
TIME                 ACTOR      ACTION
2026-06-09 10:31     gouzi      changed state from Open to Closed
```

### `yunxiao workitem comments`

```text
ID        AUTHOR     CREATED              BODY
c-1       gouzi      2026-06-09 10:31     Looks good
```

### `yunxiao workitem comment`

```text
✓ Commented on work item <id>
- Comment ID: <commentId>
```

## 6. search

### `yunxiao search repo`

```text
ID        NAME        PATH                 UPDATED
12345     api         cro/api              2026-06-09 10:31
```

### `yunxiao search code`

```text
REPOSITORY       REF       PATH                    LINE  MATCH
cro/api          master    internal/cmd/root.go    42    cobra.Command
```

### `yunxiao search commit`

```text
REPOSITORY       SHA        AUTHOR      DATE          TITLE
cro/api          a1b2c3d    gouzi       2026-06-09    fix cli output
```

### `yunxiao search mr`

```text
REPOSITORY       ID     STATE    TITLE                         UPDATED
cro/api          42     open     Add CLI output contract       2026-06-09
```

### `yunxiao search workitem`

```text
ID        TYPE      STATE     TITLE                         PROJECT
WI-123    Bug       Open      Login redirects incorrectly   CRO API
```

### `yunxiao search project`

```text
ID        NAME             STATUS    UPDATED
p-100     CRO API          active    2026-06-09
```

## 7. output / completion / alias / browse

### `--json <fields>`

```json
{
  "id": "12345",
  "name": "api"
}
```

### `--jq <expression>`

```text
<jq result>
```

### `--template <template>`

```text
<rendered template>
```

### `--web`

```text
Opening <url> in your browser.
```

无图形环境：

```text
<url>
```

### `yunxiao completion <shell>`

默认输出 shell completion 脚本：

```text
<completion script>
```

### `yunxiao alias list`

```text
ALIAS     EXPANSION
co        mr checkout
```

### `yunxiao alias set <alias> <expansion>`

```text
✓ Added alias <alias>: <expansion>
```

### `yunxiao alias delete <alias>`

```text
✓ Deleted alias <alias>
```
