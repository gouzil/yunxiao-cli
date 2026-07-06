package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gouzil/yunxiao-cli/internal/api"
)

const serviceName = "yunxiao-cli"

type CredentialStore interface {
	Get(endpoint string) (Credential, error)
	Save(credential Credential) (CredentialLocation, error)
	Delete(endpoint string) error
}

type CredentialLocation string

const (
	LocationKeyring CredentialLocation = "keyring"
	LocationFile    CredentialLocation = "config file"
)

type Credential struct {
	Endpoint string             `json:"endpoint"`
	Token    string             `json:"token"`
	User     AuthenticatedUser  `json:"user"`
	Scopes   []string           `json:"scopes,omitempty"`
	Active   bool               `json:"active"`
	Location CredentialLocation `json:"location"`
}

type AuthenticatedUser struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Organization string `json:"organization"`
}

type TokenVerifier interface {
	VerifyToken(ctx context.Context, endpoint string, token string) (AuthenticatedUser, []string, error)
}

type LoginRequest struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

type LoginResult struct {
	Credential Credential         `json:"credential"`
	Location   CredentialLocation `json:"location"`
}

type Manager struct {
	store    CredentialStore
	verifier TokenVerifier
}

func NewManager(store CredentialStore, verifier TokenVerifier) *Manager {
	return &Manager{store: store, verifier: verifier}
}

func (m *Manager) Login(ctx context.Context, request LoginRequest) (LoginResult, error) {
	endpoint := strings.TrimSpace(request.Endpoint)
	if endpoint == "" {
		endpoint = api.DefaultEndpoint
	}
	token := strings.TrimSpace(request.Token)
	if token == "" {
		return LoginResult{}, errors.New("token is required")
	}
	user, scopes, err := m.verifier.VerifyToken(ctx, endpoint, token)
	if err != nil && !canSaveWithoutUserLookup(err) {
		return LoginResult{}, err
	}
	credential := Credential{
		Endpoint: endpoint,
		Token:    token,
		User:     user,
		Scopes:   scopes,
		Active:   true,
	}
	location, err := m.store.Save(credential)
	if err != nil {
		return LoginResult{}, err
	}
	credential.Location = location
	return LoginResult{Credential: credential, Location: location}, nil
}

func canSaveWithoutUserLookup(err error) bool {
	var apiErr *api.Error
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.HTTPStatus == http.StatusForbidden || apiErr.Code == "UnsupportedInCurrentEnv"
}

func (m *Manager) Status(endpoint string) (Credential, error) {
	if endpoint == "" {
		endpoint = api.DefaultEndpoint
	}
	return m.store.Get(endpoint)
}

func (m *Manager) Logout(endpoint string) (Credential, error) {
	if endpoint == "" {
		endpoint = api.DefaultEndpoint
	}
	credential, err := m.store.Get(endpoint)
	if err != nil {
		return Credential{}, err
	}
	if err := m.store.Delete(endpoint); err != nil {
		return Credential{}, err
	}
	return credential, nil
}

func MaskToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return "unknown"
	}
	prefixLen := 4
	if len(token) < prefixLen {
		prefixLen = len(token)
	}
	return token[:prefixLen] + "********************************"
}

type KeyringBackend interface {
	Get(service string, user string) (string, error)
	Set(service string, user string, password string) error
	Delete(service string, user string) error
}

type FileCredentialStore struct {
	path string
}

func NewFileCredentialStore(path string) *FileCredentialStore {
	return &FileCredentialStore{path: path}
}

func (s *FileCredentialStore) Get(endpoint string) (Credential, error) {
	file, err := readCredentialFile(s.path)
	if err != nil {
		return Credential{}, err
	}
	for _, credential := range file.Credentials {
		if credential.Endpoint == endpoint {
			credential.Location = LocationFile
			return credential, nil
		}
	}
	return Credential{}, ErrNotLoggedIn{Endpoint: endpoint}
}

func (s *FileCredentialStore) Save(credential Credential) (CredentialLocation, error) {
	file, err := readCredentialFile(s.path)
	if err != nil {
		return "", err
	}
	found := false
	for index := range file.Credentials {
		if file.Credentials[index].Endpoint == credential.Endpoint {
			file.Credentials[index] = credential
			found = true
			break
		}
	}
	if !found {
		file.Credentials = append(file.Credentials, credential)
	}
	return LocationFile, writeCredentialFile(s.path, file)
}

func (s *FileCredentialStore) Delete(endpoint string) error {
	file, err := readCredentialFile(s.path)
	if err != nil {
		return err
	}
	next := make([]Credential, 0, len(file.Credentials))
	found := false
	for _, credential := range file.Credentials {
		if credential.Endpoint == endpoint {
			found = true
			continue
		}
		next = append(next, credential)
	}
	if !found {
		return ErrNotLoggedIn{Endpoint: endpoint}
	}
	file.Credentials = next
	return writeCredentialFile(s.path, file)
}

type credentialFile struct {
	Version     int          `json:"version"`
	Credentials []Credential `json:"credentials"`
}

func readCredentialFile(path string) (credentialFile, error) {
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return credentialFile{Version: 1}, nil
	}
	if err != nil {
		return credentialFile{}, err
	}
	file := credentialFile{}
	if err := json.Unmarshal(body, &file); err != nil {
		return credentialFile{}, err
	}
	if file.Version == 0 {
		file.Version = 1
	}
	return file, nil
}

func writeCredentialFile(path string, file credentialFile) error {
	if file.Version == 0 {
		file.Version = 1
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	body, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(path, body, 0o600)
}

type CascadingStore struct {
	keyring KeyringBackend
	file    *FileCredentialStore
}

func NewCascadingStore(keyring KeyringBackend, file *FileCredentialStore) *CascadingStore {
	return &CascadingStore{keyring: keyring, file: file}
}

func (s *CascadingStore) Get(endpoint string) (Credential, error) {
	if s.keyring != nil {
		token, err := s.keyring.Get(serviceName, endpoint)
		if err == nil && strings.TrimSpace(token) != "" {
			credential, fileErr := s.file.Get(endpoint)
			if fileErr != nil {
				credential = Credential{Endpoint: endpoint, Active: true}
			}
			credential.Token = token
			credential.Location = LocationKeyring
			return credential, nil
		}
	}
	return s.file.Get(endpoint)
}

func (s *CascadingStore) Save(credential Credential) (CredentialLocation, error) {
	if s.keyring != nil {
		if err := s.keyring.Set(serviceName, credential.Endpoint, credential.Token); err == nil {
			metadata := credential
			metadata.Token = ""
			metadata.Location = LocationKeyring
			_, fileErr := s.file.Save(metadata)
			if fileErr != nil {
				return "", fileErr
			}
			return LocationKeyring, nil
		}
	}
	return s.file.Save(credential)
}

func (s *CascadingStore) Delete(endpoint string) error {
	keyringErr := error(nil)
	if s.keyring != nil {
		keyringErr = s.keyring.Delete(serviceName, endpoint)
	}
	fileErr := s.file.Delete(endpoint)
	if fileErr == nil || keyringErr == nil {
		return nil
	}
	return fileErr
}

type ErrNotLoggedIn struct {
	Endpoint string
}

func (e ErrNotLoggedIn) Error() string {
	return fmt.Sprintf("not logged in to %s", e.Endpoint)
}
