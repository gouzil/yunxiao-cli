## 1. GitHub Actions CI 工作流

- [x] 1.1 新增 `.github/workflows/ci.yml`。
- [x] 1.2 配置 workflow 监听 `pull_request` 和 `main` 分支 `push`。
- [x] 1.3 配置 `ubuntu-latest`、`macos-latest`、`windows-latest` 三系统 matrix。
- [x] 1.4 在每个 matrix job 中使用 checkout 和 setup-go，从 `go.mod` 读取 Go 版本并启用 Go cache。
- [x] 1.5 在每个 matrix job 中依次运行 `go test ./...`、`go vet ./...` 和 `go build ./cmd/yunxiao`。
- [x] 1.6 确认 CI workflow 不包含 release 创建、artifact 上传或 tag 发布步骤。

## 2. 本地验证

- [x] 2.1 运行 `go test ./...`。
- [x] 2.2 运行 `go vet ./...`。
- [x] 2.3 运行 `go build ./cmd/yunxiao`。
- [x] 2.4 运行 `openspec validate --strict`，确认变更 artifact 可进入实现阶段。
