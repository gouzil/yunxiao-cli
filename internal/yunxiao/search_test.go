package yunxiao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gouzil/yunxiao-cli/internal/api"
)

func TestSearchUsesOrganizationAwareRepositoryAPIs(t *testing.T) {
	tests := []struct {
		name     string
		wantPath string
		call     func(context.Context, ClientServices) error
		response any
	}{
		{
			name:     "code",
			wantPath: "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/files/tree",
			call: func(ctx context.Context, service ClientServices) error {
				_, err := service.SearchCode(ctx, SearchCodeRequest{Organization: "org-1", RepositoryID: "repo-1", Ref: "master", Query: "README"})
				return err
			},
			response: []map[string]any{{"path": "README.md", "type": "blob"}},
		},
		{
			name:     "commit",
			wantPath: "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/commits",
			call: func(ctx context.Context, service ClientServices) error {
				_, err := service.SearchCommits(ctx, SearchCommitsRequest{Organization: "org-1", RepositoryID: "repo-1", Ref: "master", Query: "README"})
				return err
			},
			response: []map[string]any{{"id": "abc123", "title": "README"}},
		},
		{
			name:     "merge requests",
			wantPath: "/oapi/v1/codeup/organizations/org-1/changeRequests",
			call: func(ctx context.Context, service ClientServices) error {
				_, err := service.SearchMergeRequests(ctx, SearchMergeRequestsRequest{Organization: "org-1", RepositoryID: "repo-1", Query: "README"})
				return err
			},
			response: []map[string]any{{"localId": 1, "projectId": "repo-1", "title": "README"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantPath {
					t.Fatalf("path = %q, want %q", r.URL.Path, tt.wantPath)
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
