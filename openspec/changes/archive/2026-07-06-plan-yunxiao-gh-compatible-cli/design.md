## Context

云效开放 API 覆盖代码管理、合并请求、流水线、项目协作、工作项、成员、SSH Key、搜索等域，足以支撑一个面向日常研发协作的终端 CLI。GitHub CLI 的命令结构可以作为交互体验参考，但云效产品模型与 GitHub 不完全一致：云效的合并请求对应 `gh pr`，工作项和项目协作承接 `gh issue`/`gh project` 的核心场景，流水线承接 `gh workflow`/`gh run` 的核心场景。

当前仓库尚无实际 CLI 源码，本设计先定义能力边界、命令树、共享架构和落地顺序，为后续实现提供稳定契约。

实现语言明确为 Go，目标是产出与 `gh` 类似的跨平台单二进制 CLI。Go 版本、依赖和工程结构应优先服务于稳定分发、低运行时依赖、易测试和易集成 CI。

每个 CLI 命令到云效 API 的映射以 `api-map.md` 为实现依据。实现任务开始前必须先核对对应 API 链接、参数和响应结构；`api-map.md` 标记为“待确认”或“旧版”的命令必须先完成单独 smoke 验证。每个 CLI 命令的默认终端展示以 `output-map.md` 为输出契约，除非实现阶段确认云效 API 字段不可得并同步更新该文档。

## Goals / Non-Goals

**Goals:**

- 提供一个云效原生 CLI，覆盖日常开发最常用的认证、代码库、合并请求、流水线、工作项、项目和搜索操作。
- 保持 `gh` 风格的易用性：资源命令分组、默认仓库上下文、可脚本化 JSON 输出、浏览器打开、别名和补全。
- 使用云效术语暴露命令，避免将 GitHub 专有概念强行映射到云效。
- 建立统一 HTTP 客户端、配置加载、认证注入、分页处理、错误处理和输出渲染。
- 使用 Go 实现跨平台单二进制 CLI，并建立清晰的 `cmd/`、`internal/`、`pkg/` 边界。
- 支持后续逐步扩展更多云效 API 域，而不破坏已定义命令体验。
- 为每个命令维护明确的云效 API 映射和官方文档链接。
- 为每个命令维护明确的默认输出格式、字段和空结果/错误展示。

**Non-Goals:**

- 不一比一复刻 GitHub CLI 的全量命令。
- 不实现 GitHub 特有产品能力，例如 Codespaces、Gist、Copilot、GitHub Attestation、GitHub Extension/Skill 生态。
- 不在第一阶段实现完整 TUI、daemon、后台同步或本地缓存数据库。
- 不引入 Node.js、Python 或 shell 脚本作为 CLI 主实现语言。
- 不把云效“制品/应用交付”直接伪装成 GitHub Release；如需实现，应作为 `artifact`、`package` 或 `deploy` 后续能力单独建模。
- 不实现没有明确云效 API 依据的服务端命令；此类能力只能作为本地命令或后续待确认项。

## Decisions

### 1. 使用 Go 构建单二进制 CLI

CLI MUST 使用 Go 作为主实现语言。建议工程结构：

```text
cmd/yunxiao/
internal/cmd/
internal/api/
internal/config/
internal/auth/
internal/context/
internal/output/
internal/browser/
internal/git/
internal/yunxiao/
```

命令框架建议采用 `spf13/cobra`，配置读写可采用 `spf13/viper` 或小型自有配置层，HTTP 层基于标准库 `net/http` 封装。凭据存储优先使用系统 keychain；无法使用时回退到权限受限的配置文件。

理由：Go 与 `gh` 的技术路线一致，适合分发单文件 CLI；Cobra 在 Go CLI 生态成熟，能稳定支持子命令、flags、help、completion。

备选方案：

