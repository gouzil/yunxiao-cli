# Basic Request Retry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 默认重试只读请求遇到的临时 `EOF`、网络错误、HTTP 429 和 HTTP 5xx，同时确保写请求与永久错误不会重试。

**Architecture:** 继续使用 `internal/api.Client.Do` 作为唯一重试入口。`NewClient` 提供 2 次默认重试预算，`Do` 只在 `GET`/`HEAD` 且错误可恢复时消费该预算，不增加新依赖或新配置层。

**Tech Stack:** Go 1.25.8、标准库 `net/http`、`errors`、`io`、`net`、`httptest`

## Global Constraints

- 最多额外重试 2 次，共最多 3 次请求。
- 仅 `GET` 和 `HEAD` 自动重试。
- 可重试错误仅限 HTTP 429、HTTP 5xx、`io.EOF`、`io.ErrUnexpectedEOF` 和 `net.Error`。
- 等待保持 250ms、500ms；现有 `context` 取消逻辑保持不变。
- 不增加依赖、CLI 参数、随机抖动、`Retry-After` 解析或写请求幂等机制。

## File Map

- `internal/api/client.go`：定义默认重试预算和统一的重试判定。
- `internal/api/client_test.go`：用真实 HTTP 服务验证默认 EOF 重试、方法门禁和永久错误门禁。

---

### Task 1: 安全启用只读请求重试

**Files:**
- Modify: `internal/api/client.go:3-150`
- Test: `internal/api/client_test.go:3-75`
- Include in commit: `docs/superpowers/plans/2026-07-30-basic-request-retry.md`

**Interfaces:**
- Consumes: `NewClient(options ClientOptions) *Client`、`(*Client).Do(context.Context, Request, any) (ResponseMeta, error)`
- Produces: `defaultRetryMax = 2`、`shouldRetry(method string, err error) bool`

- [ ] **Step 1: 写出默认重试 GET/HEAD EOF 的失败测试**

在 `internal/api/client_test.go` 的 imports 中加入 `"sync/atomic"`，并追加：

```go
func TestClientRetriesEOFForReadMethodsByDefault(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			var attempts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if attempts.Add(1) < 3 {
					panic(http.ErrAbortHandler)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := NewClient(ClientOptions{Endpoint: server.URL})
			if _, err := client.Do(context.Background(), Request{Method: method, Path: "/"}, nil); err != nil {
				t.Fatalf("Do(%s) returned error: %v", method, err)
			}
			if got := attempts.Load(); got != 3 {
				t.Fatalf("attempts = %d, want 3", got)
			}
		})
	}
}
```

- [ ] **Step 2: 运行测试并确认因默认重试未启用而失败**

Run:

```bash
go test ./internal/api -run '^TestClientRetriesEOFForReadMethodsByDefault$' -count=1
```

Expected: FAIL，第一次连接中断后返回包含 `EOF` 的错误，实际请求次数为 1。

- [ ] **Step 3: 给客户端设置默认两次重试预算**

在 `internal/api/client.go` 的常量块加入：

```go
defaultRetryMax = 2
```

将 `NewClient` 中的重试次数归一化改为：

```go
retryMax := options.RetryMax
if retryMax == 0 {
	retryMax = defaultRetryMax
}
if retryMax < 0 {
	retryMax = 0
}
```

零值使用默认预算；正数继续覆盖默认值；负数继续表示不重试。

- [ ] **Step 4: 运行测试并确认默认 EOF 重试通过**

Run:

```bash
go test ./internal/api -run '^TestClientRetriesEOFForReadMethodsByDefault$' -count=1
```

Expected: PASS，GET 和 HEAD 均在第三次请求成功。

- [ ] **Step 5: 写出写请求与永久错误不重试的失败测试**

在 `internal/api/client_test.go` 追加：

```go
func TestClientDoesNotRetryEOFForWriteMethods(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		panic(http.ErrAbortHandler)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{Endpoint: server.URL})
	_, err := client.Do(context.Background(), Request{Method: http.MethodPost, Path: "/"}, nil)
	if err == nil {
		t.Fatal("expected EOF error")
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("attempts = %d, want 1", got)
	}
}

func TestClientDoesNotRetryPermanentResponseErrors(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		_, _ = w.Write([]byte(`{"broken":`))
	}))
	defer server.Close()

	client := NewClient(ClientOptions{Endpoint: server.URL})
	var response map[string]any
	_, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/"}, &response)
	if err == nil {
		t.Fatal("expected decode error")
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("attempts = %d, want 1", got)
	}
}
```

- [ ] **Step 6: 运行门禁测试并确认现有循环会错误地重试**

Run:

```bash
go test ./internal/api -run '^TestClientDoesNotRetry(EOFForWriteMethods|PermanentResponseErrors)$' -count=1
```

Expected: FAIL，两个测试的实际请求次数均为 3，而期望为 1。

- [ ] **Step 7: 只允许安全方法和临时错误进入下一次尝试**

在 `internal/api/client.go` imports 中加入 `"net"`，新增：

```go
func shouldRetry(method string, err error) bool {
	if method != http.MethodGet && method != http.MethodHead {
		return false
	}
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.Retryable()
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var networkErr net.Error
	return errors.As(err, &networkErr)
}
```

在 `Client.Do` 中用统一判定替换现有 API 错误特判：

```go
lastErr = err
if !shouldRetry(request.Method, err) {
	return meta, err
}
```

- [ ] **Step 8: 格式化并运行 API 包测试**

Run:

```bash
gofmt -w internal/api/client.go internal/api/client_test.go
go test ./internal/api -count=1
```

Expected: PASS。

- [ ] **Step 9: 运行完整验证**

Run:

```bash
./scripts/test.sh
./scripts/vet.sh
git diff --check
```

Expected: 全部退出码为 0。

- [ ] **Step 10: 提交实现**

```bash
git add internal/api/client.go internal/api/client_test.go docs/superpowers/plans/2026-07-30-basic-request-retry.md
git commit -m "fix: retry transient read request failures"
```
