## Why

当前仓库已有 Go 测试、vet、构建脚本和基础 CI，但还没有统一的格式与文本文件卫生门禁。加入 `prek` 可以在本地和 CI 使用同一套轻量检查，先挡住未运行 `gofmt`、文件缺少结尾换行、行尾空白等低价值返工。

## What Changes

- 新增仓库级 `prek` 配置，覆盖 Go、Shell、YAML 格式检查和常见文本文件换行检查。
- 将 `gofmt` 检查纳入 `prek`，失败时明确提示需要格式化的 Go 文件。
- 增加 `shfmt`、`yamlfmt`、YAML 语法检查和 `typos` 拼写检查。
- 增加文本文件卫生检查，覆盖文件结尾换行、行尾空白等容易污染 diff 的问题。
- 新增独立的 GitHub Actions format workflow，使用 `j178/prek-action@v2` 让 pull request 和 `main` push 都执行同一套门禁。
- 不改变 CLI 运行时行为、不新增 Go 运行时依赖、不扩大到发布自动化。

## Capabilities

### New Capabilities

- `quality-gates`: 定义仓库级 `prek` 质量门禁、Go/Shell/YAML 格式检查、文本文件换行检查、拼写检查，以及 CI 中的执行要求。

### Modified Capabilities

- 无。

## Impact

- 新增 `prek` 配置文件。
- 新增 `.github/workflows/fmt.yml`，通过 `j178/prek-action@v2` 执行 `prek`。
- 可能对现有未格式化文件或文本文件尾部换行问题产生失败反馈，需要实现阶段先修正当前仓库基线。
- 不影响云效 API 请求、认证、命令输出结构、扩展执行协议或 release workflow。
