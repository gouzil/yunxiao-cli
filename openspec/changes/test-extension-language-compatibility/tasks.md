## 1. 多语言真实进程测试

- [x] 1.1 在 `test/extensions/yunxiao-{shell,python,node,go}/` 提交四种最小扩展源码，入口实现同一组参数、环境和标准流标记，不增加第三方运行时依赖
- [x] 1.2 新建 `internal/extension/language_compatibility_test.go`；将 fixture 复制到单个 `t.TempDir()` 的 `sources/`，在临时目录按平台编译 Go 入口，并使用临时 `data/` 和 `state/`
- [x] 1.3 严格参考 `cli/cli` 的 build-tag 和 Windows 分发设计增加表驱动兼容性测试；`acceptance` 层在三平台检查 Bash、Python、Node.js 和 Go，使用真实 `LocalManager` 与 `RealExecRunner` 安装和分发，无后缀脚本经 `sh.exe`、`.exe` 二进制直接执行，任一运行时缺失或统一契约断言失败即报错

## 2. 凭据与 API 契约边界

- [x] 2.1 补充共享管理器断言，确认显式构造的扩展环境只包含扩展身份和非敏感执行上下文，不注入 PAT 等宿主凭据
- [x] 2.2 核对现有 `yunxiao api` 测试持续覆盖宿主鉴权、原始 JSON stdout 和诊断 stderr；仅在覆盖缺口存在时补充一次共享断言

## 3. fixture 质量门禁与 CI

- [x] 3.1 更新 `.pre-commit-config.yaml`，让 Shell/Go fixture 进入现有 `shfmt`/`gofmt`，并以固定版本增加 Python 的 Ruff format/check 和 Node.js 的 Prettier check
- [x] 3.2 更新 `.github/workflows/ci.yml`，在 Linux、macOS 和 Windows 都显式配置 Python 和 Node.js 主版本，并统一运行 `go test -tags=acceptance ./...`；Linux 继续运行 fixture 语法和静态检查
- [x] 3.3 运行 `prek run --all-files`、默认 `go test ./...`、带 tag 的目标多语言测试、Windows acceptance 交叉编译、`go vet ./...` 和 `openspec validate test-extension-language-compatibility --strict`，确认默认单测不报告兼容性测试跳过、fixture 可格式化检查、临时产物不写入工作区且全部检查通过
