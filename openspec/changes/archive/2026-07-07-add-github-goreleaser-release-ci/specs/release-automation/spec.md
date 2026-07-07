## ADDED Requirements

### Requirement: Tag-triggered GitHub Release

系统 SHALL 在推送匹配 `v*` 的 git tag 时，通过 GitHub Actions 自动运行 GoReleaser 并创建或更新对应的 GitHub Release。

#### Scenario: Push version tag starts release

- **WHEN** 维护者推送 `v0.1.0` 这类 `v*` tag 到 GitHub
- **THEN** GitHub Actions MUST 启动 release workflow
- **AND** workflow MUST 使用 GoReleaser 执行正式发布
- **AND** 发布产物 MUST 上传到该 tag 对应的 GitHub Release

#### Scenario: Branch push does not publish release

- **WHEN** 维护者推送普通分支提交
- **THEN** workflow MUST NOT 创建 GitHub Release
- **AND** workflow MUST NOT 上传正式发布产物

### Requirement: GoReleaser build metadata

系统 SHALL 通过 GoReleaser 向 `cmd/yunxiao` 注入版本、提交和构建时间，并保持 `yunxiao --version` 能展示发布版本信息。

#### Scenario: Released binary reports tag version

- **WHEN** 用户下载 GitHub Release 中的 `yunxiao` 二进制并运行 `yunxiao --version`
- **THEN** 输出 MUST 包含该 release 对应的版本号
- **AND** 输出 MUST 包含构建提交或构建时间中的至少一项可追踪信息

### Requirement: Cross-platform release artifacts

系统 SHALL 为当前支持的平台生成可下载归档产物，并为所有产物生成 checksum。

#### Scenario: Release contains supported platform archives

- **WHEN** `v*` tag 发布成功
- **THEN** GitHub Release MUST 包含 Darwin amd64、Darwin arm64、Linux amd64、Linux arm64 和 Windows amd64 的 `yunxiao` 归档产物
- **AND** GitHub Release MUST 包含 checksum 文件

### Requirement: Release workflow documentation

仓库文档 SHALL 说明本地构建、snapshot 检查和正式 tag 发布的职责边界。

#### Scenario: Maintainer follows documented release flow

- **WHEN** 维护者阅读 README 的发布说明
- **THEN** 文档 MUST 指明正式发布由 `v*` tag 触发 GitHub Actions
- **AND** 文档 MUST 指明 GitHub Release 是正式产物下载位置
- **AND** 文档 MUST 区分本地构建或 snapshot 检查与正式发布
