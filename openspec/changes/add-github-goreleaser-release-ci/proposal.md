## Why

当前仓库只有 `scripts/build.sh` 手动交叉编译二进制，缺少基于版本 tag 的自动发布流程。CLI 已经支持通过 `ldflags` 注入 `version`、`commit` 和 `date`，适合改由 GoReleaser 在 GitHub Actions 中统一构建、打包、校验并上传到 GitHub Release。

## What Changes

- 新增 GoReleaser 发布配置，复用现有 `main.version`、`main.commit`、`main.date` 注入点。
- 新增 GitHub Actions 工作流，在推送 `v*` tag 时自动运行 GoReleaser 并创建 GitHub Release。
- 生成跨平台归档产物和 checksum，覆盖当前脚本支持的 Darwin、Linux、Windows 与 amd64/arm64 组合。
- 更新 README 的构建和发布说明，明确 tag 是版本发布的唯一来源。
- 保留本地构建入口的轻量用法，但不再让手写脚本承担正式发布职责。

## Capabilities

### New Capabilities

- `release-automation`: 定义基于 GitHub Actions 和 GoReleaser 的自动版本发布、产物上传和校验要求。

### Modified Capabilities

- 无。

## Impact

- 新增 `.goreleaser.yaml` 或等价 GoReleaser 配置文件。
- 新增 `.github/workflows/release.yml`，需要 GitHub Actions 的 `contents: write` 权限创建 release 并上传产物。
- 可能调整 `scripts/build.sh` 的定位或 README 中的构建说明。
- 发布流程依赖 git tag、GoReleaser、GitHub Release 和 `GITHUB_TOKEN`。
