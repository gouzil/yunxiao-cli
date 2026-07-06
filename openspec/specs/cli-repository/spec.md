# cli-repository Specification

## Purpose
TBD - created by archiving change plan-yunxiao-gh-compatible-cli. Update Purpose after archive.
## Requirements
### Requirement: 代码库基础操作
CLI SHALL 支持云效代码库的列表、查看、创建、更新、归档、取消归档和删除操作。

#### Scenario: 列出代码库
- **WHEN** 用户执行 `yunxiao repo list`
- **THEN** CLI MUST 输出当前组织或企业下用户有权限访问的代码库列表

#### Scenario: 查看代码库
- **WHEN** 用户执行 `yunxiao repo view <repo>`
- **THEN** CLI MUST 输出该代码库的名称、路径、默认分支、可见性、归档状态和 Web URL

### Requirement: 本地仓库绑定
CLI SHALL 支持将当前本地 Git 仓库绑定到云效代码库，供后续命令默认使用。

#### Scenario: 设置默认代码库
- **WHEN** 用户执行 `yunxiao repo set-default <repo>`
- **THEN** CLI MUST 在仓库级配置中保存默认云效代码库标识

#### Scenario: 默认代码库缺失
- **WHEN** 用户执行需要代码库上下文的命令且 CLI 无法推导默认代码库
- **THEN** CLI MUST 输出缺失上下文说明和可执行的修复建议

### Requirement: 分支、提交和文件操作
CLI SHALL 支持查看云效代码库中的分支、提交、文件内容和文件树。

#### Scenario: 查看分支
- **WHEN** 用户执行 `yunxiao branch list`
- **THEN** CLI MUST 输出当前代码库的分支列表，并标识默认分支

#### Scenario: 查看文件内容
- **WHEN** 用户执行 `yunxiao file view <path> --ref <ref>`
- **THEN** CLI MUST 输出指定 ref 下文件内容或结构化文件元数据

### Requirement: SSH Key 管理
CLI SHALL 支持管理当前用户的云效 SSH Key。

#### Scenario: 添加 SSH Key
- **WHEN** 用户执行 `yunxiao ssh-key add <public-key-file>`
- **THEN** CLI MUST 将公钥添加到云效账号，并输出新 Key 的标识和标题

#### Scenario: 删除 SSH Key
- **WHEN** 用户执行 `yunxiao ssh-key delete <key-id>`
- **THEN** CLI MUST 删除指定 SSH Key，交互模式下需要用户确认
