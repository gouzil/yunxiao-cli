package yunxiao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gouzi/yunxiao-cli/internal/api"
)

func TestCodeRequestsUseOrganizationRepositoryPaths(t *testing.T) {
	tests := []struct {
		name     string
		wantPath string
		wantRef  string
		call     func(context.Context, ClientServices) error
		response any
	}{
		{
			name:     "list branches",
			wantPath: "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/branches",
			call: func(ctx context.Context, service ClientServices) error {
				_, err := service.ListBranches(ctx, ListBranchesRequest{Organization: "org-1", RepositoryID: "repo-1"})
				return err
			},
			response: map[string]any{"data": []map[string]any{{"name": "master"}}},
		},
		{
			name:     "get branch",
			wantPath: "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/branches/master",
			call: func(ctx context.Context, service ClientServices) error {
				_, err := service.GetBranch(ctx, GetBranchRequest{Organization: "org-1", RepositoryID: "repo-1", Branch: "master"})
				return err
			},
			response: map[string]any{"name": "master"},
		},
		{
			name:     "list commits",
			wantPath: "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/commits",
			wantRef:  "master",
			call: func(ctx context.Context, service ClientServices) error {
				_, err := service.ListCommits(ctx, ListCommitsRequest{Organization: "org-1", RepositoryID: "repo-1", Branch: "master"})
				return err
			},
			response: map[string]any{"data": []map[string]any{{"sha": "abc123"}}},
		},
		{
			name:     "get commit",
			wantPath: "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/commits/abc123",
			call: func(ctx context.Context, service ClientServices) error {
				_, err := service.GetCommit(ctx, GetCommitRequest{Organization: "org-1", RepositoryID: "repo-1", SHA: "abc123"})
				return err
			},
			response: map[string]any{"sha": "abc123"},
		},
		{
			name:     "get file",
			wantPath: "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/files/README.md",
			call: func(ctx context.Context, service ClientServices) error {
				_, err := service.GetFile(ctx, GetFileRequest{Organization: "org-1", RepositoryID: "repo-1", Path: "README.md"})
				return err
			},
			response: map[string]any{"path": "README.md", "content": "hello"},
		},
		{
			name:     "list files",
			wantPath: "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/files/tree",
			call: func(ctx context.Context, service ClientServices) error {
				_, err := service.ListFiles(ctx, ListFilesRequest{Organization: "org-1", RepositoryID: "repo-1"})
				return err
			},
			response: map[string]any{"data": []map[string]any{{"type": "file", "path": "README.md"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantPath {
					t.Fatalf("path = %q, want %q", r.URL.Path, tt.wantPath)
				}
				if tt.wantRef != "" && r.URL.Query().Get("refName") != tt.wantRef {
					t.Fatalf("refName = %q, want %q", r.URL.Query().Get("refName"), tt.wantRef)
				}
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
			if err := tt.call(context.Background(), service); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
		})
	}
}

func TestCodeResponsesDecodeCurrentCodeupShapes(t *testing.T) {
	tests := []struct {
		name     string
		response any
		call     func(context.Context, ClientServices) (any, error)
		assert   func(*testing.T, any)
	}{
		{
			name: "list branches decodes bare array and nested commit",
			response: []map[string]any{{
				"name":          "master",
				"defaultBranch": true,
				"protected":     false,
				"commit": map[string]any{
					"id":            "c862c200d139d9e19469bfa769f237d51f06daed",
					"authorName":    "gouziya",
					"committedDate": "2026-07-05T14:00:02+08:00",
				},
			}},
			call: func(ctx context.Context, service ClientServices) (any, error) {
				return service.ListBranches(ctx, ListBranchesRequest{Organization: "org-1", RepositoryID: "repo-1"})
			},
			assert: func(t *testing.T, value any) {
				branch := value.(BranchListResult).Branches[0]
				if branch.Name != "master" || !branch.Default || branch.Protected == nil || *branch.Protected || branch.CommitSHA == "" || branch.Author != "gouziya" {
					t.Fatalf("branch = %#v", branch)
				}
			},
		},
		{
			name: "get branch decodes direct object",
			response: map[string]any{
				"name":          "master",
				"defaultBranch": true,
				"commit":        map[string]any{"id": "c862c200d139d9e19469bfa769f237d51f06daed"},
			},
			call: func(ctx context.Context, service ClientServices) (any, error) {
				return service.GetBranch(ctx, GetBranchRequest{Organization: "org-1", RepositoryID: "repo-1", Branch: "master"})
			},
			assert: func(t *testing.T, value any) {
				branch := value.(BranchResult).Branch
				if branch.Name != "master" || !branch.Default || branch.CommitSHA == "" {
					t.Fatalf("branch = %#v", branch)
				}
			},
		},
		{
			name: "list commits decodes bare array",
			response: []map[string]any{{
				"id":            "c862c200d139d9e19469bfa769f237d51f06daed",
				"shortId":       "c862c200",
				"title":         "新建 README.md",
				"message":       "新建 README.md",
				"authorName":    "gouziya",
				"authorEmail":   "accounts@example.com",
				"committedDate": "2026-07-05T14:00:02+08:00",
			}},
			call: func(ctx context.Context, service ClientServices) (any, error) {
				return service.ListCommits(ctx, ListCommitsRequest{Organization: "org-1", RepositoryID: "repo-1", Branch: "master"})
			},
			assert: func(t *testing.T, value any) {
				commit := value.(CommitListResult).Commits[0]
				if commit.SHA == "" || commit.ShortSHA != "c862c200" || commit.AuthorMail != "accounts@example.com" || commit.Date == "" {
					t.Fatalf("commit = %#v", commit)
				}
			},
		},
		{
			name: "get file decodes direct object",
			response: map[string]any{
				"blobId":   "cf354e9078a1de3f1aa9b5182a12439551d5b051",
				"content":  "IyMg5rWL6K+V5LuT5bqT",
				"encoding": "base64",
				"filePath": "README.md",
				"ref":      "master",
				"size":     "15",
			},
			call: func(ctx context.Context, service ClientServices) (any, error) {
				return service.GetFile(ctx, GetFileRequest{Organization: "org-1", RepositoryID: "repo-1", Path: "README.md", Ref: "master"})
			},
			assert: func(t *testing.T, value any) {
				file := value.(FileResult).File
				if file.Path != "README.md" || file.Size != 15 || file.BlobID == "" || file.Encoding != "base64" {
					t.Fatalf("file = %#v", file)
				}
			},
		},
		{
			name: "list files decodes bare array from files tree",
			response: []map[string]any{{
				"id":   "cf354e9078a1de3f1aa9b5182a12439551d5b051",
				"name": "README.md",
				"path": "README.md",
				"type": "blob",
			}},
			call: func(ctx context.Context, service ClientServices) (any, error) {
				return service.ListFiles(ctx, ListFilesRequest{Organization: "org-1", RepositoryID: "repo-1", Ref: "master"})
			},
			assert: func(t *testing.T, value any) {
				file := value.(FileTreeResult).Files[0]
				if file.Path != "README.md" || file.Type != "blob" || file.BlobID == "" {
					t.Fatalf("file = %#v", file)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
			value, err := tt.call(context.Background(), service)
			if err != nil {
				t.Fatalf("call returned error: %v", err)
			}
			tt.assert(t, value)
		})
	}
}

func TestSSHKeyRequestsUseOrganizationPathsAndCurrentFields(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		wantPath string
		call     func(context.Context, ClientServices) error
		response any
	}{
		{
			name:     "list",
			method:   http.MethodGet,
			wantPath: "/oapi/v1/codeup/organizations/org-1/keys",
			call: func(ctx context.Context, service ClientServices) error {
				result, err := service.ListSSHKeys(ctx, ListSSHKeysRequest{Organization: "org-1"})
				if err != nil {
					return err
				}
				key := result.Keys[0]
				if key.ID != "919370" || key.Fingerprint != "5c:88" || key.PublicKey != "ssh-rsa AAA" {
					t.Fatalf("key = %#v", key)
				}
				return nil
			},
			response: []map[string]any{{"id": 919370, "title": "a100 vm", "fingerPrint": "5c:88", "key": "ssh-rsa AAA"}},
		},
		{
			name:     "create",
			method:   http.MethodPost,
			wantPath: "/oapi/v1/codeup/organizations/org-1/keys",
			call: func(ctx context.Context, service ClientServices) error {
				result, err := service.CreateSSHKey(ctx, CreateSSHKeyRequest{Organization: "org-1", Title: "test", PublicKey: "ssh-ed25519 AAA"})
				if err != nil {
					return err
				}
				if result.Key.ID != "1" || result.Key.Fingerprint != "fp" {
					t.Fatalf("key = %#v", result.Key)
				}
				return nil
			},
			response: map[string]any{"id": 1, "title": "test", "fingerPrint": "fp", "key": "ssh-ed25519 AAA"},
		},
		{
			name:     "delete",
			method:   http.MethodDelete,
			wantPath: "/oapi/v1/codeup/organizations/org-1/keys/1",
			call: func(ctx context.Context, service ClientServices) error {
				_, err := service.DeleteSSHKey(ctx, DeleteSSHKeyRequest{Organization: "org-1", ID: "1"})
				return err
			},
			response: map[string]any{"id": 1, "title": "test", "fingerPrint": "fp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.method {
					t.Fatalf("method = %q, want %q", r.Method, tt.method)
				}
				if r.URL.Path != tt.wantPath {
					t.Fatalf("path = %q, want %q", r.URL.Path, tt.wantPath)
				}
				if tt.name == "create" {
					body := map[string]any{}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatalf("decode request body: %v", err)
					}
					if body["key"] != "ssh-ed25519 AAA" || body["publicKey"] != nil || body["keyScope"] != "ALL" {
						t.Fatalf("body = %#v", body)
					}
				}
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
			if err := tt.call(context.Background(), service); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
		})
	}
}

func TestListSSHKeysListsAllOrganizationsWhenOrganizationMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/platform/organizations":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": "org-1", "name": "Org One"},
				{"id": "org-2", "name": "Org Two"},
			})
		case "/oapi/v1/codeup/organizations/org-1/keys":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "title": "shared", "fingerPrint": "fp-1"}})
		case "/oapi/v1/codeup/organizations/org-2/keys":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 1, "title": "shared", "fingerPrint": "fp-1"},
				{"id": 2, "title": "other", "fingerPrint": "fp-2"},
			})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListSSHKeys(context.Background(), ListSSHKeysRequest{})
	if err != nil {
		t.Fatalf("ListSSHKeys returned error: %v", err)
	}
	if len(result.Keys) != 2 || result.Keys[0].ID != "1" || result.Keys[1].ID != "2" {
		t.Fatalf("keys = %#v", result.Keys)
	}
}
