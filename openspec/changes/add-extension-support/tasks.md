## 1. 参考资料和边界确认

- [x] 1.1 对照 `gh extension` manual、install 和 exec 行为，确认 `yunxiao extension` 首期只吸收管理命令、命名约定、参数转发和冲突处理：https://cli.github.com/manual/gh_extension、https://cli.github.com/manual/gh_extension_install、https://cli.github.com/manual/gh_extension_exec
- [x] 1.2 对照 GitHub 官方扩展使用和创建文档，确认安全提示、local install、script extension 和 authoring 模板的用户体验：https://docs.github.com/en/github-cli/github-cli/using-github-cli-extensions、https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions
- [x] 1.3 对照 `gh` root 注册源码，记录核心命令、extension wrapper、alias 和官方 stub 的注册顺序结论：https://github.com/cli/cli/blob/trunk/pkg/cmd/root/root.go
- [x] 1.4 对照 `gh` extension manager、extension wrapper、extension command 和 interface 源码，确认 manager 接口、Dispatch、DisableFlagParsing、manifest 和 upgrade 的可复用设计点：https://github.com/cli/cli/blob/trunk/pkg/cmd/extension/manager.go、https://github.com/cli/cli/blob/trunk/pkg/cmd/root/extension.go、https://github.com/cli/cli/blob/trunk/pkg/cmd/extension/command.go、https://github.com/cli/cli/blob/trunk/pkg/extensions/extension.go
- [x] 1.5 对照 `gh` update notifier 源码，确认 24 小时节流、TTY/CI 判断和非阻塞提示是否按设计实现：https://github.com/cli/cli/blob/trunk/internal/update/update.go
- [x] 1.6 核对当前 repo 的 `internal/app/root.go`、`internal/config/config.go`、`internal/terminal/terminal.go` 和 alias 行为，明确扩展接入点和 alias 冲突处理方式。

## 2. 扩展数据模型和路径基础

- [x] 2.1 新增 `internal/extension` 包，定义 `Manager`、`Extension`、`ExtensionKind`、`InstallSource`、`TemplateType`、manifest 和 update state 的 typed model。
- [x] 2.2 实现扩展名称解析和校验，覆盖 `yunxiao-<name>` 前缀、短名提取、小写字母/数字/短横线约束、`.git` 后缀剥离和 Windows `.exe` 后缀。
- [x] 2.3 实现 dataDir/stateDir 解析：优先读取 `YUNXIAO_CONFIG_DIR`、`XDG_DATA_HOME`、`XDG_STATE_HOME`，macOS/Linux 缺省写入 `~/.local/share/yunxiao` 和 `~/.local/state/yunxiao`，Windows 缺省写入 `%LocalAppData%\yunxiao`，并保持 config、credentials、data、state 分离。
- [x] 2.4 实现 manifest 读写，使用权限受限文件写入 `extension.json`，并在 manifest 损坏、版本未知或字段缺失时返回可读错误。
- [x] 2.5 为路径解析、名称解析、manifest 读写和跨平台后缀逻辑添加单元测试。

## 3. 安装、列表和移除

- [x] 3.1 实现远程 Git 安装流程：clone 到临时目录、带 `--ref` 时 checkout 指定 ref、入口校验、移动到安装目录、写入 manifest、失败清理临时目录。
- [x] 3.2 实现本地目录安装流程：校验 basename 和入口文件，Unix 固定使用 symlink，Windows 固定使用 path file，并写入 local manifest。
- [x] 3.3 实现安装前安全提示：TTY 下提示来源未验证并确认，非 TTY/CI 下要求显式确认选项。
- [x] 3.4 实现 `List`，从安装目录读取 git/local 扩展，补齐来源、commit/ref、pinned/local 和 conflict 信息。
- [x] 3.5 实现 `Remove`，删除安装记录和对应 update state；local 扩展只删除安装记录，不删除用户本地源目录。
- [x] 3.6 为 remote install、local install、install failure cleanup、list empty/list conflict 和 remove 添加 manager 单元测试，Git 操作用 fake runner 覆盖。

## 4. 扩展命令和根命令执行

- [x] 4.1 在 `internal/app` 中新增 `newExtensionCommand()`，实现 `extension/extensions/ext` 命令组和 `install/list/exec/upgrade/remove/create` 子命令骨架。
- [x] 4.2 将 extension manager 注入 `RootOptions`/`Root`/`Application`，测试中允许注入 fake manager。
- [x] 4.3 在根命令构建后注册 installed extension wrappers，核心命令和 alias 冲突时跳过 wrapper，并在 `extension list` 中显示冲突状态。
- [x] 4.4 实现 wrapper 命令：短名作为 `Use`，设置 `DisableFlagParsing`，把扩展名后的参数原样交给 `Manager.Dispatch`。
- [x] 4.5 实现 `yunxiao extension exec <name> [args]`，允许显式执行冲突扩展，并复用相同 Dispatch 逻辑。
- [x] 4.6 确保扩展执行不触发核心认证前置检查；扩展需要云效 API 时通过 `yunxiao api` 自行复用认证。
- [x] 4.7 为核心命令优先、非冲突扩展直接执行、冲突扩展 exec、未知扩展和 flag 原样转发添加命令测试。

