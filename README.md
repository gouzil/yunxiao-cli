# Yunxiao CLI

Yunxiao CLI 是面向云效 OpenAPI 的命令行客户端。它保留接近 GitHub CLI (`gh`) 的使用习惯，同时使用云效自己的产品名：`mr`、`pipeline`、`run`、`workitem` 等。

当前实现覆盖认证、配置、原始 API 调用、代码仓库、合并请求、流水线、项目协作、搜索和本地扩展。命令处理层使用类型化 request/result model，不把松散的 `map[string]any` 暴露成主要接口。

## 安装和构建

需要 Go `1.25.8` 或更高版本。

```sh
go build -o bin/yunxiao ./cmd/yunxiao
```

构建多平台产物：

```sh
./scripts/build.sh
```

产物会写入 `dist/`，文件名形如 `yunxiao-darwin-arm64`、`yunxiao-linux-amd64` 和 `yunxiao-windows-amd64.exe`。

## 快速开始

登录云效：

```sh
yunxiao auth login
```

非交互环境可以从 stdin 传入 PAT：

```sh
printf '%s' "$YUNXIAO_TOKEN" | yunxiao auth login --with-token
```

配置常用上下文：

```sh
yunxiao config set endpoint openapi-rdc.aliyuncs.com
yunxiao config set organization <organization>
yunxiao config set project <project>
yunxiao config set repo <repo> --scope repo
```

查看认证和配置：

```sh
yunxiao auth status
yunxiao config list
```

## 常用命令

```sh
yunxiao repo list
yunxiao repo view <repo>
yunxiao mr list
yunxiao mr view <mr>
yunxiao mr diff <mr>
yunxiao pipeline list
yunxiao pipeline run <pipeline-id> --branch <branch>
yunxiao run watch <run-id>
yunxiao workitem list
yunxiao search repo <query>
```

未封装成一等命令的 OpenAPI 可以直接调用：

```sh
yunxiao api GET /oapi/v1/example --jq '.data'
```

`yunxiao api` 默认只把原始响应体写到 stdout；诊断信息通过 `--verbose` 写到 stderr，方便脚本和扩展稳定解析 stdout。

## 配置来源

配置按优先级合并：

1. 命令行参数：`--endpoint`、`--organization`、`--project`、`--repo`
2. 环境变量：`YUNXIAO_ENDPOINT`、`YUNXIAO_ORGANIZATION`、`YUNXIAO_PROJECT`、`YUNXIAO_REPO`
3. 仓库配置：`.yunxiao/config.json`
4. 全局配置：`~/.config/yunxiao/config.json`
5. 默认值：`endpoint=openapi-rdc.aliyuncs.com`

认证信息优先写入系统 keyring；不可用时回退到 `~/.config/yunxiao/credentials.json`。

## 扩展

扩展是本地可执行文件，不是 Go plugin。扩展仓库、目录和入口文件使用 `yunxiao-<name>` 命名。

```sh
yunxiao extension create team-report
yunxiao extension install ./yunxiao-team-report --yes
yunxiao team-report --since 7d
```

扩展不会收到云效 PAT。扩展应调用 `yunxiao api`，由宿主 CLI 读取凭据并注入 HTTP 请求。

## Agent Skill

仓库内置了 `skills/yunxiao-cli`，用于让支持 agent skills 的工具通过本机 `yunxiao` 命令操作云效。使用 Skills CLI 安装：

```sh
npx skills add gouzil/yunxiao-cli --skill yunxiao-cli
```

安装后新开 agent 会话，直接提到云效 CLI、MR、流水线、工作项、搜索或在线冒烟检查时会自动触发；也可以显式使用 `$yunxiao-cli`。

## 文档

- [命令参考](docs/command-reference.md)
- [从 `gh` 迁移](docs/migration-guide.md)
- [输出契约](docs/output-contracts.md)
- [Service 接口](docs/api-services.md)
- [扩展开发](docs/extension-authoring.md)

## 开发检查

```sh
./scripts/test.sh
./scripts/vet.sh
```

`repo unarchive` 和遗留跨仓库代码搜索 API 仍按待确认能力处理；需要真实云效环境验证后再提升为稳定能力。
