## Context

`yunxiao-cli` 是 Go CLI 项目，当前 `go.mod` 声明 Go `1.25.8`，仓库已有 `scripts/test.sh`、`scripts/vet.sh` 和 `scripts/build.sh`。现有测试覆盖分布在 `internal/**`，本地最小验证入口是 `go test ./...`、`go vet ./...` 和 `go build ./cmd/yunxiao`。

仓库当前没有 `.github` workflow。本变更只补齐持续集成测试，不接管版本发布；基于 tag 的 GoReleaser 发布由独立的 `add-github-goreleaser-release-ci` 变更负责。

## Goals / Non-Goals

**Goals:**

- 增加 `.github/workflows/ci.yml`，自动验证 pull request 和 `main` 分支 push。
- 使用 GitHub Actions matrix 覆盖 Ubuntu、macOS、Windows 三类常见系统。
- 每个系统都执行 Go 测试、vet 和 CLI 编译检查。
- 让 workflow 从 `go.mod` 读取 Go 版本，并启用 Go module cache。
- 保持实现简单，不新增仓库依赖和应用代码。

**Non-Goals:**

- 不创建 GitHub Release，不上传正式发布产物。
- 不引入 GoReleaser、Homebrew、Scoop、SBOM、签名或制品仓库逻辑。
- 不新增 golangci-lint 等额外 lint 依赖。
- 不改变 `scripts/*.sh` 的本地使用方式。
- 不改变任何云效 OpenAPI 命令、认证或输出行为。

## Decisions

### 1. 使用单个 CI workflow

新增 `.github/workflows/ci.yml`，监听：

- `pull_request`
- `push` 到 `main`

理由：一个 workflow 足够承载测试、vet 和构建检查。拆成多个 workflow 只会增加重复配置，当前仓库还没有复杂到需要分流。

备选方案：

- 分离 test、vet、build workflow：便于单独重跑，但重复 checkout/setup-go/matrix 配置。
- 监听所有分支 push：反馈更多，但个人分支 push 会浪费 Actions 时间；PR 已覆盖合并前校验。

### 2. 用 OS matrix 覆盖常见系统

CI matrix 使用：

```yaml
os: [ubuntu-latest, macos-latest, windows-latest]
```

每个 job 在对应 runner 上运行同一组 Go 命令。

理由：CLI 面向跨平台使用，Windows/macOS/Linux 都需要至少保证测试和主程序可编译。先覆盖 OS 维度，比扩大 Go 版本矩阵更有价值。

备选方案：

- 只跑 Ubuntu：最省资源，但不能发现 Windows/macOS 下的路径、终端和 keyring 差异。
- 增加 Go 多版本矩阵：当前 `go.mod` 已声明单一 Go 版本，先避免把 CI 时间翻倍。

### 3. CI 直接运行 Go 命令

workflow 直接执行：

```bash
go test ./...
go vet ./...
go build ./cmd/yunxiao
```

理由：Windows runner 上直接跑 Go 命令最少依赖 shell 细节。现有 `scripts/test.sh` 和 `scripts/vet.sh` 继续作为本地便利入口，不必为了 CI 额外改脚本。

备选方案：

- 在 CI 中调用 `scripts/test.sh` 和 `scripts/vet.sh`：复用入口，但 Windows shell 行为更容易变成无关问题。
- 新增 `make ci`：当前仓库没有 Makefile，为三条命令新增一层包装不划算。

### 4. 使用 `go-version-file` 和缓存

workflow 使用 `actions/setup-go` 的 `go-version-file: go.mod`，并开启 Go cache。

理由：Go 版本事实来源已经在 `go.mod`，CI 不应该再复制一个版本号。缓存 module 和 build cache 能减少重复运行时间，不影响行为。

备选方案：

- 在 workflow 写死 `1.25.8`：简单但会和 `go.mod` 漂移。
- 不开缓存：配置更短，但每次 CI 都重新下载依赖。

## Risks / Trade-offs

- GitHub runner 的 `*-latest` 镜像会随 GitHub 更新而变化 -> 这是 CI 预期行为；如果后续需要可复现构建，再改成固定 runner 版本。
- `go-version-file: go.mod` 依赖 setup-go 对当前 Go 版本格式的支持 -> 实现时先本地检查 workflow 语法，并在首次 PR CI 中验证。
- Windows 上的 keyring 相关测试可能暴露平台差异 -> 这正是三系统 CI 要捕获的问题；如果某个测试需要真实系统凭证，再把测试改成 mock 或跳过不适合 CI 的路径。
- CI 增加等待时间 -> 当前只有一个 Go 版本和三个 OS，先保持最小矩阵。
