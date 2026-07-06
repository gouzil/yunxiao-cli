package yunxiao

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gouzi/yunxiao-cli/internal/api"
)

type RepositoryService interface {
	ListRepositories(ctx context.Context, request ListRepositoriesRequest) (RepositoryListResult, error)
	GetRepository(ctx context.Context, request GetRepositoryRequest) (RepositoryResult, error)
	CreateRepository(ctx context.Context, request CreateRepositoryRequest) (RepositoryResult, error)
	UpdateRepository(ctx context.Context, request UpdateRepositoryRequest) (RepositoryResult, error)
	ArchiveRepository(ctx context.Context, request RepositoryActionRequest) (RepositoryActionResult, error)
	UnarchiveRepository(ctx context.Context, request RepositoryActionRequest) (RepositoryActionResult, error)
	DeleteRepository(ctx context.Context, request RepositoryActionRequest) (RepositoryActionResult, error)
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (o *Organization) UnmarshalJSON(data []byte) error {
	raw := organizationJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*o = Organization{ID: string(raw.ID), Name: raw.Name}
	return nil
}

type organizationJSON struct {
	ID   flexibleString `json:"id"`
	Name string         `json:"name"`
}

type Repository struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Path          string `json:"path"`
	DefaultBranch string `json:"defaultBranch"`
	Visibility    string `json:"visibility"`
	Archived      bool   `json:"archived"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
	WebURL        string `json:"webUrl"`
	SSHURL        string `json:"sshUrl"`
	HTTPURL       string `json:"httpUrl"`
}

func (r *Repository) UnmarshalJSON(data []byte) error {
	raw := repositoryJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = Repository{
		ID:            string(raw.ID),
		Name:          raw.Name,
		Namespace:     first(raw.Namespace.Path, raw.Namespace.Name, string(raw.Namespace.ID), string(raw.NamespaceID)),
		Path:          first(raw.PathWithNamespace, raw.Path),
		DefaultBranch: raw.DefaultBranch,
		Visibility:    raw.Visibility,
		Archived:      raw.Archived,
		CreatedAt:     raw.CreatedAt,
		UpdatedAt:     first(raw.UpdatedAt, raw.LastActivityAt),
		WebURL:        raw.WebURL,
		SSHURL:        first(raw.SSHURL, raw.SSHURLToRepo),
		HTTPURL:       first(raw.HTTPURL, raw.HTTPURLToRepo),
	}
	return nil
}

type repositoryJSON struct {
	ID                flexibleString      `json:"id"`
	Name              string              `json:"name"`
	Namespace         repositoryNamespace `json:"namespace"`
	NamespaceID       flexibleString      `json:"namespaceId"`
	Path              string              `json:"path"`
	PathWithNamespace string              `json:"pathWithNamespace"`
	DefaultBranch     string              `json:"defaultBranch"`
	Visibility        string              `json:"visibility"`
	Archived          bool                `json:"archived"`
	CreatedAt         string              `json:"createdAt"`
	UpdatedAt         string              `json:"updatedAt"`
	LastActivityAt    string              `json:"lastActivityAt"`
	WebURL            string              `json:"webUrl"`
	SSHURL            string              `json:"sshUrl"`
	SSHURLToRepo      string              `json:"sshUrlToRepo"`
	HTTPURL           string              `json:"httpUrl"`
	HTTPURLToRepo     string              `json:"httpUrlToRepo"`
}

type repositoryNamespace struct {
	ID   flexibleString `json:"id"`
	Name string         `json:"name"`
	Path string         `json:"path"`
}

func (n *repositoryNamespace) UnmarshalJSON(data []byte) error {
	text := ""
	if err := json.Unmarshal(data, &text); err == nil {
		n.Path = text
		return nil
	}
	type namespace repositoryNamespace
	value := namespace{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*n = repositoryNamespace(value)
	return nil
}

type flexibleString string

func (s *flexibleString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*s = ""
		return nil
	}
	text := ""
	if err := json.Unmarshal(data, &text); err == nil {
		*s = flexibleString(text)
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	number := json.Number("")
	if err := decoder.Decode(&number); err != nil {
		return err
	}
	*s = flexibleString(number.String())
	return nil
}

type ListRepositoriesRequest struct {
	Organization string      `json:"organization"`
	Options      ListOptions `json:"options"`
}

type GetRepositoryRequest struct {
	Organization string `json:"organization"`
	RepositoryID string `json:"repositoryId"`
}

type CreateRepositoryRequest struct {
	Organization  string `json:"organization"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Visibility    string `json:"visibility"`
	DefaultBranch string `json:"defaultBranch"`
	Description   string `json:"description"`
}