- 使用 Node.js/TypeScript：开发速度快，但运行时和分发复杂度更高。
- 使用 Python：脚本化方便，但跨平台单二进制分发和凭据存储不如 Go 直接。
- 纯标准库手写命令框架：依赖少，但命令树、补全和 help 维护成本高。

### 2. 交互、TUI 和终端组件保持 `gh` 技术栈风格

CLI SHOULD 采用与 `gh` 当前 Go 生态一致的终端组件组合，但第一阶段只使用轻量交互能力，不实现完整 TUI 应用。

建议依赖分层：

- `charm.land/huh/v2`：用于 `auth login`、删除确认、MR 创建等表单式交互。
- `charm.land/bubbles/v2` 和 `charm.land/bubbletea/v2`：用于后续需要状态机的交互组件，例如选择器、分页列表或 watch 状态。
- `charm.land/lipgloss/v2`：用于终端样式、颜色、状态符号和表格细节。
- `github.com/charmbracelet/glamour`：用于 Markdown 内容渲染，例如 MR 描述、评论和工作项描述。
- `github.com/briandowns/spinner`：用于 token 校验、API 长请求、流水线 watch 等短周期等待提示。
- `github.com/zalando/go-keyring`：用于跨平台系统凭据存储。
- `golang.org/x/term`、`github.com/mattn/go-isatty`、`github.com/mattn/go-colorable`：用于密码输入、TTY 检测和 Windows 颜色兼容。

`github.com/AlecAivazis/survey/v2` 在 `gh` 中存在，但 `yunxiao` 第一阶段 SHOULD 避免同时维护 `huh` 和 `survey` 两套 prompt 抽象；如遇到 `huh` 无法满足的单点交互，再单独评估是否引入。`github.com/rivo/tview` 和 `github.com/gdamore/tcell/v2` 也出现在 `gh` 依赖中，但它们更适合完整终端 UI，不纳入第一阶段直接依赖。

所有交互命令 MUST 支持非交互模式。交互层不得承载业务逻辑；命令参数解析、API 调用和输出渲染必须能在无 TTY、CI 和脚本场景下运行。

理由：这能复用 `gh` 已验证的 Go 终端生态，同时控制第一阶段复杂度。云效 CLI 当前最需要的是可靠 prompt、确认、Markdown 渲染、spinner 和凭据存储，而不是全屏 TUI。

备选方案：

- 完全不引入交互组件：依赖更少，但 `auth login`、确认操作和创建表单体验较差。
- 直接引入完整 TUI 框架：能力更强，但会扩大首期测试矩阵和交互状态复杂度。
- 同时使用 `huh` 和 `survey`：迁移 `gh` 经验更直接，但 prompt 风格和测试辅助会分裂。

### 3. 命令树按云效资源域建模

命令根建议为 `yunxiao`，一级命令按云效资源组织：

```text
yunxiao auth
yunxiao config
yunxiao api
yunxiao repo
yunxiao branch
yunxiao commit
yunxiao file
yunxiao mr
yunxiao pipeline
yunxiao run
yunxiao project
yunxiao workitem
yunxiao search
yunxiao ssh-key
```

理由：这样既保留 `gh` 的资源分组体验，又避免 `pr`、`issue`、`workflow` 等 GitHub 术语和云效模型产生偏差。

备选方案：

- 直接复刻 `gh` 命令名：迁移成本低，但会混淆云效语义，尤其是 issue/project/release/workflow。
- 完全按云效 API 路径生成命令：覆盖完整但体验差，难以形成高频工作流。

### 4. 统一上下文解析

CLI MUST 支持从命令参数、配置文件、环境变量和当前 Git remote 推导企业、组织、项目、仓库、分支等上下文。解析优先级为：

```text
显式参数 > 环境变量 > 当前仓库配置 > 全局配置 > Git remote 推导
```

理由：这与 `gh` 的默认仓库体验一致，同时适配云效 API 对企业/组织/仓库 ID 的依赖。

### 5. 统一 API 客户端和数据边界

