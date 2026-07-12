## Context

扩展管理器只认识名为 `yunxiao-<name>` 的平台可执行入口，`Dispatch` 最终通过 `os/exec` 启动该入口，并转发参数、标准流和 `YUNXIAO_*` 非敏感上下文。现有 `manager_test.go` 使用 fake runner 验证管理器行为，能够快速覆盖分支，但没有经过操作系统的解释器、shebang 和真实进程 I/O；现有 CI 则只显式安装 Go。

这项变更只补充契约证据，不增加语言注册表、SDK、运行时安装器或新的扩展模板。

### 目录规划

```text
internal/extension/
└── language_compatibility_test.go   # 复制 fixture、真实安装/分发与统一断言

test/extensions/
├── yunxiao-shell/
│   └── yunxiao-shell                # Bash 可执行入口
├── yunxiao-python/
│   └── yunxiao-python               # Python 可执行入口
├── yunxiao-node/
│   └── yunxiao-node                 # Node.js 可执行入口
└── yunxiao-go/
    └── main.go                      # 测试时编译为 yunxiao-go

.pre-commit-config.yaml              # fixture 格式化与静态检查

.github/workflows/
├── fmt.yml                          # 复用 prek 执行格式检查
└── ci.yml                           # Linux job 准备运行时、语法检查并执行兼容性测试

$TMPDIR/                             # 由 t.TempDir() 管理，不进入仓库
├── sources/yunxiao-<language>/      # 已提交 fixture 的运行副本
├── data/                            # LocalManager 的临时安装数据
└── state/                           # LocalManager 的临时状态
```

Go 驱动测试沿用当前仓库的同包就近放置惯例；跨语言输入则集中在顶层 `test/extensions/`，目录名和入口名直接遵守 `yunxiao-<name>` 安装契约，便于人工运行和代码审查。测试将 fixture 复制到单个 `t.TempDir()` 后再设置权限和编译 Go，确保测试不污染源码树；安装目录和状态目录也从该临时根目录派生。

## Goals / Non-Goals

**Goals:**

- 用 Shell、Python、Node.js、Go 四种入口证明同一套扩展安装与执行契约。
- 让 CI 对缺少声明运行时、参数或环境丢失、标准流转发错误等回归稳定失败。
- 复用真实 `LocalManager` 和 `RealExecRunner`，使测试经过生产执行路径。
- 保持测试临时、自包含，不维护四个示例仓库。

**Non-Goals:**

- 不承诺 CLI 自动安装 Python、Node.js 或其他语言运行时。
- 不为每种语言新增 `extension create` 模板或 SDK。
- 不在四个入口中重复测试 `yunxiao api` 的 HTTP、鉴权和 JSON 过滤逻辑；该契约继续由 API 命令测试负责。
- 不改变 Windows 的 `.exe` 入口约定，也不声称解释型脚本矩阵覆盖所有平台。

## Decisions

### 1. 提交四种入口源码，由一个 Go 集成测试统一驱动

Shell、Python、Node.js 和 Go 源码分别提交到 `test/extensions/yunxiao-<language>/`。测试实现放在 `internal/extension/language_compatibility_test.go`，将 fixture 复制到同一个 `t.TempDir()` 根目录的 `sources/` 下，用当前 Go 工具链把 `main.go` 编译为 `yunxiao-go`；同一临时根目录下的 `data/` 和 `state/` 分别传给 `LocalManager`。每个入口实现相同的最小协议：接收固定参数，读取扩展上下文环境变量，分别向 stdout 和 stderr 输出可断言的标记。随后测试通过真实 `InstallLocal` 和 `Dispatch` 执行入口，并用表驱动子测试复用断言。

备选方案是由 Go 测试动态拼接四种语言源码；文件更少，但源码无法被常规 formatter、linter 和人工审查直接覆盖。提交最小 fixture 会增加四个小源码文件，却让“支持哪些入口、每个入口实际做什么”成为可见且可独立检查的仓库资产。

### 2. fixture 使用语言原生或现有工具检查

`.pre-commit-config.yaml` 继续作为格式化入口：现有 `shfmt` 和 `gofmt` 直接覆盖 Shell 与 Go；新增 Ruff 的 format/check 覆盖 Python，新增 Prettier check 覆盖 Node.js。CI 再执行不修改文件的语法/静态检查：`bash -n`、Ruff check、`node --check` 和 Go vet/build。工具版本固定在 pre-commit hook 或 setup action 中，避免依赖开发机全局版本。

不引入 pytest、npm 应用脚手架、ESLint 配置或嵌套 Go module。fixture 没有业务依赖：Python 用标准库，Node.js 用内置模块，Go 直接属于主 module；这些额外工程文件不会提升本次兼容性证据。

### 3. 严格参考 `cli/cli` 用 acceptance build tag 分层

`cli/cli` 的默认扩展单测直接运行，不用环境变量在测试函数内调用 `t.Skip`，并在 manager 测试中对 Windows 的 `sh -c` 分发和本地安装路径文件分支单独断言；需要真实系统和 Bash 的端到端流程放在 `//go:build acceptance` 层，由 CI 显式选择。这里使用同名 `//go:build acceptance`，不加 `!windows` 编译排除；默认 `go test ./...` 在三平台运行 manager 单测，Linux CI 显式准备 Python 和 Node.js，并运行 `go test -tags=acceptance ./...`。任何必需解释器缺失都直接失败。Shell 使用 runner 自带 Bash，Go 复用已有 `setup-go`。

与 `cli/cli` 一样，Windows 覆盖由默认 manager 单测负责，真实脚本 acceptance 由明确的 Linux CI 步骤负责；测试文件本身不再用 `!windows` 隐藏。也不保留运行时环境变量或 `t.Skip`，避免默认单测以“通过”掩盖关键测试未执行。

### 4. 只跨语言验证语言相关的进程边界

四种入口统一验证本地安装、短名称分发、argv、`YUNXIAO_EXTENSION_*` 与组织/项目/仓库上下文、stdout 和 stderr。凭据不作为扩展环境变量注入的规则保留为一次共享管理器断言；`yunxiao api` 的宿主鉴权和原始 JSON stdout 继续复用现有命令测试。

把完整 API fake server 场景复制到四种语言只会重复证明“这些语言能启动子进程”，不会增加鉴权桥本身的覆盖。测试分层后，任一层失败都能更直接定位到扩展执行边界或 API 命令边界。

## Risks / Trade-offs

- [Linux acceptance 通过不能证明每种语言在每个平台都可执行] → 与 `cli/cli` 一样由三平台默认 manager 单测覆盖平台分支，Linux acceptance 覆盖真实脚本进程协议。
- [GitHub runner 镜像升级可能改变预装工具] → CI 显式使用 setup action 固定 Python 和 Node.js 主版本，启用测试时缺失 Bash 或解释器直接失败。
- [新增 Ruff 和 Prettier 检查增加 CI 下载时间] → 通过固定版本的 pre-commit hook 缓存工具，不创建 Python 或 Node.js 应用依赖树。
- [编译 Go fixture 增加少量测试时间] → 仅在显式多语言测试中编译一个最小程序，产物写入临时目录。
- [四种 fixture 输出格式可能逐渐分叉] → 每种入口只实现相同的几行标记协议，所有断言集中在一个表驱动测试中。
