## 1. GoReleaser 配置

- [x] 1.1 新增 `.goreleaser.yaml`，配置 `yunxiao` build，入口为 `./cmd/yunxiao`，二进制名为 `yunxiao`。
- [x] 1.2 在 GoReleaser build 中复用现有 ldflags，注入 `main.version`、`main.commit` 和 `main.date`。
- [x] 1.3 配置 Darwin amd64、Darwin arm64、Linux amd64、Linux arm64 和 Windows amd64 构建目标。
- [x] 1.4 配置 release 归档命名和 checksum，确保 GitHub Release 能上传各平台归档和 `checksums.txt`。

## 2. GitHub Actions 发布流程

- [x] 2.1 新增 `.github/workflows/release.yml`，仅监听 `push` 到 `v*` tag。
- [x] 2.2 在 release workflow 中声明 `contents: write` 权限，使用完整 checkout 历史并设置 Go 环境。
- [x] 2.3 在 release workflow 中使用官方 GoReleaser action 执行 `release --clean`，并通过 `GITHUB_TOKEN` 发布到 GitHub Release。
- [x] 2.4 确认普通分支 push 不会触发正式 GitHub Release 发布。

## 3. 文档和本地入口

- [x] 3.1 更新 README 的构建和发布说明，明确正式发布由 `v*` tag 自动触发 GitHub Actions。
- [x] 3.2 明确 GitHub Release 是正式产物下载位置，checksum 用于产物校验。
- [x] 3.3 保留或调整 `scripts/build.sh` 为本地开发入口，并在文档中区分它和正式发布流程。

## 4. 验证

- [x] 4.1 运行 `goreleaser check` 验证 GoReleaser 配置。
- [x] 4.2 在具备完整 `.git` 历史的 checkout 中运行 GoReleaser snapshot 验证，确认归档产物和 checksum 可生成。
- [x] 4.3 运行仓库现有测试或构建检查，确认发布配置没有破坏 CLI 编译。
- [x] 4.4 运行 `openspec validate --strict`，确认本变更 artifact 可进入实现阶段。
