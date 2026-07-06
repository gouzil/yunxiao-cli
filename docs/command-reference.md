# Yunxiao CLI 命令参考

本参考对应 `openspec/changes/plan-yunxiao-gh-compatible-cli/api-map.md` 和 `output-map.md`。

## `gh` 命令映射

| GitHub CLI | Yunxiao CLI |
| --- | --- |
| `gh auth` | `yunxiao auth` |
| `gh repo` | `yunxiao repo`, `yunxiao branch`, `yunxiao commit`, `yunxiao file`, `yunxiao ssh-key` |
| `gh pr` | `yunxiao mr` |
| `gh workflow` | `yunxiao pipeline` |
| `gh run` | `yunxiao run` |
| `gh issue` | `yunxiao workitem` |
| `gh project` | `yunxiao project` |
| `gh search` | `yunxiao search` |
| `gh api` | `yunxiao api` |
| `gh extension` | `yunxiao extension` |

## 不支持的 GitHub 专属能力

第一版不会把 GitHub 专属产品能力映射成 Yunxiao 命令：

- Codespaces
- Gist
- Copilot
- GitHub Attestation
- GitHub Release 语义

如果后续需要云效制品、包管理或部署能力，应按云效自己的概念新增命令组。

## 命令和 API 说明

实现层为第一阶段的每个命令组定义了类型化 service 方法。API 路径以 OpenSpec API map 为准，待确认或遗留映射会保留在 service result 中，不会静默隐藏。

未提升为一等命令的 API 可以使用 `yunxiao api <method> <path>` 调用。默认情况下，它只把原始 OpenAPI 响应体写到 stdout，方便脚本和扩展直接消费 JSON。`--verbose` 会把请求诊断信息写到 stderr，`--jq <expr>` 会直接过滤响应体 JSON。

## 扩展命令

`yunxiao extension` 管理本地 CLI 扩展。命令别名是 `yunxiao extensions` 和 `yunxiao ext`。

```text
yunxiao extension install <git-url-or-local-path> --yes
yunxiao extension list
yunxiao extension exec <name> [args]
yunxiao extension upgrade [name]
yunxiao extension remove <name>
yunxiao extension create <name> [--template script|go]
```

扩展仓库、目录和可执行入口都必须使用 `yunxiao-<name>` 命名。用户运行时使用短名称：

```text
yunxiao extension install ./yunxiao-team-report --yes
yunxiao team-report --since 7d
```

扩展与语言无关。Python、Node.js、Ruby、shell、Go、Rust 或其他实现都可以，只要提供当前平台可执行的入口文件。CLI 在安装扩展时不会安装运行时、执行包管理器或编译扩展源码。

扩展不会拿到云效 PAT。扩展只会收到非敏感上下文环境变量，并通过 `yunxiao api` 复用宿主 CLI 的凭据存储。
