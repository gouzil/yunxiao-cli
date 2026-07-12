## ADDED Requirements

### Requirement: 代表性语言兼容性矩阵
项目 MUST 提供带 `integration` build tag 的自动化兼容性测试，使用 Shell、Python、Node.js 和 Go 实现的真实扩展入口覆盖解释型脚本与编译型二进制，并通过生产使用的本地安装和进程分发路径执行这些入口；默认单元测试 MUST NOT 通过运行时环境变量把该测试报告为跳过。

#### Scenario: 四种入口完成本地安装与执行
- **WHEN** 兼容性测试为 Shell、Python、Node.js 和 Go 分别创建符合 `yunxiao-<name>` 命名约定的可执行入口
- **THEN** 每个入口都能由 `InstallLocal` 安装，并能使用短名称通过 `Dispatch` 成功执行

#### Scenario: 默认单元测试不报告兼容性测试跳过
- **WHEN** 开发者运行不带 build tag 的 `go test ./...`
- **THEN** 多语言真实进程测试不进入默认单元测试集合，测试输出中不会以环境变量未设置为由报告该测试跳过

### Requirement: 统一扩展进程契约
兼容性测试 MUST 对每种代表性语言应用同一组断言，至少覆盖参数传递、扩展身份与工作目录环境、组织/项目/仓库上下文以及 stdout/stderr 转发。

#### Scenario: 各语言收到相同调用上下文
- **WHEN** 测试使用固定参数和 `ExecutionContext` 分发任一种代表性语言扩展
- **THEN** 扩展输出包含原样参数、正确的 `YUNXIAO_EXTENSION`、`YUNXIAO_EXTENSION_NAME`、`YUNXIAO_EXTENSION_DIR`、`YUNXIAO_ORGANIZATION`、`YUNXIAO_PROJECT` 和 `YUNXIAO_REPO` 值

#### Scenario: 标准输出与错误输出保持分离
- **WHEN** 任一种代表性语言扩展分别向 stdout 和 stderr 写入可识别标记
- **THEN** 调用方在对应输出流中收到标记，且两个流不会被合并

### Requirement: 凭据边界复用共享验证
默认单元测试 MUST 验证扩展管理器不会把宿主凭据作为扩展专用环境变量显式注入；OpenAPI 鉴权、原始 JSON stdout 和诊断 stderr 的行为 MUST 继续由共享的 `yunxiao api` 命令测试验证，而不是在每种语言 fixture 中重复实现。

#### Scenario: 扩展环境不包含显式凭据
- **WHEN** 管理器构造扩展分发环境
- **THEN** 环境只包含扩展身份和已声明的非敏感执行上下文，不包含 PAT 或其他宿主凭据值

#### Scenario: API 桥测试保持单一来源
- **WHEN** 测试套件验证扩展所依赖的 `yunxiao api` 契约
- **THEN** 它使用现有 API 命令测试验证宿主鉴权及 stdout/stderr 边界，不要求四种语言 fixture 各自启动 HTTP 测试服务

### Requirement: CI 确定性执行
CI MUST 在一个明确的平台任务中提供并执行多语言测试所需的 Bash、Python、Node.js 和 Go 运行时，并通过 `go test -tags=integration ./...` 纳入完整矩阵；缺少任一声明运行时 MUST 导致失败。

#### Scenario: Linux CI 运行完整语言矩阵
- **WHEN** Linux CI job 执行测试
- **THEN** CI 显式准备 Python 与 Node.js、复用已配置的 Go 和 Bash，并使用 `integration` build tag 运行完整四语言兼容性测试

#### Scenario: 必需运行时缺失
- **WHEN** 使用 `integration` build tag 运行多语言兼容性测试但找不到任一声明运行时
- **THEN** 测试失败并指出缺失的运行时，而不是跳过对应语言后报告成功

### Requirement: 测试资产保持临时和隔离
代表性语言扩展源码 MUST 作为可审阅 fixture 提交到 `test/extensions/`，并使用对应语言的格式化和静态检查工具验证。测试执行时 MUST 在临时目录中复制、构建和安装这些 fixture，测试结束后由测试框架清理，并且不得依赖真实云效账号、网络请求或持久用户配置。

#### Scenario: fixture 源码进入质量门禁
- **WHEN** 格式化和静态检查工作流执行
- **THEN** Shell fixture 通过 `shfmt` 和 `bash -n`，Python fixture 通过 Ruff format/check，Node.js fixture 通过 Prettier 和 `node --check`，Go fixture 通过 `gofmt` 和 `go vet`

#### Scenario: 离线执行后清理
- **WHEN** 多语言兼容性测试执行完成
- **THEN** 运行副本、二进制、安装数据和状态都位于测试临时目录中并被清理，已提交 fixture 保持不变，且执行过程不访问云效网络服务
