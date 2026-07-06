## 1. Go 工程和基础框架

- [x] 1.1 初始化 Go module、`cmd/yunxiao` 入口和 `internal/` 包结构。
- [x] 1.2 使用 Cobra 建立根命令、子命令注册、flags、help、version 和统一错误处理。
- [x] 1.3 引入并封装首期终端依赖：`huh`、`lipgloss`、`glamour`、`spinner`、`go-keyring`、`x/term`、`isatty`、`go-colorable`。
- [x] 1.4 预留 Bubble Tea/Bubbles 交互边界，但首期只在需要状态机的命令中使用，避免实现完整 TUI。
- [x] 1.5 将 `api-map.md` 中的 API 映射转化为 Go 服务接口清单，并标记待确认和旧版 API。
- [x] 1.6 将 `output-map.md` 中的输出契约转化为 Go result model 和渲染测试用例清单。
- [x] 1.7 实现云效 HTTP 客户端，支持 endpoint、认证注入、超时、分页、错误码和请求 ID。
- [x] 1.8 实现 `yunxiao auth login/status/logout`，支持交互输入 token、`--with-token` 从 stdin 读取、个人访问令牌保存、读取和失效处理。
- [x] 1.9 实现全局配置、仓库级配置、环境变量和命令行参数的上下文解析优先级。
- [x] 1.10 实现 `yunxiao api`，支持 GET/POST/PATCH/PUT/DELETE、请求体、请求头和 verbose 诊断脱敏。
- [x] 1.11 配置 `go test ./...`、`go vet ./...` 和跨平台构建脚本。

## 2. 输出和本地体验

- [x] 2.1 建立统一输出渲染层，支持表格、纯文本和 JSON。
- [x] 2.2 为列表和详情命令实现 `--json <fields>`。
- [x] 2.3 为 JSON 输出实现 `--jq` 和 `--template`。
- [x] 2.4 实现 `--web` 行为，在图形环境打开浏览器，在无图形环境输出 URL。
- [x] 2.5 实现 shell completion 和 alias/config 基础命令。
- [x] 2.6 实现统一 prompt/confirm/form 抽象，确保所有交互命令都有等价非交互参数或 stdin 模式。
- [x] 2.7 使用 Glamour 渲染 MR、评论和工作项 Markdown，并在无 TTY 或 `--plain` 时输出纯文本。
- [x] 2.8 为 spinner、颜色、TTY 检测和 Windows 终端颜色兼容添加可测试封装。
- [x] 2.9 为 `output-map.md` 中每个命令添加 golden output 测试，覆盖默认输出、空结果和关键错误输出。

## 3. 代码库能力

- [x] 3.1 实现 `yunxiao repo list/view/create/update/archive/unarchive/delete`。
- [x] 3.2 实现 `yunxiao repo set-default`，支持当前本地 Git 仓库绑定云效代码库。
- [x] 3.3 实现 `yunxiao branch list/view`，展示默认分支和保护状态。
- [x] 3.4 实现 `yunxiao commit list/view`，支持按分支、路径和提交范围过滤。
- [x] 3.5 实现 `yunxiao file view/tree`，支持按 ref 查看文件内容和目录树。
- [x] 3.6 实现 `yunxiao ssh-key list/add/delete`。

## 4. 合并请求 MVP

- [x] 4.1 实现 `yunxiao mr list/view/create/edit`。
- [x] 4.2 实现 `yunxiao mr diff/files`，支持当前版本和指定版本。
- [x] 4.3 实现 `yunxiao mr comment`，支持正文参数和从文件读取正文。
- [x] 4.4 实现 `yunxiao mr approve/changes-requested`。
- [x] 4.5 实现 `yunxiao mr merge/close/reopen`。
- [x] 4.6 实现 `yunxiao mr status`，聚合评审、冲突、流水线和可合并状态。

## 5. 流水线和运行实例

- [x] 5.1 实现 `yunxiao pipeline list/view`。
- [x] 5.2 实现 `yunxiao pipeline run`，支持分支和变量参数。
- [x] 5.3 实现 `yunxiao run list/view`。
- [x] 5.4 实现 `yunxiao run log`，支持任务日志和失败日志过滤。
- [x] 5.5 实现 `yunxiao run watch`，支持轮询间隔、超时和终态退出码。
- [x] 5.6 实现 `yunxiao run cancel/retry/retry-task/stop-task/skip-task`。

## 6. 项目协作和工作项

- [x] 6.1 实现 `yunxiao workitem list/view/create/edit/delete`。
- [x] 6.2 实现 `yunxiao workitem activity`，展示工作项动态。
- [x] 6.3 实现 `yunxiao project list/view`。
- [x] 6.4 实现 `yunxiao project member list`。
- [x] 6.5 实现 `yunxiao project iteration list/view`。
- [x] 6.6 实现 `yunxiao project milestone list/view` 和 `yunxiao project label list`。

## 7. 搜索能力

- [x] 7.1 实现 `yunxiao search repo`。
- [x] 7.2 实现 `yunxiao search code`。
- [x] 7.3 实现 `yunxiao search commit`。
- [x] 7.4 实现 `yunxiao search mr`。
- [x] 7.5 实现 `yunxiao search workitem`。

## 8. 验证和文档

- [x] 8.1 为 HTTP 客户端、上下文解析和输出渲染添加单元测试。
- [x] 8.2 为 repo、mr、pipeline、workitem、search 命令添加 mock API 命令测试。
- [ ] 8.3 使用真实云效测试环境运行认证、MR 查看、流水线查看和搜索 smoke 测试。
- [ ] 8.4 对 `api-map.md` 中所有旧版 API 和待确认 API 执行单独 smoke，确认不可用的命令从首期范围移除或降级。
- [x] 8.5 编写命令参考文档，明确每个命令使用的云效 API 链接、默认输出示例、JSON 字段、与 `gh` 的映射关系和不支持能力。
- [x] 8.6 编写迁移指南，说明 `gh auth/repo/pr/issue/workflow/run` 到云效 CLI 的常用替代命令和输出差异。
