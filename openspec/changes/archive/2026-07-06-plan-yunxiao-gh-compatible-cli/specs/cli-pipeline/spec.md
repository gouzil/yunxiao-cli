## ADDED Requirements

### Requirement: 流水线定义操作
CLI SHALL 支持云效流水线定义的列表、查看、创建、更新、删除和触发运行。

#### Scenario: 列出流水线
- **WHEN** 用户执行 `yunxiao pipeline list`
- **THEN** CLI MUST 输出当前组织或项目下用户有权限访问的流水线列表

#### Scenario: 触发流水线
- **WHEN** 用户执行 `yunxiao pipeline run <pipeline-id> --branch <branch>`
- **THEN** CLI MUST 触发流水线运行，并输出运行实例 ID 和 Web URL

### Requirement: 运行实例操作
CLI SHALL 支持云效流水线运行实例的列表、查看、取消、重试和状态跟踪。

#### Scenario: 查看运行实例
- **WHEN** 用户执行 `yunxiao run view <run-id>`
- **THEN** CLI MUST 输出运行实例的状态、触发人、触发来源、开始时间、结束时间、阶段和任务摘要

#### Scenario: 等待运行完成
- **WHEN** 用户执行 `yunxiao run watch <run-id>`
- **THEN** CLI MUST 周期性刷新运行状态，直到运行进入终态或达到用户指定超时

### Requirement: 流水线日志和任务控制
CLI SHALL 支持查看运行日志，并对可控任务执行停止、跳过或重试。

#### Scenario: 查看日志
- **WHEN** 用户执行 `yunxiao run log <run-id> --job <job-id>`
- **THEN** CLI MUST 输出指定任务日志，并支持只查看失败任务日志

#### Scenario: 控制任务
- **WHEN** 用户执行 `yunxiao run retry-task <run-id> --job <job-id>`
- **THEN** CLI MUST 调用云效任务控制接口并输出新的任务状态
