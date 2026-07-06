## Why

`yunxiao-cli` 已经具备 `gh` 风格的核心命令、别名、补全和通用 API 逃生口，但还缺少用户自定义命令的扩展机制。引入扩展能力可以让团队在不等待核心 CLI 发版的情况下，把企业内部流程、定制查询、批量操作和实验性云效 API 包装成 `yunxiao <ext>` 命令。

当前首期规划曾将 GitHub Extension/Skill 生态列为 Non-Goal；本变更将其作为独立增量能力重新建模，参考 `gh extension` 的成熟设计，但只吸收适合云效 CLI 的执行、安装、升级和安全边界。

## What Changes

- 新增 `yunxiao extension` 管理命令，提供扩展的安装、列表、执行、升级、移除和本地开发入口。
- 支持通过根命令直接执行已安装扩展，例如 `yunxiao foo ...` 转发到扩展可执行文件。
- 支持 `yunxiao extension exec <name> ...`，在扩展名与核心命令冲突时仍可显式运行扩展。
- 支持从 Git 仓库和本地目录安装扩展，扩展仓库和可执行文件命名采用 `yunxiao-<name>` 约定。
- 扩展运行协议保持语言无关；Python、Node.js、Ruby、Rust、Go、shell 等实现只要提供同名可执行入口即可运行。
- 设计扩展安装目录、manifest、版本检测、升级状态和跨平台执行规则。
- 约束扩展不能覆盖核心命令，别名和核心命令优先级高于普通扩展。
- 为扩展执行提供明确的环境变量、stdin/stdout/stderr 转发、退出码透传和错误展示契约。
- 在任务文档中列出 GitHub CLI 扩展设计、源码和官方文档参考链接，便于实现阶段对照。

## Capabilities

### New Capabilities

- `cli-extension`: 云效 CLI 扩展的安装、发现、执行、升级、删除、本地开发和安全提示能力。

### Modified Capabilities

无。

## Impact

- 影响根命令构建、Cobra 命令注册顺序、未知命令处理、alias 优先级和 shell completion。
- 新增扩展管理包，负责安装目录、manifest、Git clone/pull、local symlink 或 path file、跨平台可执行文件查找和执行。
- 需要补充配置目录和状态目录约定：Unix 使用 `$XDG_DATA_HOME/yunxiao/extensions`、`$XDG_STATE_HOME/yunxiao/extensions`，macOS/Linux 缺省使用 `~/.local/share/yunxiao/extensions`、`~/.local/state/yunxiao/extensions`，Windows 使用 `%LocalAppData%\yunxiao\extensions`、`%LocalAppData%\yunxiao\state\extensions`。
- 需要使用本地 `git` 和 `os/exec`，并在缺少 Git、缺少可执行文件、权限不足、Windows shebang 执行等场景给出清晰错误。
- 需要新增命令测试、manager 单元测试、fixture 扩展、跨平台路径测试和文档示例。
- 安全上必须明确：扩展是用户信任并执行的本地程序，安装和升级前 CLI 只能提供来源提示与确认，不能保证扩展被云效官方验证。
