# ci-automation Specification

## Purpose
TBD - created by archiving change add-github-actions-ci. Update Purpose after archive.
## Requirements
### Requirement: Pull request CI validation

系统 SHALL 在 GitHub pull request 上自动运行持续集成检查，防止未通过测试、vet 或编译检查的变更进入主分支。

#### Scenario: Pull request starts CI

- **WHEN** 维护者创建或更新指向仓库的 pull request
- **THEN** GitHub Actions MUST 启动 CI workflow
- **AND** workflow MUST 运行测试、vet 和 CLI 编译检查

#### Scenario: Pull request CI fails on broken tests

- **WHEN** pull request 中的代码导致任一系统上的 `go test ./...` 失败
- **THEN** CI workflow MUST 失败
- **AND** 失败结果 MUST 出现在该 pull request 的检查状态中

### Requirement: Main branch push CI validation

系统 SHALL 在 `main` 分支 push 后自动运行持续集成检查，确认主分支仍然可测试、可 vet、可编译。

#### Scenario: Main push starts CI

- **WHEN** 维护者向 `main` 分支 push commit
- **THEN** GitHub Actions MUST 启动 CI workflow
- **AND** workflow MUST 使用与 pull request 相同的检查步骤

### Requirement: Common operating system matrix

系统 SHALL 在 Ubuntu、macOS 和 Windows 三类常见系统上运行同一组 CI 检查。

#### Scenario: CI covers common systems

- **WHEN** CI workflow 被触发
- **THEN** workflow MUST 为 `ubuntu-latest`、`macos-latest` 和 `windows-latest` 创建矩阵任务
- **AND** 每个矩阵任务 MUST 独立报告成功或失败状态

### Requirement: Go validation steps

系统 SHALL 在每个 CI 矩阵任务中验证 Go 测试、静态检查和 CLI 编译。

#### Scenario: CI runs Go validation commands

- **WHEN** 任一系统的 CI 矩阵任务运行
- **THEN** 任务 MUST 执行 `go test ./...`
- **AND** 任务 MUST 执行 `go vet ./...`
- **AND** 任务 MUST 执行 `go build ./cmd/yunxiao`

### Requirement: CI uses repository Go version

系统 SHALL 从仓库 Go 版本声明中配置 CI 的 Go 工具链，并缓存 Go 依赖以减少重复运行时间。

#### Scenario: CI sets up Go from go.mod

- **WHEN** CI workflow 准备 Go 环境
- **THEN** workflow MUST 从 `go.mod` 读取 Go 版本
- **AND** workflow MUST 启用 Go module/build cache

### Requirement: CI does not publish releases

系统 SHALL 保持测试 CI 与正式发布流程分离。

#### Scenario: CI completes without release side effects

- **WHEN** CI workflow 在 pull request 或 `main` push 上成功完成
- **THEN** workflow MUST NOT 创建 GitHub Release
- **AND** workflow MUST NOT 上传正式发布产物
- **AND** workflow MUST NOT 依赖 tag-triggered release workflow