## 5. 扩展进程执行协议

- [x] 5.1 实现 `Dispatch` 使用 manifest 解析绝对可执行路径，不通过 `PATH` 搜索扩展命令。
- [x] 5.2 将 stdin/stdout/stderr 原样连接给扩展进程，并为测试注入 exec runner 捕获 argv、env 和 IO。
- [x] 5.3 设置 `YUNXIAO_EXTENSION`、`YUNXIAO_EXTENSION_NAME`、`YUNXIAO_EXTENSION_DIR`、`YUNXIAO_ENDPOINT`、`YUNXIAO_ORGANIZATION`、`YUNXIAO_PROJECT`、`YUNXIAO_REPO`，并验证不会注入 token。
- [x] 5.4 验证扩展访问云效 API 的唯一鉴权路径是调用 `yunxiao api`，由宿主 CLI 从 keyring/credentials 文件读取 token 并注入 HTTP 请求。
- [x] 5.5 收紧 `yunxiao api` 脚本化契约：默认 stdout 只输出原始响应体，`--verbose` 输出到 stderr，`--jq` 直接作用于响应体 JSON。
- [x] 5.6 处理外部进程退出码，将 `exec.ExitError` 转为 `app.ExitCode` 可识别的退出码。
- [x] 5.7 实现 Windows `.exe` 优先和脚本 shebang 通过 `sh.exe -c 'command "$@"'` 执行；缺少 `sh.exe` 时提示安装 Git for Windows。
- [x] 5.8 为 Python/Node/shebang 扩展、`yunxiao api` 鉴权代理、API JSON-only stdout、API `--jq`、IO 透传、退出码透传、环境变量、绝对路径执行、权限错误和 Windows shebang 执行添加测试。

## 6. 升级和更新提示

- [x] 6.1 实现 `yunxiao extension upgrade [name]`：无参数升级全部 git-managed 扩展，带 name 时只升级指定扩展。
- [x] 6.2 对未 pinned git 扩展使用固定更新流程：默认 `git pull --ff-only`；`--force` 时执行 `git fetch origin HEAD` 后 reset 到 `FETCH_HEAD`。
- [x] 6.3 对 pinned 扩展默认拒绝升级，对 local 扩展提示不可升级，并在升级前后清理对应 update state。
- [x] 6.4 实现执行扩展后的非阻塞更新检查：TTY、非 CI、未设置 `YUNXIAO_NO_EXTENSION_UPDATE_NOTIFIER`、每 24 小时最多检查一次。
- [x] 6.5 使用 `git ls-remote` 比较远端 HEAD 和本地 HEAD，发现更新时只输出提示，不自动升级。
- [x] 6.6 为 upgrade all/single、pinned、local、force、state cleanup、notifier throttle 和禁用环境变量添加测试。

## 7. 扩展创建和文档

- [x] 7.1 实现 `yunxiao extension create <name>` 默认 script 模板，生成 `yunxiao-<name>` 目录和同名可执行脚本入口。
- [x] 7.2 实现 `yunxiao extension create <name> --template go`，生成 `main.go`、`go.mod`、`.gitignore` 和本地 build 说明。
- [x] 7.3 为 create 命令添加已存在目录、非法名称、script 模板和 go 模板测试。
- [x] 7.4 更新命令参考文档，记录 `extension` 命令、别名、参数、输出示例、错误示例、环境变量和退出码行为。
- [x] 7.5 更新迁移指南，说明 `gh extension install/list/exec/upgrade/remove/create` 到 `yunxiao extension` 的对应关系，以及不支持 GitHub Releases/browse/search 的原因。
- [x] 7.6 补充扩展作者指南，说明命名约定、Python/Node/shell 最小扩展示例、通过 `yunxiao api` 复用认证、安全审计建议和本地调试流程。

## 8. 验证

- [x] 8.1 运行 `go test ./...`，确认 extension manager、命令测试和现有命令测试全部通过。
- [x] 8.2 运行 `go vet ./...`。
- [x] 8.3 运行 `go build ./cmd/yunxiao`。
- [x] 8.4 使用本地 fixture 扩展手动 smoke：install local、list、direct exec、extension exec、remove。
- [x] 8.5 使用本地 bare Git fixture smoke：remote install、`--ref`、upgrade、pinned upgrade failure。
- [x] 8.6 运行 `openspec validate --strict`，确认 `add-extension-support` 的 proposal/design/specs/tasks 可归档。
