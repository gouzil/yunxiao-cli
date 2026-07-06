package yunxiao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gouzi/yunxiao-cli/internal/api"
)

func TestListRepositoriesUsesOrganizationPathAndDecodesArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("organizationId"); got != "" {
			t.Fatalf("organizationId query = %q", got)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id":                2813489,
			"name":              "demo-repo",
			"path":              "demo-repo",
			"pathWithNamespace": "org/demo-repo",
			"visibility":        "private",
			"updatedAt":         "2024-10-05T15:30:45Z",
			"webUrl":            "https://example.com/org/demo-repo",
		}})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListRepositories(context.Background(), ListRepositoriesRequest{Organization: "org-1"})
	if err != nil {
		t.Fatalf("ListRepositories returned error: %v", err)
	}
	if len(result.Repositories) != 1 {
		t.Fatalf("repositories = %#v", result.Repositories)
	}
	repo := result.Repositories[0]
	if repo.ID != "2813489" || repo.Path != "org/demo-repo" {
		t.Fatalf("repo = %#v", repo)
	}
}

func TestGetRepositoryUsesOrganizationPathAndDecodesDetailFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories/7123450" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                7123450,
			"name":              "test-repo",
			"path":              "test-repo",
			"pathWithNamespace": "org-1/test-repo",
			"defaultBranch":     "master",
			"visibility":        "internal",
			"sshUrlToRepo":      "git@codeup.aliyun.com:org-1/test-repo.git",
			"httpUrlToRepo":     "https://codeup.aliyun.com/org-1/test-repo.git",
			"namespace": map[string]any{
				"id":   2025069,
				"path": "org-1",
			},
		})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.GetRepository(context.Background(), GetRepositoryRequest{Organization: "org-1", RepositoryID: "7123450"})
	if err != nil {
		t.Fatalf("GetRepository returned error: %v", err)
	}
	repo := result.Repository
	if repo.DefaultBranch != "master" || repo.SSHURL == "" || repo.HTTPURL == "" {
		t.Fatalf("repo = %#v", repo)
	}
}

func TestListRepositoriesListsAllJoinedOrganizationsWithoutOrganization(t *testing.T) {
	paths := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/oapi/v1/platform/organizations":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": "org-1", "name": "Org One"},
				{"id": "org-2", "name": "Org Two"},
			})
		case "/oapi/v1/codeup/organizations/org-1/repositories":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "repo-1", "name": "repo-one"}})
		case "/oapi/v1/codeup/organizations/org-2/repositories":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "repo-2", "name": "repo-two"}})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListRepositories(context.Background(), ListRepositoriesRequest{})
	if err != nil {
		t.Fatalf("ListRepositories returned error: %v", err)
	}
	if len(result.Repositories) != 2 {
		t.Fatalf("repositories = %#v", result.Repositories)
	}
	if got := paths; len(got) != 3 || got[0] != "/oapi/v1/platform/organizations" {
		t.Fatalf("paths = %#v", got)
	}
}

func TestRepositoryWriteActionsUseOrganizationPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			if body["name"] != "demo" || body["path"] != "demo" || body["description"] != "desc" {
				t.Fatalf("create body = %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 123, "name": "demo"})
		case r.Method == http.MethodPut && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories/123":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			if body["name"] != "demo-updated" {
				t.Fatalf("update body = %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 123, "name": "demo-updated"})
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories/123/archive":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": true})
		case r.Method == http.MethodDelete && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories/123":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	created, err := service.CreateRepository(context.Background(), CreateRepositoryRequest{Organization: "org-1", Name: "demo", Description: "desc"})
	if err != nil {
		t.Fatalf("CreateRepository returned error: %v", err)
	}
	if created.Repository.ID != "123" {
		t.Fatalf("created = %#v", created.Repository)
	}

	updated, err := service.UpdateRepository(context.Background(), UpdateRepositoryRequest{Organization: "org-1", RepositoryID: "123", Name: "demo-updated"})
	if err != nil {
		t.Fatalf("UpdateRepository returned error: %v", err)
	}
	if updated.Repository.Name != "demo-updated" {
		t.Fatalf("updated = %#v", updated.Repository)
	}

	archived, err := service.ArchiveRepository(context.Background(), RepositoryActionRequest{Organization: "org-1", RepositoryID: "123"})
	if err != nil {
		t.Fatalf("ArchiveRepository returned error: %v", err)
	}
	if archived.Action != "archive" || !archived.Supported {
		t.Fatalf("archived = %#v", archived)
	}

	deleted, err := service.DeleteRepository(context.Background(), RepositoryActionRequest{Organization: "org-1", RepositoryID: "123"})
	if err != nil {
		t.Fatalf("DeleteRepository returned error: %v", err)
	}
	if deleted.Repository.ID != "123" {
		t.Fatalf("deleted = %#v", deleted)
	}
}
