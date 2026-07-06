## ADDED Requirements

### Requirement: 扩展管理命令
CLI SHALL provide an `extension` command group for managing installed extensions, with `extensions` and `ext` aliases.

#### Scenario: 查看扩展管理帮助
- **WHEN** 用户执行 `yunxiao extension --help`
- **THEN** CLI MUST 显示扩展管理命令，并包含 `install`、`list`、`exec`、`upgrade`、`remove` 和 `create` 子命令

#### Scenario: 使用短别名访问扩展管理
- **WHEN** 用户执行 `yunxiao ext list`
- **THEN** CLI MUST 执行与 `yunxiao extension list` 相同的逻辑

### Requirement: 扩展命名约定
CLI SHALL require extension repositories, local directories, and executable files to use the `yunxiao-<name>` naming convention.

#### Scenario: 安装合法命名的扩展
- **WHEN** 用户安装 basename 为 `yunxiao-team-report` 的远程 Git 仓库或本地目录
- **THEN** CLI MUST 将扩展短名解析为 `team-report`

#### Scenario: 拒绝非法命名的扩展
- **WHEN** 用户安装 basename 不以 `yunxiao-` 开头的远程 Git 仓库或本地目录
- **THEN** CLI MUST 拒绝安装并说明扩展名称必须使用 `yunxiao-<name>` 约定

#### Scenario: 校验扩展可执行文件
- **WHEN** 安装来源中不存在名为 `yunxiao-<name>` 的可执行文件
- **THEN** CLI MUST 拒绝安装并提示需要在扩展根目录提供同名可执行文件

### Requirement: 远程 Git 扩展安装
CLI SHALL install remote extensions by cloning full Git URLs into the local extension data directory.

#### Scenario: 安装远程 Git 扩展
- **WHEN** 用户执行 `yunxiao extension install <git-url>`
- **THEN** CLI MUST clone 该仓库、校验扩展入口、写入 manifest，并输出安装成功摘要

#### Scenario: 按 ref 安装远程 Git 扩展
- **WHEN** 用户执行 `yunxiao extension install <git-url> --ref v1.2.0`
- **THEN** CLI MUST checkout 指定 tag、branch 或 commit，并在 manifest 中记录 pinned 来源

#### Scenario: 远程安装失败清理临时目录
- **WHEN** clone、checkout 或入口校验失败
- **THEN** CLI MUST 清理临时安装目录，并且不得留下半安装扩展

### Requirement: 本地扩展安装
CLI SHALL install local extensions from existing directories without copying user source files into managed storage.

#### Scenario: 安装本地扩展
- **WHEN** 用户执行 `yunxiao extension install ./yunxiao-local-tool`
- **THEN** CLI MUST 在 Unix 平台记录指向该本地目录的 symlink，在 Windows 平台记录指向该本地目录的 path file，并将扩展标记为 local

#### Scenario: 本地扩展不可升级
- **WHEN** 用户对 local 扩展执行 `yunxiao extension upgrade <name>`
- **THEN** CLI MUST 拒绝升级并提示该扩展由本地目录管理

### Requirement: 扩展列表
CLI SHALL list installed extensions with their name, kind, source, version or commit, pinned state, local state, and conflict state.

#### Scenario: 列出已安装扩展
- **WHEN** 用户执行 `yunxiao extension list`
- **THEN** CLI MUST 输出已安装扩展的短名、类型、来源和当前版本或 commit

#### Scenario: 无已安装扩展
- **WHEN** 用户执行 `yunxiao extension list` 且没有已安装扩展
- **THEN** CLI MUST 输出稳定的空结果提示

#### Scenario: 显示命令冲突
- **WHEN** 已安装扩展短名与核心命令或 alias 冲突
- **THEN** CLI MUST 在列表中标记冲突，并提示可使用 `yunxiao extension exec <name>` 显式运行

### Requirement: 扩展直接执行
CLI SHALL allow installed non-conflicting extensions to be executed as root-level commands.

#### Scenario: 直接执行扩展
- **WHEN** 用户执行 `yunxiao team-report --since 7d`
- **THEN** CLI MUST 运行已安装扩展 `yunxiao-team-report` 的可执行文件，并将 `--since 7d` 原样转发

#### Scenario: 核心命令优先于扩展
- **WHEN** 已安装扩展短名与核心命令同名
- **THEN** CLI MUST 保持核心命令行为不变，并且不得用扩展覆盖核心命令

#### Scenario: 未安装扩展
- **WHEN** 用户执行不存在的扩展短名
- **THEN** CLI MUST 输出未知命令错误，并且不得尝试通过 PATH 执行同名系统命令

### Requirement: 显式执行冲突扩展
CLI SHALL provide `yunxiao extension exec <name> [args]` for running installed extensions by short name even when root-level dispatch is unavailable.

#### Scenario: 执行冲突扩展
- **WHEN** 已安装扩展短名与核心命令冲突，用户执行 `yunxiao extension exec <name> --flag`
- **THEN** CLI MUST 运行该扩展并将 `--flag` 原样转发

#### Scenario: exec 目标不存在
- **WHEN** 用户执行 `yunxiao extension exec missing`
- **THEN** CLI MUST 输出未找到该扩展的错误

