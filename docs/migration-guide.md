# 从 `gh` 迁移到 `yunxiao`

`yunxiao` 借鉴 `gh` 的终端使用习惯，但命令语义使用云效产品名。

## 常见替换

| 原 `gh` 命令 | 使用 `yunxiao` |
| --- | --- |
| `gh auth login` | `yunxiao auth login` |
| `gh auth status` | `yunxiao auth status` |
| `gh repo list` | `yunxiao repo list` |
| `gh repo view` | `yunxiao repo view` |
| `gh pr list` | `yunxiao mr list` |
| `gh pr view` | `yunxiao mr view` |
| `gh pr create` | `yunxiao mr create --source <branch> --target <branch> --title <title>` |
| `gh pr diff` | `yunxiao mr diff <mr>` |
| `gh pr merge` | `yunxiao mr merge <mr>` |
| `gh workflow list` | `yunxiao pipeline list` |
| `gh workflow run` | `yunxiao pipeline run <pipeline-id> --branch <branch>` |
| `gh run view` | `yunxiao run view <run-id>` |
| `gh run watch` | `yunxiao run watch <run-id>` |
| `gh issue list` | `yunxiao workitem list` |
| `gh issue view` | `yunxiao workitem view <workitem>` |
| `gh project list` | `yunxiao project list` |
| `gh api` | `yunxiao api` |
| `gh extension install <source>` | `yunxiao extension install <source> --yes` |
| `gh extension list` | `yunxiao extension list` |
| `gh extension exec <name>` | `yunxiao extension exec <name>` |
| `gh extension upgrade [name]` | `yunxiao extension upgrade [name]` |
| `gh extension remove <name>` | `yunxiao extension remove <name>` |
| `gh extension create <name>` | `yunxiao extension create <name>` |

## 输出差异

- `yunxiao` 使用 camelCase JSON 字段。
- `yunxiao --json <fields>` 必须显式指定字段。
- 对一等资源命令，`--jq` 在 `--json` 之后执行。
- 对 `yunxiao api`，stdout 默认是原始响应体，`--jq` 直接作用于响应 JSON。
- `workitem` 不是 GitHub Issue 的一对一复制；它遵循云效需求、缺陷、任务等工作项类型。
- `pipeline` 表示云效流水线定义；`run` 表示一次流水线运行实例。
- 扩展使用 `yunxiao-<name>` 可执行文件约定。扩展从完整 Git URL 或本地目录安装；当前不支持 GitHub Releases、`gh extension browse` 或基于 topic 的扩展搜索。
- 扩展不会直接收到 token。扩展内应调用 `yunxiao api` 复用宿主认证。
