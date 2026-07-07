package yunxiao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gouzil/yunxiao-cli/internal/api"
)

func TestGetUserByTokenUsesDocumentedPlatformUserEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oapi/v1/platform/user" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("x-yunxiao-token"); got != "secret-token" {
			t.Fatalf("token header = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":               "user-1",
			"name":             "示例用户名",
			"nickName":         "示例昵称",
			"username":         "demo_username",
			"email":            "test@example.com",
			"lastOrganization": "org-1",
		})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.GetUserByToken(context.Background(), GetUserByTokenRequest{Token: "secret-token"})
	if err != nil {
		t.Fatalf("GetUserByToken returned error: %v", err)
	}
	if result.User.ID != "user-1" {
		t.Fatalf("user ID = %q", result.User.ID)
	}
	if result.User.DisplayName != "示例昵称" {
		t.Fatalf("display name = %q", result.User.DisplayName)
	}
	if result.User.Username != "demo_username" {
		t.Fatalf("username = %q", result.User.Username)
	}
	if result.User.Organization != "org-1" {
		t.Fatalf("organization = %q", result.User.Organization)
	}
}

func TestTokenVerifierUsesRequestedEndpoint(t *testing.T) {
	wrongServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "wrong endpoint", http.StatusTeapot)
	}))
	defer wrongServer.Close()
	rightServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oapi/v1/platform/user" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":       "user-2",
			"nickName": "Endpoint User",
		})
	}))
	defer rightServer.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: wrongServer.URL}))
	verifier := NewTokenVerifier(service)
	user, _, err := verifier.VerifyToken(context.Background(), rightServer.URL, "secret-token")
	if err != nil {
		t.Fatalf("VerifyToken returned error: %v", err)
	}
	if user.ID != "user-2" {
		t.Fatalf("user = %#v", user)
	}
}