### Requirement: 扩展执行协议
CLI SHALL execute extensions by absolute path and preserve process IO and exit behavior.

#### Scenario: 运行非 Go 扩展
- **WHEN** 用户安装的扩展入口是带 shebang 的 Python、Node.js、Ruby 或 shell 可执行文件
- **THEN** CLI MUST 按同一个绝对路径执行协议运行该扩展，不得要求扩展使用 Go 实现

#### Scenario: 标准输入输出透传
- **WHEN** 用户通过管道向扩展传入标准输入
- **THEN** CLI MUST 将 stdin 传给扩展进程，并将扩展 stdout 和 stderr 原样连接到当前进程

#### Scenario: 退出码透传
- **WHEN** 扩展进程以非零退出码结束
- **THEN** CLI MUST 以相同退出码结束

#### Scenario: 执行环境变量
- **WHEN** CLI 启动扩展进程
- **THEN** CLI MUST 设置 `YUNXIAO_EXTENSION`、`YUNXIAO_EXTENSION_NAME` 和可解析的云效上下文环境变量，并且 MUST NOT 通过环境变量注入认证 token

#### Scenario: 扩展调用云效 API
- **WHEN** 扩展需要访问云效开放 API
- **THEN** 扩展 MUST 通过执行 `yunxiao api` 复用宿主 CLI 鉴权，并且 CLI MUST 从本机 credential store 读取 token 后注入 HTTP 请求，不得把 token 暴露给扩展进程

#### Scenario: API 代理返回 JSON 响应体
- **WHEN** 扩展执行 `yunxiao api GET <path>` 且云效 OpenAPI 返回 JSON
- **THEN** CLI MUST 只在 stdout 输出该 JSON 响应体，不得输出表格、标题、状态摘要或其他非 JSON 内容

#### Scenario: API 代理 JSON 过滤
- **WHEN** 扩展执行 `yunxiao api GET <path> --jq <expression>` 且云效 OpenAPI 返回 JSON
- **THEN** CLI MUST 对响应体 JSON 应用 jq 表达式，并在 stdout 输出过滤结果

### Requirement: 扩展升级
CLI SHALL support upgrading git-managed extensions while preserving safe behavior for pinned and local extensions.

#### Scenario: 升级所有可升级扩展
- **WHEN** 用户执行 `yunxiao extension upgrade`
- **THEN** CLI MUST 尝试升级所有 git-managed 且未 pinned 的扩展，并为每个扩展输出结果

#### Scenario: 升级单个扩展
- **WHEN** 用户执行 `yunxiao extension upgrade team-report`
- **THEN** CLI MUST 只升级短名为 `team-report` 的扩展

#### Scenario: pinned 扩展默认不可升级
- **WHEN** 用户升级 pinned 扩展且未指定 force 选项
- **THEN** CLI MUST 拒绝升级并说明该扩展被固定到指定 ref

### Requirement: 扩展移除
CLI SHALL remove installed extensions and their update state.

#### Scenario: 移除已安装扩展
- **WHEN** 用户执行 `yunxiao extension remove team-report`
- **THEN** CLI MUST 删除该扩展的安装记录和对应 update state，并输出成功摘要

#### Scenario: 移除不存在扩展
- **WHEN** 用户执行 `yunxiao extension remove missing`
- **THEN** CLI MUST 输出未找到该扩展的错误

### Requirement: 扩展创建
CLI SHALL provide lightweight scaffolding for creating new extensions.

#### Scenario: 创建脚本扩展
- **WHEN** 用户执行 `yunxiao extension create team-report`
- **THEN** CLI MUST 创建 `yunxiao-team-report` 目录和同名可执行脚本入口

#### Scenario: 创建 Go 扩展
- **WHEN** 用户执行 `yunxiao extension create team-report --template go`
- **THEN** CLI MUST 创建 Go 项目入口、`.gitignore` 和本地构建说明

### Requirement: 扩展安全提示
CLI SHALL make the trust boundary of third-party extensions visible during install and upgrade operations.

#### Scenario: 交互式安装提示
- **WHEN** 用户在 TTY 中安装远程扩展
- **THEN** CLI MUST 提示扩展来源未被云效 CLI 自动验证，并要求用户确认后继续

#### Scenario: 非交互安装
- **WHEN** 用户在非 TTY/CI 环境安装扩展
- **THEN** CLI MUST 要求 `--yes` 显式确认；缺少 `--yes` 时 MUST 失败并提示如何以非交互方式确认

### Requirement: 扩展更新提示
CLI SHALL check for extension updates opportunistically without slowing normal extension execution.

#### Scenario: TTY 中提示可用更新
- **WHEN** 用户运行 git-managed 扩展且该扩展距离上次检查超过 24 小时
- **THEN** CLI MUST 启动非阻塞更新检查，并在命令结束前检查完成且存在新 HEAD 时提示可用更新，同时 MUST NOT 阻塞扩展命令完成

#### Scenario: 禁用更新提示
- **WHEN** 环境变量 `YUNXIAO_NO_EXTENSION_UPDATE_NOTIFIER` 非空
- **THEN** CLI MUST NOT 自动检查扩展更新
