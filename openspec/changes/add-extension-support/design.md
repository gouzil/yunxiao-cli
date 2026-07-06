## Context

当前 `yunxiao-cli` 是 Go + Cobra 的单二进制 CLI，根命令在 `internal/app/root.go` 中集中注册核心命令，配置和 alias 信息保存在 `internal/config` 的全局配置文件中，终端 IO 抽象集中在 `internal/terminal`。这给扩展系统提供了三个天然接入点：

```text
用户输入
  │
  ▼
yunxiao root command
  │
  ├─ core command: auth/repo/mr/pipeline/...
  ├─ extension management: extension/ext
  ├─ alias expansion: 已有配置模型，扩展系统读取 alias 配置并保持 alias 优先级
  └─ installed extension wrapper: yunxiao <name> ...
       │
       ▼
    exec absolute/path/to/yunxiao-<name> args...
```

`gh` 的扩展机制是最直接的参考：扩展仓库以 `gh-` 开头，仓库根目录包含同名可执行文件，`gh <extname>` 会把参数转发给该可执行文件，核心命令不能被扩展覆盖，冲突时可使用 `gh extension exec <name>` 显式执行扩展。相关参考：

- `gh extension` manual: https://cli.github.com/manual/gh_extension
- `gh extension install`: https://cli.github.com/manual/gh_extension_install
- `gh extension exec`: https://cli.github.com/manual/gh_extension_exec
- 创建 GitHub CLI 扩展: https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions
- 使用 GitHub CLI 扩展: https://docs.github.com/en/github-cli/github-cli/using-github-cli-extensions
- `gh` root 注册扩展和 alias 的源码: https://github.com/cli/cli/blob/trunk/pkg/cmd/root/root.go
- `gh` extension manager 源码: https://github.com/cli/cli/blob/trunk/pkg/cmd/extension/manager.go
- `gh` extension Cobra wrapper 源码: https://github.com/cli/cli/blob/trunk/pkg/cmd/root/extension.go
- `gh` extension command 源码: https://github.com/cli/cli/blob/trunk/pkg/cmd/extension/command.go
- `gh` extension interface 源码: https://github.com/cli/cli/blob/trunk/pkg/extensions/extension.go
- `gh` update notifier 源码: https://github.com/cli/cli/blob/trunk/internal/update/update.go

云效 CLI 不能简单照搬 `gh` 的全部机制。`gh extension install` 会先通过 GitHub Releases 判断是否存在预编译二进制资产，失败后再 clone 仓库；这依赖 GitHub API、release 语义和资产命名。云效 CLI 首期应采用云平台中立的 Git clone 和本地目录模型，避免把 GitHub Releases 或云效制品能力硬塞进扩展安装。

## Goals / Non-Goals

**Goals:**

- 提供 `yunxiao extension` 管理命令，别名为 `extensions` 和 `ext`。
- 支持 `yunxiao extension install <source>` 安装扩展，`<source>` 固定解析为 full Git URL 或本地目录路径。
- 支持 `yunxiao extension list/remove/upgrade/exec/create` 的首期闭环。
- 支持 `yunxiao <name> ...` 直接运行已安装扩展，并将 stdin/stdout/stderr 和 argv 透传给扩展可执行文件。
- 扩展运行协议必须语言无关；CLI 只负责执行当前平台上的同名可执行入口，不限制扩展由 Python、Node.js、Ruby、Rust、Go、shell 还是其他语言实现。
- 确保核心命令优先，扩展不能覆盖核心命令；冲突扩展仍能通过 `yunxiao extension exec <name>` 运行。
- 定义稳定的安装目录、状态目录、manifest、更新检测和跨平台执行规则。
- 给扩展作者提供最低限度的环境变量契约，允许扩展通过 `yunxiao api` 复用认证，而不是把 token 明文注入环境。
- 扩展鉴权固定采用宿主代理模式：扩展进程调用 `yunxiao api`，由宿主 CLI 从 keyring 或 credentials 文件读取 PAT 并注入云效 HTTP 请求。
- 保持默认测试无需真实远程仓库和云效凭据。

**Non-Goals:**

- 不实现 GitHub topic 搜索、浏览 TUI、公共扩展市场。
- 不实现基于 GitHub Releases 的预编译二进制下载。
- 不实现云效制品、应用交付、私有 registry 分发；这些能力必须作为后续 capability 单独设计。
- 不给第三方扩展做签名认证、沙箱隔离、权限授权、安全背书。
- 不允许扩展覆盖核心命令、内置 help、completion、auth、config、api 等核心入口。
- 不把扩展 API 做成稳定 Go SDK；扩展首期通过进程协议、环境变量和 `yunxiao api` 集成。

