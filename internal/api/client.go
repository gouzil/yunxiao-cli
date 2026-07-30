package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultEndpoint = "openapi-rdc.aliyuncs.com"
	tokenHeader     = "x-yunxiao-token"
	defaultRetryMax = 2
)

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type QueryParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PageRequest struct {
	Page    int `json:"page"`
	PerPage int `json:"perPage"`
}

type PageInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	PrevPage   int `json:"prevPage"`
	NextPage   int `json:"nextPage"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type Request struct {
	Method  string       `json:"method"`
	Path    string       `json:"path"`
	Query   []QueryParam `json:"query,omitempty"`
	Headers []Header     `json:"headers,omitempty"`
	Body    []byte       `json:"-"`
}

type ResponseMeta struct {
	StatusCode int      `json:"statusCode"`
	RequestID  string   `json:"requestId"`
	Page       PageInfo `json:"page"`
	Headers    []Header `json:"headers"`
}

type ClientOptions struct {
	Endpoint   string
	Token      string
	HTTPClient *http.Client
	Timeout    time.Duration
	RetryMax   int
	Debug      bool
	DebugOut   io.Writer
}

type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
	retryMax   int
	debug      bool
	debugOut   io.Writer
}

func NewClient(options ClientOptions) *Client {
	endpoint := strings.TrimSpace(options.Endpoint)
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	retryMax := options.RetryMax
	if retryMax == 0 {
		retryMax = defaultRetryMax
	}
	if retryMax < 0 {
		retryMax = 0
	}
	return &Client{
		endpoint:   endpoint,
		token:      strings.TrimSpace(options.Token),
		httpClient: httpClient,
		retryMax:   retryMax,
		debug:      options.Debug,
		debugOut:   options.DebugOut,
	}
}

func (c *Client) Endpoint() string {
	return c.endpoint
}

func (c *Client) WithToken(token string) *Client {
	next := *c
	next.token = strings.TrimSpace(token)
	return &next
}

func (c *Client) WithEndpoint(endpoint string) *Client {
	next := *c
	if strings.TrimSpace(endpoint) != "" {
		next.endpoint = strings.TrimSpace(endpoint)
	}
	return &next
}

func (c *Client) Do(ctx context.Context, request Request, out any) (ResponseMeta, error) {
	if request.Method == "" {
		request.Method = http.MethodGet
	}
	var lastErr error
	attempts := c.retryMax + 1
	for attempt := 0; attempt < attempts; attempt++ {
		meta, err := c.doOnce(ctx, request, out)
		if err == nil {
			return meta, nil
		}
		lastErr = err
		if !shouldRetry(request.Method, err) {
			return meta, err
		}
		if attempt+1 < attempts {
			select {
			case <-ctx.Done():
				return ResponseMeta{}, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 250 * time.Millisecond):
			}
		}
	}
	return ResponseMeta{}, lastErr
}

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

func (c *Client) doOnce(ctx context.Context, request Request, out any) (ResponseMeta, error) {
	endpoint, err := c.buildURL(request)
	if err != nil {
		return ResponseMeta{}, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, endpoint, bytes.NewReader(request.Body))
	if err != nil {
		return ResponseMeta{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	if c.token != "" {
		httpRequest.Header.Set(tokenHeader, c.token)
	}
	for _, header := range request.Headers {
		if strings.EqualFold(header.Name, tokenHeader) {
			continue
		}
		httpRequest.Header.Set(header.Name, header.Value)
	}
	if c.debug && c.debugOut != nil {
		fmt.Fprintf(c.debugOut, "> %s %s\n", httpRequest.Method, redactURL(httpRequest.URL.String()))
	}
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return ResponseMeta{}, err
	}
	defer response.Body.Close()
	meta := responseMeta(response)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return meta, err
	}
	if c.debug && c.debugOut != nil {
		fmt.Fprintf(c.debugOut, "< HTTP %d\n", response.StatusCode)
		if meta.RequestID != "" {
			fmt.Fprintf(c.debugOut, "< x-request-id: %s\n", meta.RequestID)
		}
		fmt.Fprintln(c.debugOut)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return meta, errorFromResponse(response.StatusCode, meta.RequestID, body)
	}
	if out == nil {
		return meta, nil
	}
	raw, ok := out.(*RawBody)
	if ok {
		raw.Body = append(raw.Body[:0], body...)
		raw.Meta = meta
		return meta, nil
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return meta, nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return meta, fmt.Errorf("decode response: %w", err)
	}
	return meta, nil
}

func (c *Client) buildURL(request Request) (string, error) {
	if strings.HasPrefix(request.Path, "http://") || strings.HasPrefix(request.Path, "https://") {
		parsed, err := url.Parse(request.Path)
		if err != nil {
			return "", err
		}
		appendQuery(parsed, request.Query)
		return parsed.String(), nil
	}
	base := c.endpoint
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	path := request.Path
	if path == "" {
		path = "/"
	}
	rawQuery := ""
	if splitPath, splitQuery, ok := strings.Cut(path, "?"); ok {
		path = splitPath
		rawQuery = splitQuery
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	if rawQuery != "" {
		parsed.RawQuery = rawQuery
	}
	appendQuery(parsed, request.Query)
	return parsed.String(), nil
}

func appendQuery(parsed *url.URL, params []QueryParam) {
	values := parsed.Query()
	for _, param := range params {
		if param.Name == "" || param.Value == "" {
			continue
		}
		values.Set(param.Name, param.Value)
	}
	parsed.RawQuery = values.Encode()
}

func responseMeta(response *http.Response) ResponseMeta {
	meta := ResponseMeta{
		StatusCode: response.StatusCode,
		RequestID:  firstHeader(response.Header, "x-request-id", "x-acs-request-id"),
		Headers:    headersFromHTTP(response.Header),
		Page: PageInfo{
			Page:       headerInt(response.Header, "x-page"),
			PerPage:    headerInt(response.Header, "x-per-page"),
			PrevPage:   headerInt(response.Header, "x-prev-page"),
			NextPage:   headerInt(response.Header, "x-next-page"),
			Total:      headerInt(response.Header, "x-total"),
			TotalPages: headerInt(response.Header, "x-total-pages"),
		},
	}
	return meta
}

func headersFromHTTP(headers http.Header) []Header {
	result := make([]Header, 0, len(headers))
	for name, values := range headers {
		result = append(result, Header{Name: name, Value: strings.Join(values, ", ")})
	}
	return result
}

func firstHeader(headers http.Header, names ...string) string {
	for _, name := range names {
		if value := headers.Get(name); value != "" {
			return value
		}
	}
	return ""
}

func headerInt(headers http.Header, name string) int {
	value := strings.TrimSpace(headers.Get(name))
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

type RawBody struct {
	Body []byte       `json:"-"`
	Meta ResponseMeta `json:"meta"`
}

type ErrorBody struct {
	RequestID string `json:"requestId"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

type Error struct {
	Summary    string `json:"summary"`
	RequestID  string `json:"requestId"`
	HTTPStatus int    `json:"httpStatus"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Summary
}

func (e *Error) Retryable() bool {
	return e.HTTPStatus == http.StatusTooManyRequests || e.HTTPStatus >= 500
}

func errorFromResponse(status int, requestID string, body []byte) error {
	payload := ErrorBody{}
	_ = json.Unmarshal(body, &payload)
	code := firstNonEmpty(payload.Code, payload.ErrorCode, http.StatusText(status))
	message := firstNonEmpty(payload.Message, payload.ErrorMsg, strings.TrimSpace(string(body)), http.StatusText(status))
	return &Error{
		Summary:    "Yunxiao API request failed",
		RequestID:  firstNonEmpty(payload.RequestID, requestID),
		HTTPStatus: status,
		Code:       code,
		Message:    message,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func redactURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	query := parsed.Query()
	for name := range query {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "credential") {
			query.Set(name, "REDACTED")
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func EncodeJSONBody(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	body, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}
	return body, nil
}
