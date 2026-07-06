## ADDED Requirements

### Requirement: 工作项操作
CLI SHALL 使用云效工作项承接 GitHub Issue 类场景，支持创建、查看、列表、搜索、更新、删除和动态查看。

#### Scenario: 创建工作项
- **WHEN** 用户执行 `yunxiao workitem create --type <type> --title <title>`
- **THEN** CLI MUST 创建对应类型的云效工作项，并输出工作项 ID、标题、状态和 Web URL

#### Scenario: 搜索工作项
- **WHEN** 用户执行 `yunxiao workitem list --state <state> --assignee <user>`
- **THEN** CLI MUST 输出匹配条件的工作项列表

### Requirement: 项目协作操作
CLI SHALL 支持云效项目的查看、列表、成员、迭代、里程碑和标签等项目协作能力。

#### Scenario: 查看项目
- **WHEN** 用户执行 `yunxiao project view <project>`
- **THEN** CLI MUST 输出项目名称、标识、状态、成员摘要、迭代摘要和 Web URL

#### Scenario: 查看迭代
- **WHEN** 用户执行 `yunxiao project iteration list <project>`
- **THEN** CLI MUST 输出项目下迭代列表及其时间范围和状态

### Requirement: GitHub issue/project 语义边界
CLI SHALL 在帮助和文档中明确 `workitem` 与 GitHub Issue、`project` 与 GitHub Project 的差异。

#### Scenario: 查看帮助
- **WHEN** 用户执行 `yunxiao workitem --help`
- **THEN** CLI MUST 使用云效工作项术语描述命令，而不是声称其完全等同 GitHub Issue
