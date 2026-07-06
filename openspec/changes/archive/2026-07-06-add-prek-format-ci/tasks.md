## 1. Prek 配置

- [x] 1.1 新增 `.pre-commit-config.yaml`，配置 `pre-commit/pre-commit-hooks` 的 `end-of-file-fixer`、`trailing-whitespace` 和 `mixed-line-ending`。
- [x] 1.2 在 `.pre-commit-config.yaml` 中新增 local `gofmt` hook，限定 `.go` 文件并调用标准 `gofmt`。
- [x] 1.3 新增 `shfmt` hook，格式化 Shell 脚本。
- [x] 1.4 新增 `yamlfmt` 和 `check-yaml` hook，格式化并校验 YAML 文件。
- [x] 1.5 新增 `typos` hook，检查常见英文拼写错误。
- [x] 1.6 运行 `prek validate-config` 和 `prek list`，确认配置可被 `prek` 读取且 hook 列表正确。

## 2. CI 接入

- [x] 2.1 新增 `.github/workflows/fmt.yml`，定义运行在 `ubuntu-latest` 的 `fmt` job。
- [x] 2.2 在 `fmt` job 中使用 `j178/prek-action@v2` 执行 `prek run --all-files`。
- [x] 2.3 在 `fmt` job 中设置 Go，支持 `yamlfmt` hook 安装。
- [x] 2.4 确认现有三系统 Go validation job 仍然运行 `go test ./...`、`go vet ./...` 和 `go build ./cmd/yunxiao`，不引入 release 副作用。

## 3. 基线修正

- [x] 3.1 运行 `prek run --all-files`，修正当前仓库中被 hook 修改或报告的 Go/Shell/YAML 格式、拼写、文件结尾换行、行尾空白和混用换行符问题。
- [x] 3.2 复跑 `prek run --all-files`，确认所有 hook 通过。

## 4. 验证

- [x] 4.1 运行 `go test ./...`。
- [x] 4.2 运行 `go vet ./...`。
- [x] 4.3 运行 `go build ./cmd/yunxiao`。
- [x] 4.4 运行 `openspec validate add-prek-format-ci --strict`。
