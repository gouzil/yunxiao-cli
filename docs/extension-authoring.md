# Yunxiao CLI 扩展开发

扩展就是可执行文件。它不是 Go plugin，也不导入 `yunxiao-cli` 的 `internal/` 包。

## 命名

仓库、目录和可执行入口都使用 `yunxiao-<name>` 约定。

```text
yunxiao-team-report/
  yunxiao-team-report
```

用户运行时使用短名称：

```sh
yunxiao team-report --since 7d
```

## 语言支持

只要入口文件能在当前平台执行，扩展可以用任意语言实现。

Python:

```python
#!/usr/bin/env python3
import subprocess

result = subprocess.check_output([
    "yunxiao", "api", "GET", "/oapi/v1/example", "--jq", ".data"
], text=True)
print(result, end="")
```

Node.js:

```javascript
#!/usr/bin/env node
const { execFileSync } = require("node:child_process");

const result = execFileSync("yunxiao", ["api", "GET", "/oapi/v1/example", "--jq", ".data"], {
  encoding: "utf8",
});
process.stdout.write(result);
```

Shell:

```sh
#!/usr/bin/env bash
set -euo pipefail

yunxiao api GET /oapi/v1/example --jq '.data'
```

CLI 安装扩展时不会安装运行时、执行包管理器或编译扩展源码。

## 认证

扩展不会收到个人访问令牌。宿主 CLI 只会传递非敏感上下文变量，例如 `YUNXIAO_ENDPOINT`、`YUNXIAO_ORGANIZATION`、`YUNXIAO_PROJECT` 和 `YUNXIAO_REPO`。

调用云效 OpenAPI 时，扩展应执行 `yunxiao api`。宿主 CLI 会从 keyring 或 `credentials.json` 读取凭据，并把 token 注入 HTTP 请求。

## 本地调试

创建并安装一个本地扩展：

```sh
yunxiao extension create team-report
yunxiao extension install ./yunxiao-team-report --yes
yunxiao team-report
```

本地安装会指向源码目录，所以修改扩展文件后不需要重新安装。

## 安全

安装扩展意味着信任扩展代码。安装或升级前应先审查源码。Yunxiao CLI 不会签名、沙箱隔离或背书第三方扩展。