type UpdateRepositoryRequest struct {
	Organization string `json:"organization"`
	RepositoryID string `json:"repositoryId"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Visibility   string `json:"visibility"`
}

type RepositoryActionRequest struct {
	Organization string `json:"organization"`
	RepositoryID string `json:"repositoryId"`
	Yes          bool   `json:"yes"`
}

type createRepositoryPayload struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
	NamespaceID string `json:"namespaceId,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
}

type updateRepositoryPayload struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
}

type RepositoryListResult struct {
	Repositories []Repository     `json:"repositories"`
	Meta         api.ResponseMeta `json:"meta"`
}

type RepositoryResult struct {
	Repository Repository       `json:"repository"`
	Meta       api.ResponseMeta `json:"meta"`
}

type RepositoryActionResult struct {
	Repository Repository       `json:"repository"`
	Action     string           `json:"action"`
	Meta       api.ResponseMeta `json:"meta"`
	Supported  bool             `json:"supported"`
}

type repositoryListResponse struct {
	Data       []Repository `json:"data"`
	Result     []Repository `json:"result"`
	Items      []Repository `json:"items"`
	Records    []Repository `json:"records"`
	TotalCount int          `json:"totalCount"`
}

func (r *repositoryListResponse) UnmarshalJSON(data []byte) error {
	repositories := []Repository{}
	if err := json.Unmarshal(data, &repositories); err == nil {
		r.Data = repositories
		return nil
	}
	type wrapped repositoryListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = repositoryListResponse(value)
	return nil
}

type repositoryResponse struct {
	Data       Repository `json:"data"`
	Repository Repository `json:"repository"`
}

func (r *repositoryResponse) UnmarshalJSON(data []byte) error {
	repository := Repository{}
	if err := json.Unmarshal(data, &repository); err == nil {
		r.Data = repository
		return nil
	}
	type wrapped repositoryResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = repositoryResponse(value)
	return nil
}

type organizationListResponse struct {
	Organizations []Organization `json:"organizations"`
	Data          []Organization `json:"data"`
	Result        []Organization `json:"result"`
}

func (r *organizationListResponse) UnmarshalJSON(data []byte) error {
	organizations := []Organization{}
	if err := json.Unmarshal(data, &organizations); err == nil {
		r.Organizations = organizations
		return nil
	}
	type wrapped organizationListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = organizationListResponse(value)
	return nil
}

func (s ClientServices) ListRepositories(ctx context.Context, request ListRepositoriesRequest) (RepositoryListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listRepositoriesForAllOrganizations(ctx, request)
	}
	return s.listRepositoriesForOrganization(ctx, request)
}

func (s ClientServices) listRepositoriesForAllOrganizations(ctx context.Context, request ListRepositoriesRequest) (RepositoryListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return RepositoryListResult{}, err
	}
	repositories := []Repository{}
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listRepositoriesForOrganization(ctx, ListRepositoriesRequest{Organization: organization.ID, Options: request.Options})
		if err != nil {
			return RepositoryListResult{}, err
		}
		repositories = append(repositories, result.Repositories...)
		meta = result.Meta
	}
	return RepositoryListResult{Repositories: repositories, Meta: meta}, nil
}

func (s ClientServices) listOrganizations(ctx context.Context) ([]Organization, api.ResponseMeta, error) {
	response := organizationListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: "/oapi/v1/platform/organizations"}, &response)
	if err != nil {
		return nil, meta, err
	}
	return firstOrganizations(response.Organizations, response.Data, response.Result), meta, nil
}

