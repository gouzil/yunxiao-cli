package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gouzil/yunxiao-cli/internal/api"
)

func TestLoginSavesTokenWhenUserLookupIsForbidden(t *testing.T) {
	store := &memoryCredentialStore{}
	manager := NewManager(store, fakeVerifier{err: &api.Error{HTTPStatus: http.StatusForbidden, Code: "Forbidden"}})

	result, err := manager.Login(context.Background(), LoginRequest{Endpoint: "openapi-rdc.aliyuncs.com", Token: "secret-token"})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if store.saved.Token != "secret-token" {
		t.Fatalf("saved token = %q", store.saved.Token)
	}
	if result.Credential.User.DisplayName != "" {
		t.Fatalf("user = %#v", result.Credential.User)
	}
}

func TestLoginRejectsInvalidToken(t *testing.T) {
	manager := NewManager(&memoryCredentialStore{}, fakeVerifier{err: &api.Error{HTTPStatus: http.StatusUnauthorized, Code: "InvalidTokenError"}})

	_, err := manager.Login(context.Background(), LoginRequest{Endpoint: "openapi-rdc.aliyuncs.com", Token: "bad-token"})
	if err == nil {
		t.Fatal("expected error")
	}
}

type memoryCredentialStore struct {
	saved Credential
}

func (s *memoryCredentialStore) Get(endpoint string) (Credential, error) {
	if s.saved.Endpoint == endpoint {
		return s.saved, nil
	}
	return Credential{}, ErrNotLoggedIn{Endpoint: endpoint}
}

func (s *memoryCredentialStore) Save(credential Credential) (CredentialLocation, error) {
	s.saved = credential
	return LocationFile, nil
}

func (s *memoryCredentialStore) Delete(endpoint string) error {
	if s.saved.Endpoint != endpoint {
		return ErrNotLoggedIn{Endpoint: endpoint}
	}
	s.saved = Credential{}
	return nil
}

type fakeVerifier struct {
	err error
}

func (v fakeVerifier) VerifyToken(ctx context.Context, endpoint string, token string) (AuthenticatedUser, []string, error) {
	if v.err != nil {
		return AuthenticatedUser{}, nil, v.err
	}
	if token == "" {
		return AuthenticatedUser{}, nil, errors.New("empty token")
	}
	return AuthenticatedUser{ID: "user-1", DisplayName: "Test User"}, []string{"scope-1"}, nil
}
