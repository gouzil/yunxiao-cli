# cli-merge-request Specification

## Purpose
TBD - created by archiving change plan-yunxiao-gh-compatible-cli. Update Purpose after archive.
## Requirements
### Requirement: 合并请求基础操作
CLI SHALL 支持云效合并请求的创建、查看、列表、更新、关闭、重开和合并。

#### Scenario: 创建合并请求
- **WHEN** 用户执行 `yunxiao mr create --source <branch> --target <branch> --title <title>`
- **THEN** CLI MUST 创建云效合并请求，并输出 MR 编号、标题、状态和 Web URL

#### Scenario: 查看合并请求
- **WHEN** 用户执行 `yunxiao mr view <mr>`
- **THEN** CLI MUST 输出 MR 的标题、作者、源分支、目标分支、状态、评审状态、流水线状态和 Web URL

### Requirement: 合并请求变更查看
CLI SHALL 支持查看合并请求差异、变更文件和版本信息。

#### Scenario: 查看 diff
- **WHEN** 用户执行 `yunxiao mr diff <mr>`
- **THEN** CLI MUST 输出该合并请求当前版本的文本 diff

#### Scenario: 查看变更文件
- **WHEN** 用户执行 `yunxiao mr files <mr>`
- **THEN** CLI MUST 输出该合并请求包含的变更文件、变更类型和统计信息

### Requirement: 合并请求评论和评审
CLI SHALL 支持合并请求评论、回复、评审通过和要求修改。

#### Scenario: 发表评论
- **WHEN** 用户执行 `yunxiao mr comment <mr> --body <text>`
- **THEN** CLI MUST 在对应合并请求中创建评论，并输出评论标识

#### Scenario: 提交评审结论
- **WHEN** 用户执行 `yunxiao mr approve <mr>` 或 `yunxiao mr changes-requested <mr>`
- **THEN** CLI MUST 向云效提交对应评审结论，并输出最新评审状态

### Requirement: 合并请求状态检查
CLI SHALL 在合并请求详情中展示与合并相关的状态信息，包括评审、冲突、流水线和可合并状态。

#### Scenario: 查看状态
- **WHEN** 用户执行 `yunxiao mr status <mr>`
- **THEN** CLI MUST 输出该合并请求是否可合并，以及阻塞合并的具体原因