## Decisions

### 1. 扩展命名使用 `yunxiao-<name>`

扩展仓库、本地目录和可执行文件 MUST 使用 `yunxiao-<name>` 命名。用户执行时使用短名：

```text
仓库/目录: yunxiao-team-report
可执行文件: yunxiao-team-report
运行命令:   yunxiao team-report --since 7d
显式执行:   yunxiao extension exec team-report --since 7d
```

名称校验：

- `<name>` MUST 只包含小写字母、数字和 `-`，且不能以 `-` 开头或结尾。
- 安装输入如果是本地目录，目录 basename MUST 是 `yunxiao-<name>`。
- 安装输入如果是远程 Git URL，去掉 `.git` 后的 repo basename MUST 是 `yunxiao-<name>`。
- Windows 上可执行文件固定为 `yunxiao-<name>.exe`。

理由：与 `gh-<name>` 约定一致，降低扩展作者心智成本，同时不会和系统 PATH 里的普通命令混淆。

### 2. 首期只支持 Git clone 和本地目录安装

安装来源分两类：

```text
remote git:
  yunxiao extension install https://codeup.aliyun.com/org/repo/yunxiao-foo.git
  yunxiao extension install git@codeup.aliyun.com:org/repo/yunxiao-foo.git --ref v1.2.0

local:
  yunxiao extension install ./yunxiao-foo
```

远程安装流程：

1. 解析 repo basename 并校验 `yunxiao-` 前缀。
2. clone 到临时目录。
3. 如指定 `--ref`，checkout 到该 tag/branch/commit。
4. 校验仓库根目录存在当前平台的扩展可执行文件；Unix 为 `yunxiao-<name>`，Windows 为 `yunxiao-<name>.exe`。
5. 移动到扩展安装目录。
6. 写入 manifest。

本地安装流程：

1. 校验本地目录 basename 和根目录可执行文件。
2. Unix 平台在安装目录下创建指向本地目录的 symlink。
3. Windows 平台在安装目录下写入包含真实路径的 path file。
4. 写入 local manifest。

理由：这与 `gh` 的 script extension 路径一致，但不依赖 GitHub API。远程仓库里的扩展必须提交可执行入口。解释型语言扩展使用 shebang 入口，例如 `#!/usr/bin/env python3` 或 `#!/usr/bin/env node`；编译型语言扩展提交当前平台可执行产物，作者使用本地安装调试已构建产物。CLI 不自动安装运行时、不自动执行 package manager、不自动编译源码。预编译二进制 release 下载不纳入本变更。

### 3. 使用独立 extension manager 包

新增 `internal/extension` 包，提供可测试接口：

```go
type Manager interface {
    List() ([]Extension, error)
    Install(ctx context.Context, source InstallSource) (Extension, error)
    InstallLocal(ctx context.Context, dir string) (Extension, error)
    Upgrade(ctx context.Context, name string, force bool) error
    Remove(ctx context.Context, name string) error
    Dispatch(ctx context.Context, name string, args []string, io terminal.IOStreams) (handled bool, err error)
    Create(ctx context.Context, name string, template TemplateType) error
    UpdateDir(name string) string
}
```

`Extension` 使用 typed model，不用 `map[string]any`：

```go
type Extension struct {
    Name           string
    FullName       string
    Kind           ExtensionKind // git/local
    ExecutablePath string
    SourceURL      string
    SourceRef      string
    CurrentCommit  string
    Pinned         bool
    LocalPath      string
}
```

manager 需要注入：

- dataDir/stateDir 函数，便于测试和平台路径解析。
- git runner，便于单元测试不用真实网络。
- exec runner，便于测试 argv、env 和退出码。
- shell finder，便于 Windows shebang 执行。
- clock，便于测试 24 小时更新检测。

理由：扩展系统会触碰文件系统、Git、进程执行和跨平台行为，必须通过接口注入降低测试成本。

### 4. 安装目录和 manifest 稳定化

目录布局：

```text
dataDir/
  extensions/
    yunxiao-foo/
      yunxiao-foo
      extension.json

stateDir/
  extensions/
    yunxiao-foo/
      state.yml
```

默认目录：

