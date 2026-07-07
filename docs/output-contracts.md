# Yunxiao CLI 输出契约

所有面向人的输出都由同一份类型化模型渲染；`--json`、`--jq` 和 `--template` 也使用这份模型。

## 通用规则

- 列表命令使用稳定表格输出。
- 空列表命令输出 `No <resource> found.`
- 详情命令输出标题和稳定字段名。
- `--json <fields>` 接收逗号分隔的 camelCase JSON 字段，字段来自类型化结果。
- `repo list --json id,name,path,webUrl,defaultBranch` 会按仓库条目输出数组；如需响应包装和分页元数据，使用 `--json repositories,meta`。
- `--jq <expr>` 只会在 `--json` 之后执行。
- `--template <tmpl>` 使用类型化结果值执行 Go template。
- token 和凭据在展示前会被脱敏。

## `yunxiao api`

`yunxiao api <method> <path>` 是脚本和扩展的逃生口。它的 stdout 契约不同于一等资源命令：

- 默认 stdout 只包含原始 OpenAPI 响应体。
- 如果响应体是 JSON，`--jq <expr>` 会直接过滤这份响应 JSON。
- `--verbose` 把请求诊断信息写到 stderr，而不是 stdout。
- 扩展作者应通过这个命令复用宿主认证，不要直接接收 token。

## 初始 Golden 覆盖

当前重点测试覆盖：

| 范围 | 测试文件 |
| --- | --- |
| HTTP token 注入、响应元数据、API 错误格式、debug 脱敏 | `internal/api/client_test.go` |
| 配置优先级：显式参数、环境变量、仓库配置、全局配置、默认值 | `internal/config/config_test.go` |
| JSON 字段选择、jq 渲染、空表输出 | `internal/output/output_test.go` |
| headless 环境下 `--web` fallback 行为 | `internal/browser/browser_test.go` |
| API 原始 JSON stdout 和扩展命令分发 | `internal/app/command_test.go` |

后续可以按命令继续补充 golden 测试，表格定义在 `internal/output/views.go`。
