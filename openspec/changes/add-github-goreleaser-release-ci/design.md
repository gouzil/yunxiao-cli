## Context

当前 CLI 入口在 `cmd/yunxiao/main.go` 定义 `version`、`commit` 和 `date` 变量，根命令通过 `BuildInfo` 展示 `yunxiao --version`。`scripts/build.sh` 已经用 `go build -ldflags` 为 Darwin、Linux、Windows 的 amd64/arm64 产出二进制，但它只负责本地构建，不会创建版本、归档、checksum 或上传 Release。

本变更把正式发布职责交给 GitHub Actions + GoReleaser。Git tag 是版本唯一来源；推送 `v*` tag 后，CI 自动构建、归档、生成 checksum，并上传到 GitHub Release。

## Goals / Non-Goals

**Goals:**

- 在仓库中增加 GoReleaser 配置，覆盖当前 `scripts/build.sh` 的平台矩阵。
- 增加 GitHub Actions release workflow，在 `v*` tag push 时自动运行 GoReleaser。
- 让 GitHub Release 持有所有跨平台归档产物和 checksum。
- 保持 CLI 内部版本展示链路不变，只由 GoReleaser 注入构建信息。
- 在 README 中说明本地构建、snapshot 检查和正式 tag 发布流程。

**Non-Goals:**

- 不引入 Homebrew、Scoop、Docker 镜像、签名、公证或 SBOM。
- 不实现云效制品仓库、阿里云 OSS 或公司内部制品上传。
- 不新增运行时版本配置文件，不让应用代码读取 GitHub API。
- 不改变任何云效 OpenAPI 命令行为。

## Decisions

### 1. tag 是版本唯一来源

正式版本发布 MUST 通过 `v*` git tag 触发，例如 `v0.1.0`。GoReleaser 从 tag 解析版本，并通过 ldflags 注入 `main.version`、`main.commit` 和 `main.date`。

理由：当前 CLI 已有构建变量，不需要再维护 `VERSION` 文件或额外版本命令。tag 也是 GitHub Release 的自然索引。

备选方案：

- 使用 `VERSION` 文件：会制造第二个版本源，容易和 tag 漂移。
- 继续依赖 `VERSION` 环境变量：适合本地脚本，不适合可审计的自动发布。

### 2. 使用 GoReleaser 管理正式发布产物

新增 `.goreleaser.yaml`，配置一个 `yunxiao` build，入口为 `./cmd/yunxiao`，目标平台为：

```text
darwin/amd64
darwin/arm64
linux/amd64
linux/arm64
windows/amd64
```

GoReleaser MUST 生成归档产物和 `checksums.txt`，并使用 GitHub Release publisher 上传。

理由：这复用当前脚本的真实平台矩阵，同时补齐归档、checksum、changelog 和 release 上传。手写 shell 继续扩展这些能力没有价值。

备选方案：

- 只用 GitHub Actions 手写 `go build` 和 `gh release upload`：配置更长，checksum 和归档规则要自己维护。
- 保留脚本作为正式发布入口：仍然缺少 release 元数据和 GitHub Release 生命周期管理。

### 3. GitHub Actions 只在 tag 上发布

新增 `.github/workflows/release.yml`。workflow MUST 监听 `push.tags: ["v*"]`，授予 `contents: write`，checkout MUST 使用完整历史，GoReleaser step 使用官方 action 并执行 `release --clean`。

建议 action 组合：

```text
actions/checkout@v6
actions/setup-go@v6
goreleaser/goreleaser-action@v7
```

GoReleaser 版本线固定为 `~> v2`，避免 CI 随 `latest` 非预期升级。

理由：tag-only 发布避免普通分支 push 意外创建 Release；完整 git 历史让 changelog 和 tag 解析可靠。

备选方案：

- pull request 和分支 push 都跑 `release`：会误触发发布语义。
- 使用 `latest` GoReleaser：省配置，但发布工具链漂移不可控。

### 4. 本地脚本降级为开发便利入口

`scripts/build.sh` 不再是正式发布路径。实现时可以选择保留当前脚本用于快速本地 build，或把它改成 GoReleaser snapshot 包装器。无论哪种，README MUST 明确正式发布只走 tag-triggered GitHub Actions。

理由：本地调试和正式发布的边界要清楚。脚本保留为开发便利即可，不应该承担版本事实来源。

备选方案：

- 删除脚本：更干净，但会破坏 README 里的现有本地构建习惯。
- 让脚本创建 tag 并推送：过度封装，容易隐藏发布前检查。

## Risks / Trade-offs

- GitHub Actions 权限不足会导致 Release 创建失败 → workflow 显式声明 `permissions.contents: write`，使用默认 `GITHUB_TOKEN`。
- checkout 历史不完整会影响 changelog 和 tag 解析 → `fetch-depth: 0`。
- 本地无 `.git` 工作区无法完整模拟 release → 本地只要求 `goreleaser check`；正式验证在 GitHub Actions 完整 clone 中完成。
- GoReleaser action 或 GitHub 官方 action 大版本后续变化 → 固定 major 版本，升级作为独立维护项处理。