- 优先尊重 `YUNXIAO_CONFIG_DIR`、`XDG_DATA_HOME`、`XDG_STATE_HOME`。
- macOS/Linux 缺省目录固定为 `~/.local/share/yunxiao/extensions` 和 `~/.local/state/yunxiao/extensions`。
- Windows 缺省目录固定为 `%LocalAppData%\yunxiao\extensions` 和 `%LocalAppData%\yunxiao\state\extensions`。
- data/state 与 config/credentials MUST 分离。

manifest 固定使用 `extension.json`，与当前配置 JSON 保持一致：

```json
{
  "version": 1,
  "name": "foo",
  "fullName": "yunxiao-foo",
  "kind": "git",
  "sourceURL": "https://codeup.aliyun.com/org/repo/yunxiao-foo.git",
  "sourceRef": "v1.2.0",
  "currentCommit": "abc123",
  "pinned": true,
  "executablePath": "/.../extensions/yunxiao-foo/yunxiao-foo"
}
```

理由：manifest 让 `list`、`upgrade` 和错误诊断不必反推 Git remote/path。与 `gh` 的 binary manifest 思路相同，但格式贴合当前 repo 的 JSON 配置。

### 5. 根命令注册顺序采用核心优先

构建根命令时按以下顺序注册：

```text
1. core commands
2. extension management command: extension/extensions/ext
3. alias commands and alias expansion entries
4. installed extension wrappers that do not conflict with existing root commands
```

扩展 wrapper：

- `Use` 为短名 `<name>`。
- `DisableFlagParsing = true`，保证扩展自己的 flags 不被核心 CLI 解析。
- `RunE` 调用 `Manager.Dispatch`。
- wrapper 不做认证前置检查，因为扩展可以是纯本地工具；需要云效认证时，由扩展调用 `yunxiao api` 复用认证。

冲突规则：

- 如果短名与核心命令、`extension`、`ext`、`extensions`、alias 冲突，根命令不注册该扩展 wrapper。
- `yunxiao extension list` MUST 显示冲突状态。
- 用户可执行 `yunxiao extension exec <name> ...` 绕过冲突。

理由：这保留 `gh` 的安全边界，也避免扩展安装后改变核心命令行为。

### 6. 执行模型只运行绝对路径

执行扩展时：

- 通过 manifest 解析可执行文件绝对路径。
- 不通过 `PATH` 搜索扩展名，避免 PATH hijack。
- 不解析扩展源码语言；扩展文件由操作系统、shebang 或 Windows 执行规则决定如何启动。
- stdin/stdout/stderr 原样连接到扩展进程。
- `argv[0]` 为扩展可执行文件路径，后续 args 原样来自用户输入。
- 外部进程退出码 MUST 透传为 `yunxiao` 进程退出码。
- `exec.ErrNotFound`、权限不足、缺少可执行位、manifest 损坏等错误使用统一错误格式输出。

环境变量：

```text
YUNXIAO_EXTENSION=1
YUNXIAO_EXTENSION_NAME=<name>
YUNXIAO_EXTENSION_DIR=<installed-dir-or-local-dir>
YUNXIAO_ENDPOINT=<resolved endpoint>
YUNXIAO_ORGANIZATION=<resolved organization, if any>
YUNXIAO_PROJECT=<resolved project, if any>
YUNXIAO_REPO=<resolved repo, if any>
```

认证 token 不得注入环境变量。扩展如需调用云效 API，应执行 `yunxiao api ...`，复用用户本机 credential store 和现有脱敏逻辑。

鉴权传递固定为：

```text
extension process
  │
  │ exec: yunxiao api GET /...
  ▼
yunxiao api
  │
  │ read token from keyring / credentials.json
  │ inject Authorization into HTTP request
  ▼
Yunxiao OpenAPI
```

扩展进程只能获得 endpoint、organization、project、repo 等非敏感上下文环境变量。PAT 不通过 argv、stdin、stdout、stderr、环境变量或临时文件传递给扩展。

`yunxiao api` 的脚本化输出契约固定为：

- 默认 stdout 只输出云效 OpenAPI 原始响应体；云效返回 JSON 时，stdout 就是可直接解析的 JSON。
- `--verbose` 的请求摘要、HTTP 状态和 request id 必须输出到 stderr，不得污染 stdout。
- `--jq <expr>` 必须直接作用于响应体 JSON，而不是作用于 `RawAPIResult.body` 字符串包装。
- `--template <template>` 必须直接作用于响应体 JSON 对象；非 JSON 响应使用原始文本。
- `--json <fields>` 不作为 `yunxiao api` 的主要脚本接口；扩展脚本使用默认原始 JSON 或 `--jq`。

Windows 行为：

