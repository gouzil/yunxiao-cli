## Why

扩展被定义为“任意语言实现的本地可执行文件”，但现有自动化测试主要通过 Go fake runner 验证分发参数，尚未证明真实的 Shell、Python、Node.js 和 Go 入口都能完成安装与执行。需要一组小型兼容性测试，防止入口发现、环境传递或 I/O 转发的改动悄悄破坏语言无关契约。

## What Changes

- 在 `test/extensions/` 提交 Shell、Python、Node.js、Go 四种可审阅的代表性扩展源码，并增加真实进程兼容性测试，覆盖解释型脚本和编译型二进制。
- 对每种入口统一验证本地安装、短名称分发、参数传递、非敏感上下文环境变量以及 stdout/stderr 转发。
- 验证扩展进程不会收到云效 PAT 等宿主凭据；`yunxiao api` 的鉴权与原始 JSON 契约继续由现有 API 命令测试覆盖，不在每种语言中重复搭建 HTTP 场景。
- 在 CI 中显式提供测试所需运行时，并只在适合执行脚本入口的平台运行多语言矩阵；保留现有跨平台 Go 测试作为 Windows 行为的覆盖。
- 将 fixture 源码接入对应的格式化与静态检查：Shell 使用 `shfmt` 和 `bash -n`，Python 使用 Ruff，Node.js 使用 Prettier 和 `node --check`，Go 使用 `gofmt` 和 `go vet`。

## Capabilities

### New Capabilities

- `extension-language-compatibility-testing`: 定义代表性语言扩展必须通过的安装、执行、环境隔离和 I/O 契约测试。

### Modified Capabilities

无。

## Impact

- 新增 `test/extensions/yunxiao-{shell,python,node,go}/` 源码和 `internal/extension/language_compatibility_test.go`，并更新 `.pre-commit-config.yaml` 与 `.github/workflows/ci.yml`。
- 不修改公开 CLI 命令、扩展清单格式或运行时行为，不引入生产依赖。
- CI 需要可用的 Bash、Python、Node.js 与 Go；运行测试时只在 `t.TempDir()` 下创建副本、安装数据、状态和 Go 编译产物，不修改已提交的 fixture 源码。
