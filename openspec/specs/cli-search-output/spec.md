# cli-search-output Specification

## Purpose
TBD - created by archiving change plan-yunxiao-gh-compatible-cli. Update Purpose after archive.
## Requirements
### Requirement: 搜索能力
CLI SHALL 支持跨云效资源的搜索命令，至少覆盖代码库、代码、提交、合并请求和工作项。

#### Scenario: 搜索代码
- **WHEN** 用户执行 `yunxiao search code <query> --repo <repo>`
- **THEN** CLI MUST 输出匹配的文件路径、代码片段、分支或提交引用

#### Scenario: 搜索合并请求
- **WHEN** 用户执行 `yunxiao search mr <query> --state open`
- **THEN** CLI MUST 输出匹配的合并请求列表

### Requirement: 结构化输出
CLI SHALL 为可列表化和可查看的命令提供结构化输出能力，包括 `--json`、`--jq` 和 `--template`。

#### Scenario: JSON 输出
- **WHEN** 用户执行支持结构化输出的命令并指定 `--json <fields>`
- **THEN** CLI MUST 只输出请求字段的合法 JSON

#### Scenario: JQ 过滤
- **WHEN** 用户同时指定 `--json` 和 `--jq <expression>`
- **THEN** CLI MUST 对 JSON 结果应用 jq 表达式并输出过滤后的结果

### Requirement: 默认终端输出契约
CLI SHALL 为每个命令提供稳定的人类可读默认输出，并与 `output-map.md` 中定义的字段和格式保持一致。

#### Scenario: 认证状态展示
- **WHEN** 用户执行 `yunxiao auth status`
- **THEN** CLI MUST 按 endpoint 分组展示登录账号、active 状态、Git 协议、脱敏 token 和 token scopes 或 `unknown`

#### Scenario: 列表命令展示
- **WHEN** 用户执行列表类命令且未指定 `--json`
- **THEN** CLI MUST 使用稳定表头输出资源列表，并在无结果时输出 `No <resource> found.`

#### Scenario: 写操作展示
- **WHEN** 用户执行创建、更新、删除、关闭、合并或触发运行类命令且操作成功
- **THEN** CLI MUST 输出一行成功摘要，并输出关键资源标识和 Web URL 或后续操作所需字段

### Requirement: 本地易用性
CLI SHALL 提供浏览器打开、shell 补全、别名和一致错误格式等本地体验能力。

#### Scenario: 浏览器打开
- **WHEN** 用户对支持 Web URL 的资源命令指定 `--web`
- **THEN** CLI MUST 在默认浏览器打开该资源页面，或在无图形环境下输出 URL

#### Scenario: shell 补全
- **WHEN** 用户执行 `yunxiao completion zsh`
- **THEN** CLI MUST 输出可安装的 zsh 补全脚本
