# cli-auth-config Specification

## Purpose
TBD - created by archiving change plan-yunxiao-gh-compatible-cli. Update Purpose after archive.
## Requirements
### Requirement: 个人访问令牌认证
CLI SHALL 支持使用云效个人访问令牌完成认证，并在后续 API 请求中自动注入认证信息。

#### Scenario: 使用令牌登录
- **WHEN** 用户执行 `yunxiao auth login` 并输入有效个人访问令牌
- **THEN** CLI MUST 保存认证凭据并能够通过 `yunxiao auth status` 显示已登录状态

#### Scenario: 交互式登录提示
- **WHEN** 用户执行 `yunxiao auth login`
- **THEN** CLI MUST 提示输入云效 endpoint 和个人访问令牌，并在校验 token 时输出校验状态

#### Scenario: 从标准输入登录
- **WHEN** 用户执行 `yunxiao auth login --with-token` 并通过标准输入提供个人访问令牌
- **THEN** CLI MUST 从标准输入读取 token、校验 token、保存凭据，并输出与交互式登录一致的成功摘要

#### Scenario: 认证失败
- **WHEN** 用户提供无效、过期或权限不足的令牌
- **THEN** CLI MUST 输出可读错误，并且不得保存无效凭据

#### Scenario: 禁止明文参数传令牌
- **WHEN** 用户查看 `yunxiao auth login --help`
- **THEN** CLI MUST 不提供通过普通命令行参数传入 token 的选项，并引导用户使用交互输入或 `--with-token`

### Requirement: 配置和上下文解析
CLI SHALL 支持全局配置、仓库级配置、环境变量和命令行参数，并用确定性优先级解析云效 endpoint、企业、组织、项目和代码库上下文。

#### Scenario: 显式参数优先
- **WHEN** 同一个上下文值同时存在于配置文件、环境变量和命令行参数
- **THEN** CLI MUST 使用命令行参数中的值

#### Scenario: 默认仓库上下文
- **WHEN** 用户在已绑定云效代码库的本地 Git 仓库中执行资源命令
- **THEN** CLI MUST 默认使用该本地仓库绑定的云效代码库上下文

### Requirement: 通用 API 命令
CLI SHALL 提供低层 `api` 命令，用于访问尚未封装为一等命令的云效开放 API。

#### Scenario: GET 请求
- **WHEN** 用户执行 `yunxiao api GET <path>`
- **THEN** CLI MUST 使用当前认证和 endpoint 发起请求，并将响应输出到标准输出

#### Scenario: 请求诊断
- **WHEN** 用户为 `api` 命令指定 `--verbose`
- **THEN** CLI MUST 输出请求方法、路径、状态码和请求标识，并且不得明文输出认证令牌
