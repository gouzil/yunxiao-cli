package yunxiao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gouzi/yunxiao-cli/internal/api"
)

func TestListPipelinesUsesOrganizationPathAndDecodesArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oapi/v1/flow/organizations/org-1/pipelines" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("projectId") != "project-1" {
			t.Fatalf("query = %q", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"pipelineId":   123,
			"pipelineName": "build",
			"createTime":   float64(1781502690000),
		}})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListPipelines(context.Background(), ListPipelinesRequest{Organization: "org-1", ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("ListPipelines returned error: %v", err)
	}
	if len(result.Pipelines) != 1 || result.Pipelines[0].ID != "123" || result.Pipelines[0].Name != "build" || result.Pipelines[0].UpdatedAt == "" {
		t.Fatalf("pipelines = %#v", result.Pipelines)
	}
}

func TestListPipelinesListsAllOrganizationsWhenOrganizationMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/platform/organizations":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": "org-1", "name": "Org One"},
				{"id": "org-2", "name": "Org Two"},
			})
		case "/oapi/v1/flow/organizations/org-1/pipelines":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 101, "name": "build-a"}})
		case "/oapi/v1/flow/organizations/org-2/pipelines":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 202, "name": "build-b"}})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListPipelines(context.Background(), ListPipelinesRequest{})
	if err != nil {
		t.Fatalf("ListPipelines returned error: %v", err)
	}
	if len(result.Pipelines) != 2 || result.Pipelines[0].ID != "101" || result.Pipelines[1].ID != "202" {
		t.Fatalf("pipelines = %#v", result.Pipelines)
	}
}

func TestGetPipelineTriesOrganizationsAndDecodesBareObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/platform/organizations":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": "org-1", "name": "Org One"},
				{"id": "org-2", "name": "Org Two"},
			})
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe-1":
			http.Error(w, `{"errorCode":"NotFound","errorMessage":"Not Found"}`, http.StatusNotFound)
		case "/oapi/v1/flow/organizations/org-2/pipelines/pipe-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"pipelineId":   123,
				"pipelineName": "build",
				"createTime":   float64(1781502690000),
			})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.GetPipeline(context.Background(), GetPipelineRequest{PipelineID: "pipe-1"})
	if err != nil {
		t.Fatalf("GetPipeline returned error: %v", err)
	}
	if result.Pipeline.ID != "123" || result.Pipeline.Name != "build" || result.Pipeline.UpdatedAt == "" {
		t.Fatalf("pipeline = %#v", result.Pipeline)
	}
}

func TestListRunsUsesOrganizationPipelinePathAndDecodesArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oapi/v1/flow/organizations/org-1/pipelines/pipe-1/runs" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"pipelineRunId": 456,
			"pipelineId":    123,
			"startTime":     float64(1783099538000),
			"endTime":       float64(1783099560000),
			"status":        "SUCCESS",
			"triggerMode":   3,
		}})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListRuns(context.Background(), ListRunsRequest{Organization: "org-1", PipelineID: "pipe-1"})
	if err != nil {
		t.Fatalf("ListRuns returned error: %v", err)
	}
	if len(result.Runs) != 1 || result.Runs[0].ID != "456" || result.Runs[0].PipelineID != "123" || result.Runs[0].StartedAt == "" || result.Runs[0].TriggerMode != "3" {
		t.Fatalf("runs = %#v", result.Runs)
	}
}

func TestListRunsTriesOrganizationsWhenOrganizationMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/platform/organizations":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": "org-1", "name": "Org One"},
				{"id": "org-2", "name": "Org Two"},
			})
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe-1/runs":
			http.Error(w, `{"errorCode":"NotFound","errorMessage":"Not Found"}`, http.StatusNotFound)
		case "/oapi/v1/flow/organizations/org-2/pipelines/pipe-1/runs":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"pipelineRunId": 456, "pipelineId": "pipe-1", "status": "SUCCESS"}})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListRuns(context.Background(), ListRunsRequest{PipelineID: "pipe-1"})
	if err != nil {
		t.Fatalf("ListRuns returned error: %v", err)
	}
	if len(result.Runs) != 1 || result.Runs[0].ID != "456" {
		t.Fatalf("runs = %#v", result.Runs)
	}
}

