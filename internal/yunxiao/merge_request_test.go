package yunxiao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gouzil/yunxiao-cli/internal/api"
)

func TestListMergeRequestsUsesOrganizationPathAndFiltersRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/changeRequests" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"localId": 1, "mrBizId": "mr-1", "projectId": 7123450, "title": "test repo MR", "state": "UNDER_REVIEW", "author": map[string]any{"name": "gouziya"}, "detailUrl": "https://example.test/mr/1"},
			{"localId": 2, "mrBizId": "mr-2", "projectId": 7123444, "title": "other repo MR"},
		})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListMergeRequests(context.Background(), ListMergeRequestsRequest{Organization: "org-1", RepositoryID: "7123450"})
	if err != nil {
		t.Fatalf("ListMergeRequests returned error: %v", err)
	}
	if len(result.MergeRequests) != 1 {
		t.Fatalf("merge requests = %#v", result.MergeRequests)
	}
	mr := result.MergeRequests[0]
	if mr.IID != "1" || mr.ID != "mr-1" || mr.Author != "gouziya" || mr.WebURL == "" {
		t.Fatalf("merge request = %#v", mr)
	}
}

func TestMergeRequestDetailUsesOrganizationRepositoryPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/changeRequests/1" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"localId":                1,
			"mrBizId":                "mr-1",
			"title":                  "test repo MR",
			"allRequirementsPass":    true,
			"conflictCheckStatus":    "NO_CONFLICT",
			"totalCommentCount":      6,
			"unResolvedCommentCount": 6,
			"lastPipeline":           map[string]any{"status": "success"},
			"webUrl":                 "https://codeup.aliyun.com/org-1/test-repo",
			"detailUrl":              "https://codeup.aliyun.com/org-1/test-repo/change/1",
		})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.GetMergeRequest(context.Background(), GetMergeRequestRequest{Organization: "org-1", RepositoryID: "repo-1", MergeRequestID: "1"})
	if err != nil {
		t.Fatalf("GetMergeRequest returned error: %v", err)
	}
	if result.MergeRequest.IID != "1" || result.MergeRequest.ID != "mr-1" {
		t.Fatalf("merge request = %#v", result.MergeRequest)
	}
	if result.MergeRequest.Mergeable == nil || !*result.MergeRequest.Mergeable || result.MergeRequest.HasConflicts == nil || *result.MergeRequest.HasConflicts || result.MergeRequest.ReviewStatus != "6 unresolved comments" || result.MergeRequest.PipelineStatus != "success" {
		t.Fatalf("merge request = %#v", result.MergeRequest)
	}
	if result.MergeRequest.WebURL != "https://codeup.aliyun.com/org-1/test-repo/change/1" {
		t.Fatalf("web URL = %q", result.MergeRequest.WebURL)
	}
}

func TestCreateMergeRequestUsesCurrentPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories/7123450/changeRequests" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["organization"] != nil || body["repositoryId"] != nil {
			t.Fatalf("body includes internal fields: %#v", body)
		}
		if body["sourceProjectId"] != float64(7123450) || body["targetProjectId"] != float64(7123450) {
			t.Fatalf("body = %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"bizId":        "mr-biz-1",
			"localId":      1,
			"status":       "UNDER_REVIEW",
			"title":        "test repo MR",
			"sourceBranch": "feature",
			"targetBranch": "master",
			"webUrl":       "https://codeup.aliyun.com/org-1/test-repo",
			"detailUrl":    "https://codeup.aliyun.com/org-1/test-repo/change/1",
			"updateTime":   "2026-07-05T21:02:45+08:00",
		})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.CreateMergeRequest(context.Background(), CreateMergeRequestRequest{
		Organization: "org-1",
		RepositoryID: "7123450",
		SourceBranch: "feature",
		TargetBranch: "master",
		Title:        "test repo MR",
	})
	if err != nil {
		t.Fatalf("CreateMergeRequest returned error: %v", err)
	}
	mr := result.MergeRequest
	if mr.ID != "mr-biz-1" || mr.IID != "1" || mr.State != "UNDER_REVIEW" || mr.UpdatedAt == "" {
		t.Fatalf("merge request = %#v", mr)
	}
	if mr.WebURL != "https://codeup.aliyun.com/org-1/test-repo/change/1" {
		t.Fatalf("web URL = %q", mr.WebURL)
	}
}

