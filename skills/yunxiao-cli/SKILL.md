---
name: yunxiao-cli
description: 当用户需要通过本机 yunxiao CLI 操作云效时使用，包括认证、配置、代码库、合并请求、流水线、工作项、搜索、原始 OpenAPI 调用、扩展开发或只读在线冒烟检查。
---

# Yunxiao CLI

把本机 `yunxiao` 命令当作云效操作的真实入口。优先使用一等命令；只有没有封装命令时才用 `yunxiao api`。

## 快速流程

1. 先确认 CLI、认证和当前配置：
   ```sh
   yunxiao --help
   yunxiao auth status
   yunxiao config list
   ```
2. 如果未登录，使用其中一种方式：
   ```sh
   yunxiao auth login
   printf '%s' "$YUNXIAO_TOKEN" | yunxiao auth login --with-token
   ```
3. 命令缺上下文时再设置配置：
   ```sh
   yunxiao config set endpoint openapi-rdc.aliyuncs.com
   yunxiao config set organization <organization>
   yunxiao config set project <project>
   yunxiao config set repo <repo> --scope repo
   ```
4. 常见操作优先用一等命令：
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

## 原始 API

没有一等命令时，用 `yunxiao api <method> <path>` 调 OpenAPI：

```sh
yunxiao api GET /oapi/v1/example --jq '.data'
yunxiao api POST /oapi/v1/example --body '{"name":"demo"}'
yunxiao api PATCH /oapi/v1/example --input payload.json
```

保持 stdout 可解析：`yunxiao api` 默认只把原始响应体写到 stdout；`--verbose` 诊断信息写到 stderr。

## 扩展

扩展是本地可执行文件，命名为 `yunxiao-<name>`：

```sh
yunxiao extension create team-report
yunxiao extension install ./yunxiao-team-report --yes
yunxiao team-report --since 7d
```

不要把 PAT 传给扩展。扩展代码应调用 `yunxiao api`，由宿主 CLI 读取凭据并注入 HTTP 请求。

## 在线检查

- 冒烟检查优先用只读命令：`auth status`、`repo list`、`mr view`、`mr diff`、`pipeline list`。
- 真实 OpenAPI 检查优先使用 `YUNXIAO_ENDPOINT=https://openapi-rdc.aliyuncs.com`。
- 没有实际凭据和命令输出时，不要声称已经完成在线验证。