func TestGetRunFindsPipelineWhenContextMissingAndDecodesCurrentShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/platform/organizations":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "org-1", "name": "Org One"}})
		case "/oapi/v1/flow/organizations/org-1/pipelines":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"pipelineId": 111, "pipelineName": "other"},
				{"pipelineId": 123, "pipelineName": "build"},
			})
		case "/oapi/v1/flow/organizations/org-1/pipelines/111/runs/13":
			http.Error(w, `{"errorCode":"NotFound","errorMessage":"Not Found"}`, http.StatusNotFound)
		case "/oapi/v1/flow/organizations/org-1/pipelines/123/runs/13":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"pipelineRunId": 13,
				"pipelineId":    123,
				"status":        "SUCCESS",
				"triggerMode":   3,
				"createTime":    float64(1783099538000),
				"updateTime":    float64(1783099560000),
				"stages": []map[string]any{{
					"stageInfo": map[string]any{
						"jobs": []map[string]any{{"id": 462665928, "name": "主机部署", "status": "SUCCESS"}},
					},
				}},
			})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.GetRun(context.Background(), GetRunRequest{RunID: "13"})
	if err != nil {
		t.Fatalf("GetRun returned error: %v", err)
	}
	if result.Run.ID != "13" || result.Run.PipelineID != "123" || result.Run.StartedAt == "" || result.Run.FinishedAt == "" || result.Run.TriggerMode != "3" {
		t.Fatalf("run = %#v", result.Run)
	}
	if len(result.Run.Jobs) != 1 || result.Run.Jobs[0].ID != "462665928" {
		t.Fatalf("jobs = %#v", result.Run.Jobs)
	}
}

func TestGetRunLogUsesResolvedOrganizationPipelinePath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oapi/v1/flow/organizations/org-1/pipelines/123/runs/13/job/job-1/log" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"content": "ok"})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.GetRunLog(context.Background(), GetRunLogRequest{Organization: "org-1", PipelineID: "123", RunID: "13", JobID: "job-1"})
	if err != nil {
		t.Fatalf("GetRunLog returned error: %v", err)
	}
	if result.Lines != "ok" {
		t.Fatalf("lines = %q", result.Lines)
	}
}

func TestGetRunLogUsesFirstRunJobWhenJobMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/flow/organizations/org-1/pipelines/123/runs/13":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"pipelineRunId": 13,
				"pipelineId":    123,
				"status":        "SUCCESS",
				"stages": []map[string]any{{
					"stageInfo": map[string]any{
						"jobs": []map[string]any{{"id": 462665928, "name": "主机部署"}},
					},
				}},
			})
		case "/oapi/v1/flow/organizations/org-1/pipelines/123/runs/13/job/462665928/log":
			_ = json.NewEncoder(w).Encode(map[string]any{"content": "auto"})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.GetRunLog(context.Background(), GetRunLogRequest{Organization: "org-1", PipelineID: "123", RunID: "13"})
	if err != nil {
		t.Fatalf("GetRunLog returned error: %v", err)
	}
	if result.JobID != "462665928" || result.Lines != "auto" {
		t.Fatalf("result = %#v", result)
	}
}

func TestWatchRunTreatsUppercaseStatusAsTerminal(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/oapi/v1/flow/organizations/org-1/pipelines/123/runs/13" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"pipelineRunId": 13, "pipelineId": 123, "status": "SUCCESS"})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.WatchRun(context.Background(), WatchRunRequest{
		Organization: "org-1",
		PipelineID:   "123",
		RunID:        "13",
		Interval:     time.Millisecond,
		Timeout:      time.Second,
	})
	if err != nil {
		t.Fatalf("WatchRun returned error: %v", err)
	}
	if result.Run.Status != "SUCCESS" || calls != 1 {
		t.Fatalf("run = %#v calls = %d", result.Run, calls)
	}
}