func TestMergeRequestFilesUsePatchSets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/codeup/organizations/org-1/repositories/7123450/changeRequests/2/diffs/patches":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"patchSetBizId": "source-patch", "relatedMergeItemType": "MERGE_SOURCE"},
				{"patchSetBizId": "target-patch", "relatedMergeItemType": "MERGE_TARGET"},
			})
		case "/oapi/v1/codeup/organizations/org-1/repositories/7123450/changeRequests/2/diffs/changeTree":
			if r.URL.Query().Get("fromPatchSetId") != "target-patch" || r.URL.Query().Get("toPatchSetId") != "source-patch" {
				t.Fatalf("query = %q", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"changedTreeItems": []map[string]any{{
				"addLines": 3,
				"delLines": 0,
				"newFile":  true,
				"newPath":  "codex-live-test.txt",
			}}})
		default:
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListMergeRequestFiles(context.Background(), ListMergeRequestFilesRequest{Organization: "org-1", RepositoryID: "7123450", MergeRequestID: "2"})
	if err != nil {
		t.Fatalf("ListMergeRequestFiles returned error: %v", err)
	}
	if len(result.Files) != 1 || result.Files[0].Path != "codex-live-test.txt" || result.Files[0].Status != "added" || result.Files[0].Additions != 3 {
		t.Fatalf("files = %#v", result.Files)
	}
}

func TestCommentMergeRequestUsesCurrentGlobalPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/codeup/organizations/org-1/repositories/7123450/changeRequests/3/diffs/patches":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"patchSetBizId": "source-patch", "relatedMergeItemType": "MERGE_SOURCE"},
				{"patchSetBizId": "target-patch", "relatedMergeItemType": "MERGE_TARGET"},
			})
		case "/oapi/v1/codeup/organizations/org-1/repositories/7123450/changeRequests/3/comments":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["comment_type"] != "GLOBAL_COMMENT" || body["content"] != "direct" || body["patchset_biz_id"] != "source-patch" {
				t.Fatalf("body = %#v", body)
			}
			if body["body"] != nil || body["organization"] != nil {
				t.Fatalf("body includes internal fields: %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"comment_biz_id": "comment-1"})
		default:
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.CommentMergeRequest(context.Background(), CommentMergeRequestRequest{Organization: "org-1", RepositoryID: "7123450", MergeRequestID: "3", Body: "direct"})
	if err != nil {
		t.Fatalf("CommentMergeRequest returned error: %v", err)
	}
	if result.ID != "comment-1" {
		t.Fatalf("comment id = %q", result.ID)
	}
}

func TestCommentMergeRequestUsesCurrentInlinePayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/codeup/organizations/org-1/repositories/7123450/changeRequests/3/diffs/patches":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"patchSetBizId": "source-patch", "relatedMergeItemType": "MERGE_SOURCE"},
				{"patchSetBizId": "target-patch", "relatedMergeItemType": "MERGE_TARGET"},
			})
		case "/oapi/v1/codeup/organizations/org-1/repositories/7123450/changeRequests/3/comments":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["comment_type"] != "INLINE_COMMENT" || body["file_path"] != "main.go" || body["line_number"] != float64(3) {
				t.Fatalf("body = %#v", body)
			}
			if body["from_patchset_biz_id"] != "target-patch" || body["to_patchset_biz_id"] != "source-patch" || body["patchset_biz_id"] != "source-patch" {
				t.Fatalf("body = %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"comment_biz_id": "comment-inline"})
		default:
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.CommentMergeRequest(context.Background(), CommentMergeRequestRequest{Organization: "org-1", RepositoryID: "7123450", MergeRequestID: "3", Body: "inline", FilePath: "main.go", LineNumber: 3})
	if err != nil {
		t.Fatalf("CommentMergeRequest returned error: %v", err)
	}
	if result.ID != "comment-inline" {
		t.Fatalf("comment id = %q", result.ID)
	}
}

func TestResolveMergeRequestCommentUsesCurrentPayload(t *testing.T) {
	for _, resolved := range []bool{true, false} {
		t.Run(fmt.Sprint(resolved), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut || r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories/7123450/changeRequests/3/comments/comment-inline" {
					t.Fatalf("%s %s", r.Method, r.URL.Path)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body["resolved"] != resolved {
					t.Fatalf("body = %#v", body)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"result": true})
			}))
			defer server.Close()

			service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
			result, err := service.ResolveMergeRequestComment(context.Background(), ResolveMergeRequestCommentRequest{Organization: "org-1", RepositoryID: "7123450", MergeRequestID: "3", CommentID: "comment-inline", Resolved: resolved})
			if err != nil {
				t.Fatalf("ResolveMergeRequestComment returned error: %v", err)
			}
			if result.ID != "comment-inline" {
				t.Fatalf("comment id = %q", result.ID)
			}
		})
	}
}

func TestMergeMergeRequestSendsMergeType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories/7123450/changeRequests/2/merge" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["mergeType"] != "ff-only" {
			t.Fatalf("body = %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"localId": 2, "status": "MERGED", "mergedRevision": "abc123"})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.MergeMergeRequest(context.Background(), MergeRequestActionRequest{Organization: "org-1", RepositoryID: "7123450", MergeRequestID: "2"})
	if err != nil {
		t.Fatalf("MergeMergeRequest returned error: %v", err)
	}
	if result.MergeRequest.IID != "2" || result.MergeCommit != "abc123" {
		t.Fatalf("result = %#v", result)
	}
}