所有命令通过共享云效 HTTP 客户端访问 API。客户端负责：

- 注入 PAT 或后续支持的认证凭据。
- 组装 endpoint、organization、repositoryId、projectId 等上下文。
- 处理分页、重试、限流、超时、错误码和请求 ID。
- 支持 `--json`、`--jq`、`--template` 的结构化输出。
- 支持 `--debug` 或 `--verbose` 输出请求诊断信息，但默认不泄露 token。

命令层只做参数解析、上下文选择、业务调用和输出渲染。

Go 包边界建议：

- `internal/yunxiao` 定义云效 API 请求/响应模型和服务方法。
- `internal/api` 提供底层 HTTP、分页、错误、重试和调试输出。
- `internal/cmd` 只注册 Cobra 命令并调用服务层。
- `internal/output` 负责表格、JSON、jq、模板和终端格式。

人类可读输出和 JSON 输出必须共用同一结构化数据模型。命令不得通过拼接默认输出再反向生成 JSON；默认输出、`--json`、`--jq`、`--template` 都应从同一个 typed result 渲染。

### 6. 合并请求优先级最高

第一阶段优先实现 `mr`，因为它最接近 `gh pr` 的核心价值，也是代码评审工作流最高频入口。建议首批命令：

```text
yunxiao mr list
yunxiao mr view
yunxiao mr create
yunxiao mr diff
yunxiao mr files
yunxiao mr comment
yunxiao mr approve
yunxiao mr changes-requested
yunxiao mr merge
yunxiao mr close
yunxiao mr reopen
```

### 7. 流水线拆分为定义和运行实例

`pipeline` 表示流水线定义，`run` 表示运行实例：

```text
yunxiao pipeline list/view/create/update/delete/run
yunxiao run list/view/log/watch/cancel/retry/stop-task/skip-task
```

理由：这比强行使用 `workflow/run` 更贴近云效，也保留了 GitHub Actions 的使用习惯。

### 8. 工作项承接 issue，项目承接 project

云效没有 GitHub Issue 的一比一模型。CLI SHOULD 使用：

- `workitem` 表示需求、缺陷、任务等可跟踪工作。
- `project` 表示项目空间、迭代、里程碑、成员、标签等协作结构。

如果后续为了迁移用户降低学习成本，可以提供 `issue` alias，但文档和主命令应使用 `workitem`。

### 9. 搜索和 API 作为通用逃生口

`search` 用于高频查询，`api` 用于未封装 API 的调试和脚本化调用。任何新 API 域在正式抽象为一级命令前，都可以先通过 `api` 支撑高级用户。

## Risks / Trade-offs

- [Risk] 云效 API 的资源 ID 和 Git remote URL 之间不总是能可靠互推。→ Mitigation：提供 `repo set-default` 或本地配置绑定，remote 推导失败时给出明确修复命令。
- [Risk] 照顾 `gh` 用户会诱导过度兼容 GitHub 命名。→ Mitigation：主命令使用云效术语，只在低成本处提供 alias。
- [Risk] 工作项、项目、标签与 GitHub issue/project 的语义差异会导致用户误解。→ Mitigation：规格中明确映射边界，并在命令帮助中使用云效术语。
- [Risk] 流水线日志、任务控制、分页列表可能涉及大响应和长轮询。→ Mitigation：日志命令支持分页/流式/时间范围，watch 命令设置轮询间隔和超时。
- [Risk] PAT 泄露风险。→ Mitigation：认证信息存储使用系统凭据管理或权限受限文件，debug 输出默认脱敏。
- [Risk] Cobra/Viper 依赖便利但可能带来全局状态和测试复杂度。→ Mitigation：命令构造函数显式注入 IO、配置和客户端，单元测试不依赖真实全局环境。
- [Risk] 首期范围过大。→ Mitigation：任务拆分为基础框架、MR MVP、repo/search、pipeline、workitem/project、本地体验多个可独立验收批次。
