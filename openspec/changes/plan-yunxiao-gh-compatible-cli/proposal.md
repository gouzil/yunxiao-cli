## Why

当前项目希望借鉴 GitHub CLI 的高频开发协作体验，为云效提供一个面向终端的统一 CLI。云效开放 API 已覆盖代码库、合并请求、流水线、项目协作、工作项、搜索、SSH Key 等核心域，具备实现类 `gh` 日常工作流的基础。

本变更用于先确定能力边界和实现路径：优先覆盖可直接映射的云效能力，避免照搬 GitHub 特有命令，形成后续实现可执行的规格和任务清单。

## What Changes

- 新增云效 CLI 的能力规划，命令风格参考 `gh`，但资源命名遵循云效语义。
- 明确 CLI 使用 Go 实现，采用与 `gh` 相同的单二进制分发思路，优先使用 Go 生态成熟 CLI/HTTP/测试组件。
- 定义认证和配置能力，支持个人访问令牌、企业/组织标识、默认仓库上下文和 API 调试入口。
- 定义代码库能力，覆盖仓库查询、列表、创建、更新、归档、删除、分支、提交、文件和 SSH Key 等操作。
- 定义合并请求能力，覆盖 MR 创建、查看、列表、差异、文件变更、评论、评审、合并、关闭、重开、更新和状态检查。
- 定义流水线能力，覆盖流水线列表/查看/运行、运行实例查看、日志、终止、重试、任务控制和变量相关操作。
- 定义项目协作能力，用云效工作项、项目、迭代、里程碑、标签等能力承接 `gh issue` 和 `gh project` 的主要场景。
- 定义搜索和通用 API 能力，覆盖代码、仓库、提交、合并请求、工作项等查询场景，并保留低层 `api` 命令。
- 定义本地 CLI 体验能力，包括配置、输出格式、JSON/JQ/模板、浏览器打开、补全和别名。
- 补充每个规划命令对应的云效 API、官方文档链接和确认状态，作为实现阶段的接口依据。
- 明确不做一比一复刻的 GitHub 特有能力：Codespaces、Gist、Copilot、GitHub Attestation、GitHub Extension/Skill 生态、GitHub Release 语义等。

## Capabilities

### New Capabilities

- `cli-auth-config`: 云效 CLI 的认证、配置、上下文解析和低层 API 调试能力。
- `cli-repository`: 云效代码库、分支、提交、文件和 SSH Key 相关 CLI 能力。
- `cli-merge-request`: 云效合并请求相关 CLI 能力。
- `cli-pipeline`: 云效流水线和运行实例相关 CLI 能力。
- `cli-project-workitem`: 云效项目协作、工作项、迭代、里程碑和标签相关 CLI 能力。
- `cli-search-output`: 云效搜索、输出格式、本地易用性和 `gh` 风格体验能力。

### Modified Capabilities

无。

## Impact

- 影响后续 CLI 命令树、配置文件格式、认证存储、HTTP 客户端、输出渲染和错误处理设计。
- 影响后续 Go module、包结构、命令框架、跨平台构建、发布产物和测试组织方式。
- 需要依赖云效开放 API，至少覆盖 Codeup、流水线、Projex、组织/成员、SSH Key、搜索等接口域。
- 需要建立 GitHub CLI 到云效 CLI 的能力映射，避免将 GitHub 专有概念强行暴露为云效命令。
- 需要维护 `api-map.md` 中的命令到 API 映射；旧版 API 或待确认 API 必须先 smoke 再进入正式实现。
- 需要后续实现统一的命令测试、API mock、真实环境 smoke 测试和文档示例。