- 优先执行 `.exe`。
- 如果安装的是脚本且存在 shebang，固定使用 `sh.exe -c 'command "$@"'` 方式转发，缺少 `sh.exe` 时提示安装 Git for Windows。
- path file local install 需要读取目标目录后再解析可执行文件。

### 7. 升级和更新提示按来源区分

`yunxiao extension upgrade`：

- git 扩展：如果未 pin，执行 `git pull --ff-only`；如果指定 `--force`，执行 `git fetch origin HEAD` 后 reset 到 `FETCH_HEAD`。
- pinned git 扩展：默认拒绝升级；用户必须使用 `--force` 解除 pin 后更新到远端 HEAD。
- local 扩展：不可升级，提示它由本地目录管理。
- 升级前清理对应 stateDir，避免旧更新检查状态影响新来源。

直接执行扩展时的更新提示：

- 仅在 stdout/stderr 都是 TTY、非 CI、未设置 `YUNXIAO_NO_EXTENSION_UPDATE_NOTIFIER` 时启用。
- 每个扩展最多 24 小时检查一次。
- git 扩展通过 `git ls-remote` 比较远端 HEAD 和本地 HEAD。
- 检查在后台执行；如果扩展命令结束前结果未返回，不阻塞用户。

理由：保留 `gh` 的低打扰体验，但实现不依赖 release API。

### 8. 扩展创建提供轻量模板

`yunxiao extension create <name>` 生成 `yunxiao-<name>` 目录。首期模板：

- `--template script`：生成可执行 shell 脚本入口。
- `--template go`：生成 `main.go`、`go.mod`、`.gitignore` 和本地 build 说明，但不自动发布。

默认模板为 `script`。`go` 模板可以调用本机 `go mod init` 和 `go build`，但失败时应保留已生成文件并提示用户手动继续。

理由：扩展生态要成立，需要低成本 authoring。模板只负责脚手架，不把发布、制品和 registry 纳入首期。

## Risks / Trade-offs

- [Risk] 远程 Git 安装要求仓库根目录提交可执行文件，Go 扩展作者可能不想提交二进制。→ Mitigation：支持 local install 和 `extension create --template go`，二进制分发不纳入本变更。
- [Risk] 扩展是任意本地程序，可能读取文件并调用网络。→ Mitigation：安装和升级时显示未验证来源提示，默认 TTY 确认，文档要求审计源码。
- [Risk] alias 当前实现可能只有管理命令，没有完整执行语义。→ Mitigation：本变更不补 alias 执行语义；扩展冲突规则读取 alias 配置并禁止同名 wrapper。
- [Risk] Windows symlink 和脚本执行差异大。→ Mitigation：Windows local install 固定使用 path file，脚本执行固定依赖 Git for Windows 的 `sh.exe` 并给出明确错误。
- [Risk] 后台更新检查可能引入慢命令和网络噪音。→ Mitigation：只在 TTY/非 CI 下启用，24 小时节流，不阻塞扩展退出。
- [Risk] `DisableFlagParsing` 会让 root-level flags 放在扩展名之后时被扩展接收，而不是被核心 CLI 解析。→ Mitigation：文档明确 `yunxiao foo --bar` 中 `--bar` 属于扩展；核心上下文通过环境变量传递。
- [Risk] 远程 Git URL 支持范围过宽，可能导致 URL 解析边界复杂。→ Mitigation：首期只要求 full Git URL 和本地 path；不做 `owner/repo` shorthand。

## Migration Plan

1. 新增 extension manager 和命令，不改变现有核心命令行为。
2. 根命令仅为不冲突扩展注册 wrapper，冲突扩展保持不可直接覆盖。
3. `extension` 能力上线后，已有用户配置无需迁移。
4. 后续引入 registry/binary release 时，必须在 manifest 中提升 version，并保留读取 v1 manifest 的能力。

Rollback 策略：

- 如扩展系统出现严重问题，回滚操作固定为从根命令移除 installed extension wrapper 注册，保留 `extension list/remove` 用于清理已安装扩展。
- manifest 和安装目录不影响核心命令，可由用户手动删除 dataDir/stateDir。

## Fixed Scope Decisions

- 本变更只支持 full Git URL 和本地目录安装，不支持企业内部官方扩展 registry。
- 本变更不提供 `YUNXIAO_EXTENSION_PATH`、`yunxiao extension env`。
- 本变更不补齐 alias 执行语义，只处理扩展与已配置 alias 的冲突。
- 本变更不设计二进制分发，后续必须另建 OpenSpec change。
