package yunxiao

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gouzi/yunxiao-cli/internal/api"
)

type APIStatus string

const (
	APIStatusCurrent        APIStatus = "current"
	APIStatusPendingConfirm APIStatus = "pending_confirmation"
	APIStatusLegacy         APIStatus = "legacy"
	APIStatusLocal          APIStatus = "local"
)

type Mapping struct {
	Command string    `json:"command"`
	APIName string    `json:"apiName"`
	Method  string    `json:"method"`
	Path    string    `json:"path"`
	DocURL  string    `json:"docUrl"`
	Status  APIStatus `json:"status"`
	Notes   string    `json:"notes"`
}

type ListOptions struct {
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
	Query   string `json:"query,omitempty"`
}

type User struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Organization string `json:"organization"`
}

type TimeValue struct {
	Value string `json:"value"`
}

func FormatTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z0700", "2006-01-02 15:04:05"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Local().Format("2006-01-02 15:04")
		}
	}
	return value
}

func Unknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}

func YesNoUnknown(value *bool) string {
	if value == nil {
		return "unknown"
	}
	if *value {
		return "true"
	}
	return "false"
}

func BoolPtr(value bool) *bool {
	return &value
}

type ServiceSet struct {
	Auth     AuthService
	Repo     RepositoryService
	Branch   BranchService
	Commit   CommitService
	File     FileService
	SSHKey   SSHKeyService
	MR       MergeRequestService
	Pipeline PipelineService
	Run      RunService
	Project  ProjectService
	WorkItem WorkItemService
	Search   SearchService
	RawAPI   RawAPIService
}

type RawAPIService interface {
	Request(ctx context.Context, request RawAPIRequest) (RawAPIResult, error)
}

type RawAPIRequest struct {
	Method  string       `json:"method"`
	Path    string       `json:"path"`
	Headers []api.Header `json:"headers,omitempty"`
	Body    []byte       `json:"-"`
	Verbose bool         `json:"verbose"`
}

type RawAPIResult struct {
	Method    string           `json:"method"`
	URL       string           `json:"url"`
	Status    int              `json:"status"`
	RequestID string           `json:"requestId"`
	Body      string           `json:"body"`
	Meta      api.ResponseMeta `json:"meta"`
}

type ClientServices struct {
	client *api.Client
}

func NewClientServices(client *api.Client) ClientServices {
	return ClientServices{client: client}
}

func (s ClientServices) Request(ctx context.Context, request RawAPIRequest) (RawAPIResult, error) {
	method := strings.ToUpper(strings.TrimSpace(request.Method))
	if method == "" {
		method = http.MethodGet
	}
	raw := api.RawBody{}
	meta, err := s.client.Do(ctx, api.Request{
		Method:  method,
		Path:    request.Path,
		Headers: request.Headers,
		Body:    request.Body,
	}, &raw)
	if err != nil {
		return RawAPIResult{}, err
	}
	return RawAPIResult{
		Method:    method,
		URL:       request.Path,
		Status:    meta.StatusCode,
		RequestID: meta.RequestID,
		Body:      string(raw.Body),
		Meta:      meta,
	}, nil
}

func queryParams(options ListOptions) []api.QueryParam {
	params := make([]api.QueryParam, 0, 3)
	if options.Page > 0 {
		params = append(params, api.QueryParam{Name: "page", Value: strconv.Itoa(options.Page)})
	}
	if options.PerPage > 0 {
		params = append(params, api.QueryParam{Name: "perPage", Value: strconv.Itoa(options.PerPage)})
	}
	if strings.TrimSpace(options.Query) != "" {
		params = append(params, api.QueryParam{Name: "search", Value: options.Query})
	}
	return params
}

func addParam(params []api.QueryParam, name string, value string) []api.QueryParam {
	if strings.TrimSpace(value) == "" {
		return params
	}
	return append(params, api.QueryParam{Name: name, Value: value})
}

func pathf(format string, args ...any) string {
	escaped := make([]any, len(args))
	for index, arg := range args {
		escaped[index] = strings.Trim(urlPathEscape(fmt.Sprint(arg)), "/")
	}
	return fmt.Sprintf(format, escaped...)
}

func urlPathEscape(value string) string {
	value = strings.ReplaceAll(value, " ", "%20")
	value = strings.ReplaceAll(value, "#", "%23")
	value = strings.ReplaceAll(value, "?", "%3F")
	return value
}