func (s ClientServices) listRepositoriesForOrganization(ctx context.Context, request ListRepositoriesRequest) (RepositoryListResult, error) {
	response := repositoryListResponse{}
	params := queryParams(request.Options)
	requestPath := "/oapi/v1/codeup/repositories"
	if strings.TrimSpace(request.Organization) != "" {
		requestPath = pathf("/oapi/v1/codeup/organizations/%s/repositories", request.Organization)
	}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: requestPath, Query: params}, &response)
	if err != nil {
		return RepositoryListResult{}, err
	}
	return RepositoryListResult{Repositories: firstRepositories(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) GetRepository(ctx context.Context, request GetRepositoryRequest) (RepositoryResult, error) {
	response := repositoryResponse{}
	requestPath := pathf("/oapi/v1/codeup/repositories/%s", request.RepositoryID)
	if strings.TrimSpace(request.Organization) != "" {
		requestPath = pathf("/oapi/v1/codeup/organizations/%s/repositories/%s", request.Organization, request.RepositoryID)
	}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: requestPath}, &response)
	if err != nil {
		return RepositoryResult{}, err
	}
	return RepositoryResult{Repository: firstRepository(response.Data, response.Repository), Meta: meta}, nil
}

func (s ClientServices) CreateRepository(ctx context.Context, request CreateRepositoryRequest) (RepositoryResult, error) {
	payload := createRepositoryPayload{
		Name:        request.Name,
		Path:        request.Name,
		Description: request.Description,
		NamespaceID: request.Namespace,
		Visibility:  request.Visibility,
	}
	body, err := api.EncodeJSONBody(payload)
	if err != nil {
		return RepositoryResult{}, err
	}
	response := repositoryResponse{}
	path := "/oapi/v1/codeup/repositories"
	if strings.TrimSpace(request.Organization) != "" {
		path = pathf("/oapi/v1/codeup/organizations/%s/repositories", request.Organization)
	}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPost, Path: path, Body: body}, &response)
	if err != nil {
		return RepositoryResult{}, err
	}
	return RepositoryResult{Repository: firstRepository(response.Data, response.Repository), Meta: meta}, nil
}

func (s ClientServices) UpdateRepository(ctx context.Context, request UpdateRepositoryRequest) (RepositoryResult, error) {
	payload := updateRepositoryPayload{
		Name:        request.Name,
		Description: request.Description,
		Visibility:  request.Visibility,
	}
	body, err := api.EncodeJSONBody(payload)
	if err != nil {
		return RepositoryResult{}, err
	}
	response := repositoryResponse{}
	path := pathf("/oapi/v1/codeup/repositories/%s", request.RepositoryID)
	if strings.TrimSpace(request.Organization) != "" {
		path = pathf("/oapi/v1/codeup/organizations/%s/repositories/%s", request.Organization, request.RepositoryID)
	}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPut, Path: path, Body: body}, &response)
	if err != nil {
		return RepositoryResult{}, err
	}
	return RepositoryResult{Repository: firstRepository(response.Data, response.Repository), Meta: meta}, nil
}

func (s ClientServices) ArchiveRepository(ctx context.Context, request RepositoryActionRequest) (RepositoryActionResult, error) {
	return s.repositoryAction(ctx, request.Organization, request.RepositoryID, http.MethodPost, "archive", "archive")
}

func (s ClientServices) UnarchiveRepository(ctx context.Context, request RepositoryActionRequest) (RepositoryActionResult, error) {
	return RepositoryActionResult{Action: "unarchive", Supported: false}, nil
}

func (s ClientServices) DeleteRepository(ctx context.Context, request RepositoryActionRequest) (RepositoryActionResult, error) {
	return s.repositoryAction(ctx, request.Organization, request.RepositoryID, http.MethodDelete, "", "delete")
}

func (s ClientServices) repositoryAction(ctx context.Context, organization string, repositoryID string, method string, resource string, action string) (RepositoryActionResult, error) {
	response := repositoryResponse{}
	path := pathf("/oapi/v1/codeup/repositories/%s", repositoryID)
	if strings.TrimSpace(organization) != "" {
		path = pathf("/oapi/v1/codeup/organizations/%s/repositories/%s", organization, repositoryID)
	}
	if strings.TrimSpace(resource) != "" {
		path += "/" + strings.Trim(resource, "/")
	}
	meta, err := s.client.Do(ctx, api.Request{Method: method, Path: path}, &response)
	if err != nil {
		return RepositoryActionResult{}, err
	}
	return RepositoryActionResult{Repository: firstRepository(response.Data, response.Repository, Repository{ID: repositoryID}), Action: action, Meta: meta, Supported: true}, nil
}

func firstRepositories(values ...[]Repository) []Repository {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstOrganizations(values ...[]Organization) []Organization {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstRepository(values ...Repository) Repository {
	for _, value := range values {
		if value.ID != "" || value.Name != "" || value.Path != "" {
			return value
		}
	}
	return Repository{}
}
