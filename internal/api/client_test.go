package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestClientInjectsTokenAndDecodesResponseMeta(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(tokenHeader); got != "secret-token" {
			t.Fatalf("token header = %q", got)
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Fatalf("page query = %q", got)
		}
		w.Header().Set("x-acs-request-id", "req-123")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(struct {
			Name string `json:"name"`
		}{Name: "repo"})
	}))
	defer server.Close()

	client := NewClient(ClientOptions{Endpoint: server.URL, Token: "secret-token"})
	var response struct {
		Name string `json:"name"`
	}
	meta, err := client.Do(context.Background(), Request{
		Method: http.MethodGet,
		Path:   "/repositories",
		Query:  []QueryParam{{Name: "page", Value: "2"}},
	}, &response)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if response.Name != "repo" {
		t.Fatalf("decoded name = %q", response.Name)
	}
	if meta.RequestID != "req-123" {
		t.Fatalf("request ID = %q", meta.RequestID)
	}
}

func TestClientRedactsTokenHeaderAndReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(tokenHeader); got != "top-secret" {
			t.Fatalf("token header = %q", got)
		}
		w.Header().Set("x-request-id", "req-fail")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"Forbidden","message":"denied"}`))
	}))
	defer server.Close()

	debug := strings.Builder{}
	client := NewClient(ClientOptions{Endpoint: server.URL, Token: "top-secret", Debug: true, DebugOut: &debug})
	_, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/x?token=top-secret"}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	if apiErr.RequestID != "req-fail" || apiErr.Code != "Forbidden" || apiErr.Message != "denied" {
		t.Fatalf("unexpected API error: %#v", apiErr)
	}
	if strings.Contains(debug.String(), "top-secret") {
		t.Fatalf("debug output leaked token: %s", debug.String())
	}
}

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
