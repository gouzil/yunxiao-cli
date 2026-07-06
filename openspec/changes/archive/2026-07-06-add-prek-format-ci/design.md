## Context

`yunxiao-cli` 是 Go CLI 仓库，当前已有 `go test ./...`、`go vet ./...`、`go build ./cmd/yunxiao` 和基础 GitHub Actions CI。仓库还没有 `prek` 配置，格式和文本文件卫生问题主要依赖人工发现，容易在 PR 里产生低价值返工。

本变更只补一个轻量质量门禁：本地可运行，CI 也运行同一套检查。它不改变 CLI 运行时代码，不接管 release workflow，也不引入 Go lint 框架。

## Goals / Non-Goals

**Goals:**

- 新增 `.pre-commit-config.yaml` 作为 `prek` 配置入口。
- 用 `prek` 检查 Go 文件是否满足 `gofmt`。
- 用 `prek` 检查 Shell 文件是否满足 `shfmt`。
- 用 `prek` 检查 YAML 文件是否满足 `yamlfmt`，并保留 YAML 语法检查。
- 用 `prek` 运行 `typos` 拼写检查。
- 用 `prek` 检查文本文件结尾换行、行尾空白和混用换行符。
- 在 GitHub Actions CI 中运行 `prek run --all-files`。
- 保持检查集合短小，先覆盖确定会减少返工的问题。

**Non-Goals:**

- 不新增 `golangci-lint`、ShellCheck、yamllint、自定义 lint 平台或 Makefile。
- 不把 release workflow、GoReleaser、制品上传纳入本变更。
- 不改变现有命令、API 请求、输出结构、认证或扩展执行行为。
- 不要求 CI 在每个 OS matrix job 中重复跑同一套文本检查。

## Decisions

### 1. 使用 `.pre-commit-config.yaml`

`prek` 支持 `.pre-commit-config.yaml` 和 `prek.toml`，本变更使用 `.pre-commit-config.yaml`。

理由：这是 `prek sample-config` 直接产出的默认格式，也兼容现有 pre-commit hook 生态；后续开发者不需要学习仓库私有格式。

备选方案：

- 使用 `prek.toml`：更贴近 `prek`，但仓库没有现成约定，且可复用示例更少。
- 自写 `scripts/check-format.sh`：短期可行，但会把 hook 选择、文件过滤和 CI 调用重新实现一遍。

### 2. `gofmt` 使用 local hook

Go 格式检查使用 `repo: local` hook 调用 `gofmt`，限定 Go 文件。

理由：Go 工具链已经由仓库和 CI 提供，直接用标准 `gofmt` 最少依赖。实现时可用 `gofmt -w` 作为 hook 行为；CI 中如果 hook 修改文件，`prek` 会失败并展示需要提交的格式化 diff。

备选方案：

- 引入第三方 Go formatter hook：没有必要，`gofmt` 已经是标准工具。
- 只在 CI 运行 `gofmt -l`：少一个 hook，但本地和 CI 就不是同一套入口。

### 3. 文本检查复用 `pre-commit-hooks`

文本文件卫生检查复用 `pre-commit/pre-commit-hooks` 中的 `end-of-file-fixer`、`trailing-whitespace` 和 `mixed-line-ending`。

理由：这些 hook 已经覆盖文件结尾换行、行尾空白和混用换行符，避免为通用文本规则写维护脚本。

备选方案：

- 写本地脚本检查 EOF 和空白：依赖更少，但需要自己处理二进制文件、Windows 换行和文件过滤。
- 只跑 `git diff --check`：能发现行尾空白，但不能完整覆盖文件结尾换行和混用换行符。

### 4. Shell、YAML 和拼写检查使用现成 hook

Shell 格式化使用 `scop/pre-commit-shfmt` 的 `shfmt` hook；YAML 格式化使用 `google/yamlfmt`，并额外启用 `pre-commit-hooks` 的 `check-yaml`；拼写检查使用 `crate-ci/typos`。

理由：这三个 hook 覆盖常见低价值返工，且都能由 `prek` 管理安装；不需要在仓库里维护脚本。`yamlfmt` hook 需要 Go，因此 format workflow 显式设置 Go。

备选方案：

- 引入 ShellCheck 或 yamllint：检查更强，但会把本变更从格式门禁扩成 lint 规则治理。
- 用 Prettier 格式化 YAML：可行，但会引入 Node 工具链；仓库当前主要是 Go，`yamlfmt` 更贴近现有工具链。

### 5. CI 用独立 format workflow

新增 `.github/workflows/fmt.yml`，运行在 `ubuntu-latest`，步骤为 checkout、通过 `j178/prek-action@v2` 执行 `prek run --all-files`。

理由：格式和文本检查与 OS 无关，跑一次即可；三系统 Go matrix 继续负责测试、vet 和编译。

备选方案：

- 在每个 OS matrix job 都跑 `prek`：反馈更重复，CI 时间更长。
- 只要求开发者本地运行 `prek`：不能保证 PR 合并前自动拦截。

## Risks / Trade-offs

- hook 仓库需要 CI 拉取并安装 -> 固定 `rev`，并使用 `j178/prek-action@v2` 安装和执行 `prek`。
- `gofmt -w` 会修改工作区 -> 本地是期望行为；CI 中修改文件会导致检查失败，提示提交格式化结果。
- `typos --write-changes` 会自动修正常见英文拼写 -> 若遇到项目专有词误报，再用配置白名单处理。
- 当前仓库可能已有格式或换行基线问题 -> 实现阶段先运行 `prek run --all-files`，把必要的格式化/换行修正和配置一起提交。
