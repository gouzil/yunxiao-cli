## Why

当前仓库已有 Go 单元测试、vet 脚本和本地构建脚本，但没有 GitHub Actions 持续集成来在合并前自动验证。随着 CLI 覆盖认证、输出格式、扩展执行和云效 API 封装，最小可用的 CI 应该先保证常见操作系统上的测试、静态检查和可编译性。

## What Changes

- 新增 GitHub Actions CI 工作流，在 pull request 和 `main` 分支 push 时运行。
- CI 使用 Go 版本文件读取仓库声明的 Go 版本，并启用 Go module cache。
- CI 覆盖常见系统矩阵：Ubuntu、macOS、Windows。
- 每个系统都运行 `go test ./...`、`go vet ./...` 和 `go build ./cmd/yunxiao`。
- CI 不负责正式发布，不创建 GitHub Release，也不上传发布产物。

## Capabilities

### New Capabilities

- `ci-automation`: 定义 GitHub Actions 持续集成的触发条件、系统矩阵、测试/检查步骤和非发布边界。

### Modified Capabilities

- 无。

## Impact

- 新增 `.github/workflows/ci.yml`。
- 复用现有 Go module、测试文件和 `cmd/yunxiao` 构建入口。
- 不新增运行时代码、不引入新 Go 依赖、不改变云效 CLI 命令行为。
- 与现有 `add-github-goreleaser-release-ci` 规划保持边界：本变更只做测试 CI，release CI 仍由发布自动化变更负责。
