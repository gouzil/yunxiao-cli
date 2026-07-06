package yunxiao

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gouzi/yunxiao-cli/internal/api"
)

type BranchService interface {
	ListBranches(ctx context.Context, request ListBranchesRequest) (BranchListResult, error)
	GetBranch(ctx context.Context, request GetBranchRequest) (BranchResult, error)
}

type CommitService interface {
	ListCommits(ctx context.Context, request ListCommitsRequest) (CommitListResult, error)
	GetCommit(ctx context.Context, request GetCommitRequest) (CommitResult, error)
}

type FileService interface {
	GetFile(ctx context.Context, request GetFileRequest) (FileResult, error)
	ListFiles(ctx context.Context, request ListFilesRequest) (FileTreeResult, error)
}

type SSHKeyService interface {
	ListSSHKeys(ctx context.Context, request ListSSHKeysRequest) (SSHKeyListResult, error)
	CreateSSHKey(ctx context.Context, request CreateSSHKeyRequest) (SSHKeyResult, error)
	DeleteSSHKey(ctx context.Context, request DeleteSSHKeyRequest) (SSHKeyActionResult, error)
}

type Branch struct {
	Name          string `json:"name"`
	Default       bool   `json:"default"`
	Protected     *bool  `json:"protected"`
	CommitSHA     string `json:"commitSha"`
	Author        string `json:"author"`
	UpdatedAt     string `json:"updatedAt"`
	RepositoryRef string `json:"repositoryRef"`
}

func (b *Branch) UnmarshalJSON(data []byte) error {
	raw := branchJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*b = Branch{
		Name:          raw.Name,
		Default:       raw.Default || raw.DefaultBranch,
		Protected:     raw.Protected,
		CommitSHA:     first(raw.CommitSHA, raw.Commit.SHA),
		Author:        first(raw.Author, raw.Commit.AuthorName),
		UpdatedAt:     first(raw.UpdatedAt, raw.Commit.Date),
		RepositoryRef: raw.RepositoryRef,
	}
	return nil
}

type branchJSON struct {
	Name          string `json:"name"`
	Default       bool   `json:"default"`
	DefaultBranch bool   `json:"defaultBranch"`
	Protected     *bool  `json:"protected"`
	CommitSHA     string `json:"commitSha"`
	Commit        Commit `json:"commit"`
	Author        string `json:"author"`
	UpdatedAt     string `json:"updatedAt"`
	RepositoryRef string `json:"repositoryRef"`
}

type Commit struct {
	SHA        string         `json:"sha"`
	ShortSHA   string         `json:"shortSha"`
	Title      string         `json:"title"`
	Message    string         `json:"message"`
	AuthorName string         `json:"authorName"`
	AuthorMail string         `json:"authorEmail"`
	Date       string         `json:"date"`
	Statuses   []CommitStatus `json:"statuses,omitempty"`
}

func (c *Commit) UnmarshalJSON(data []byte) error {
	raw := commitJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*c = Commit{
		SHA:        first(raw.SHA, raw.ID),
		ShortSHA:   first(raw.ShortSHA, raw.ShortID),
		Title:      raw.Title,
		Message:    raw.Message,
		AuthorName: raw.AuthorName,
		AuthorMail: first(raw.AuthorMail, raw.AuthorEmail),
		Date:       first(raw.Date, raw.CommittedDate, raw.AuthoredDate, raw.CreatedAt),
		Statuses:   raw.Statuses,
	}
	return nil
}

type commitJSON struct {
	SHA           string         `json:"sha"`
	ID            string         `json:"id"`
	ShortSHA      string         `json:"shortSha"`
	ShortID       string         `json:"shortId"`
	Title         string         `json:"title"`
	Message       string         `json:"message"`
	AuthorName    string         `json:"authorName"`
	AuthorMail    string         `json:"authorEmail"`
	AuthorEmail   string         `json:"authorMail"`
	Date          string         `json:"date"`
	CommittedDate string         `json:"committedDate"`
	AuthoredDate  string         `json:"authoredDate"`
	CreatedAt     string         `json:"createdAt"`
	Statuses      []CommitStatus `json:"statuses,omitempty"`
}

type CommitStatus struct {
	Context   string `json:"context"`
	State     string `json:"state"`
	TargetURL string `json:"targetUrl"`
}

type FileEntry struct {
	Type     string `json:"type"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	BlobID   string `json:"blobId"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
	Ref      string `json:"ref"`
}

func (f *FileEntry) UnmarshalJSON(data []byte) error {
	raw := fileEntryJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	size := int64(0)
	if raw.Size != "" {
		parsed, err := strconv.ParseInt(string(raw.Size), 10, 64)
		if err != nil {
			return err
		}
		size = parsed
	}
	*f = FileEntry{
		Type:     raw.Type,
		Path:     first(raw.Path, raw.FilePath, raw.Name),
		Size:     size,
		BlobID:   first(raw.BlobID, string(raw.ID)),
		Encoding: raw.Encoding,
		Content:  raw.Content,
		Ref:      raw.Ref,
	}
	return nil
}

type fileEntryJSON struct {
	Type     string         `json:"type"`
	Path     string         `json:"path"`
	FilePath string         `json:"filePath"`
	Name     string         `json:"name"`
	Size     flexibleString `json:"size"`
	BlobID   string         `json:"blobId"`
	ID       flexibleString `json:"id"`
	Encoding string         `json:"encoding"`
	Content  string         `json:"content"`
	Ref      string         `json:"ref"`
}

type SSHKey struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"publicKey"`
	CreatedAt   string `json:"createdAt"`
}

func (k *SSHKey) UnmarshalJSON(data []byte) error {
	raw := sshKeyJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*k = SSHKey{
		ID:          string(raw.ID),
		Title:       raw.Title,
		Fingerprint: first(raw.Fingerprint, raw.FingerPrint),
		PublicKey:   first(raw.PublicKey, raw.Key),
		CreatedAt:   raw.CreatedAt,
	}
	return nil
}

type sshKeyJSON struct {
	ID          flexibleString `json:"id"`
	Title       string         `json:"title"`
	Fingerprint string         `json:"fingerprint"`
	FingerPrint string         `json:"fingerPrint"`
	PublicKey   string         `json:"publicKey"`
	Key         string         `json:"key"`
	CreatedAt   string         `json:"createdAt"`
}

type ListBranchesRequest struct {
	Organization string      `json:"organization"`
	RepositoryID string      `json:"repositoryId"`
	Options      ListOptions `json:"options"`
}

type GetBranchRequest struct {
	Organization string `json:"organization"`
	RepositoryID string `json:"repositoryId"`
	Branch       string `json:"branch"`
}

type ListCommitsRequest struct {
	Organization string      `json:"organization"`
	RepositoryID string      `json:"repositoryId"`
	Branch       string      `json:"branch"`
	Path         string      `json:"path"`
	Since        string      `json:"since"`
	Until        string      `json:"until"`
	Options      ListOptions `json:"options"`
}

type GetCommitRequest struct {
	Organization string `json:"organization"`
	RepositoryID string `json:"repositoryId"`
	SHA          string `json:"sha"`
}

type GetFileRequest struct {
	Organization string `json:"organization"`
	RepositoryID string `json:"repositoryId"`
	Path         string `json:"path"`
	Ref          string `json:"ref"`
	Metadata     bool   `json:"metadata"`
}

type ListFilesRequest struct {
	Organization string      `json:"organization"`
	RepositoryID string      `json:"repositoryId"`
	Path         string      `json:"path"`
	Ref          string      `json:"ref"`
	Options      ListOptions `json:"options"`
}

type ListSSHKeysRequest struct {
	Organization string      `json:"organization"`
	Options      ListOptions `json:"options"`
}

type CreateSSHKeyRequest struct {
	Organization string `json:"organization"`
	Title        string `json:"title"`
	PublicKey    string `json:"key"`
	KeyScope     string `json:"keyScope,omitempty"`
}

type DeleteSSHKeyRequest struct {
	Organization string `json:"organization"`
	ID           string `json:"id"`
	Yes          bool   `json:"yes"`
}

type BranchListResult struct {
	Branches []Branch         `json:"branches"`
	Meta     api.ResponseMeta `json:"meta"`
}

type BranchResult struct {
	Branch Branch           `json:"branch"`
	Meta   api.ResponseMeta `json:"meta"`
}

type CommitListResult struct {
	Commits []Commit         `json:"commits"`
	Meta    api.ResponseMeta `json:"meta"`
}

type CommitResult struct {
	Commit Commit           `json:"commit"`
	Meta   api.ResponseMeta `json:"meta"`
}

type FileResult struct {
	File FileEntry        `json:"file"`
	Meta api.ResponseMeta `json:"meta"`
}

type FileTreeResult struct {
	Files []FileEntry      `json:"files"`
	Meta  api.ResponseMeta `json:"meta"`
}

type SSHKeyListResult struct {
	Keys []SSHKey         `json:"keys"`
	Meta api.ResponseMeta `json:"meta"`
}

type SSHKeyResult struct {
	Key  SSHKey           `json:"key"`
	Meta api.ResponseMeta `json:"meta"`
}

type SSHKeyActionResult struct {
	ID   string           `json:"id"`
	Meta api.ResponseMeta `json:"meta"`
}

type branchListResponse struct {
	Data    []Branch `json:"data"`
	Result  []Branch `json:"result"`
	Items   []Branch `json:"items"`
	Records []Branch `json:"records"`
}

func (r *branchListResponse) UnmarshalJSON(data []byte) error {
	branches := []Branch{}
	if err := json.Unmarshal(data, &branches); err == nil {
		r.Data = branches
		return nil
	}
	type wrapped branchListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = branchListResponse(value)
	return nil
}

type branchResponse struct {
	Data   Branch `json:"data"`
	Branch Branch `json:"branch"`
}

func (r *branchResponse) UnmarshalJSON(data []byte) error {
	type wrapped branchResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err == nil && firstBranch(value.Data, value.Branch).Name != "" {
		*r = branchResponse(value)
		return nil
	}
	branch := Branch{}
	if err := json.Unmarshal(data, &branch); err != nil {
		return err
	}
	r.Data = branch
	return nil
}

type commitListResponse struct {
	Data    []Commit `json:"data"`
	Result  []Commit `json:"result"`
	Items   []Commit `json:"items"`
	Records []Commit `json:"records"`
}

func (r *commitListResponse) UnmarshalJSON(data []byte) error {
	commits := []Commit{}
	if err := json.Unmarshal(data, &commits); err == nil {
		r.Data = commits
		return nil
	}
	type wrapped commitListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = commitListResponse(value)
	return nil
}

type commitResponse struct {
	Data   Commit `json:"data"`
	Commit Commit `json:"commit"`
}

func (r *commitResponse) UnmarshalJSON(data []byte) error {
	type wrapped commitResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err == nil && firstCommit(value.Data, value.Commit).SHA != "" {
		*r = commitResponse(value)
		return nil
	}
	commit := Commit{}
	if err := json.Unmarshal(data, &commit); err != nil {
		return err
	}
	r.Data = commit
	return nil
}

type fileTreeResponse struct {
	Data    []FileEntry `json:"data"`
	Result  []FileEntry `json:"result"`
	Items   []FileEntry `json:"items"`
	Records []FileEntry `json:"records"`
}

func (r *fileTreeResponse) UnmarshalJSON(data []byte) error {
	files := []FileEntry{}
	if err := json.Unmarshal(data, &files); err == nil {
		r.Data = files
		return nil
	}
	type wrapped fileTreeResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = fileTreeResponse(value)
	return nil
}

type fileResponse struct {
	Data FileEntry `json:"data"`
	File FileEntry `json:"file"`
}

func (r *fileResponse) UnmarshalJSON(data []byte) error {
	type wrapped fileResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err == nil && firstFile(value.Data, value.File).Path != "" {
		*r = fileResponse(value)
		return nil
	}
	file := FileEntry{}
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}
	r.Data = file
	return nil
}

type sshKeyListResponse struct {
	Data    []SSHKey `json:"data"`
	Result  []SSHKey `json:"result"`
	Items   []SSHKey `json:"items"`
	Records []SSHKey `json:"records"`
}

func (r *sshKeyListResponse) UnmarshalJSON(data []byte) error {
	keys := []SSHKey{}
	if err := json.Unmarshal(data, &keys); err == nil {
		r.Data = keys
		return nil
	}
	type wrapped sshKeyListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = sshKeyListResponse(value)
	return nil
}

type sshKeyResponse struct {
	Data SSHKey `json:"data"`
	Key  SSHKey `json:"key"`
}

func (r *sshKeyResponse) UnmarshalJSON(data []byte) error {
	key := SSHKey{}
	if err := json.Unmarshal(data, &key); err == nil && firstSSHKey(key).ID != "" {
		r.Data = key
		return nil
	}
	type wrapped sshKeyResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = sshKeyResponse(value)
	return nil
}

func (s ClientServices) ListBranches(ctx context.Context, request ListBranchesRequest) (BranchListResult, error) {
	response := branchListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: repositoryCodePath(request.Organization, request.RepositoryID, "branches"), Query: queryParams(request.Options)}, &response)
	if err != nil {
		return BranchListResult{}, err
	}
	return BranchListResult{Branches: firstBranches(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) GetBranch(ctx context.Context, request GetBranchRequest) (BranchResult, error) {
	response := branchResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: repositoryCodePath(request.Organization, request.RepositoryID, pathf("branches/%s", request.Branch))}, &response)
	if err != nil {
		return BranchResult{}, err
	}
	return BranchResult{Branch: firstBranch(response.Data, response.Branch), Meta: meta}, nil
}

func (s ClientServices) ListCommits(ctx context.Context, request ListCommitsRequest) (CommitListResult, error) {
	params := queryParams(request.Options)
	params = addParam(params, "refName", request.Branch)
	params = addParam(params, "path", request.Path)
	params = addParam(params, "since", request.Since)
	params = addParam(params, "until", request.Until)
	response := commitListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: repositoryCodePath(request.Organization, request.RepositoryID, "commits"), Query: params}, &response)
	if err != nil {
		return CommitListResult{}, err
	}
	return CommitListResult{Commits: firstCommits(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) GetCommit(ctx context.Context, request GetCommitRequest) (CommitResult, error) {
	response := commitResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: repositoryCodePath(request.Organization, request.RepositoryID, pathf("commits/%s", request.SHA))}, &response)
	if err != nil {
		return CommitResult{}, err
	}
	return CommitResult{Commit: firstCommit(response.Data, response.Commit), Meta: meta}, nil
}

func (s ClientServices) GetFile(ctx context.Context, request GetFileRequest) (FileResult, error) {
	params := []api.QueryParam{}
	params = addParam(params, "ref", request.Ref)
	response := fileResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: repositoryCodePath(request.Organization, request.RepositoryID, pathf("files/%s", request.Path)), Query: params}, &response)
	if err != nil {
		return FileResult{}, err
	}
	return FileResult{File: firstFile(response.Data, response.File), Meta: meta}, nil
}

func (s ClientServices) ListFiles(ctx context.Context, request ListFilesRequest) (FileTreeResult, error) {
	params := queryParams(request.Options)
	params = addParam(params, "path", request.Path)
	params = addParam(params, "ref", request.Ref)
	response := fileTreeResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: repositoryCodePath(request.Organization, request.RepositoryID, "files/tree"), Query: params}, &response)
	if err != nil {
		return FileTreeResult{}, err
	}
	return FileTreeResult{Files: firstFiles(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func repositoryCodePath(organization string, repositoryID string, resource string) string {
	if strings.TrimSpace(organization) != "" {
		return pathf("/oapi/v1/codeup/organizations/%s/repositories/%s/%s", organization, repositoryID, resource)
	}
	return pathf("/oapi/v1/codeup/repositories/%s/%s", repositoryID, resource)
}

func (s ClientServices) ListSSHKeys(ctx context.Context, request ListSSHKeysRequest) (SSHKeyListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listSSHKeysForAllOrganizations(ctx, request)
	}
	return s.listSSHKeysForOrganization(ctx, request)
}

func (s ClientServices) listSSHKeysForAllOrganizations(ctx context.Context, request ListSSHKeysRequest) (SSHKeyListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return SSHKeyListResult{}, err
	}
	keys := []SSHKey{}
	seen := map[string]bool{}
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listSSHKeysForOrganization(ctx, ListSSHKeysRequest{Organization: organization.ID, Options: request.Options})
		if err != nil {
			return SSHKeyListResult{}, err
		}
		for _, key := range result.Keys {
			identity := first(key.ID, key.Fingerprint)
			if identity != "" && seen[identity] {
				continue
			}
			if identity != "" {
				seen[identity] = true
			}
			keys = append(keys, key)
		}
		meta = result.Meta
	}
	return SSHKeyListResult{Keys: keys, Meta: meta}, nil
}

func (s ClientServices) listSSHKeysForOrganization(ctx context.Context, request ListSSHKeysRequest) (SSHKeyListResult, error) {
	response := sshKeyListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: sshKeyPath(request.Organization, ""), Query: queryParams(request.Options)}, &response)
	if err != nil {
		return SSHKeyListResult{}, err
	}
	return SSHKeyListResult{Keys: firstSSHKeys(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) CreateSSHKey(ctx context.Context, request CreateSSHKeyRequest) (SSHKeyResult, error) {
	if request.KeyScope == "" {
		request.KeyScope = "ALL"
	}
	body, err := api.EncodeJSONBody(request)
	if err != nil {
		return SSHKeyResult{}, err
	}
	response := sshKeyResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPost, Path: sshKeyPath(request.Organization, ""), Body: body}, &response)
	if err != nil {
		return SSHKeyResult{}, err
	}
	return SSHKeyResult{Key: firstSSHKey(response.Data, response.Key), Meta: meta}, nil
}

func (s ClientServices) DeleteSSHKey(ctx context.Context, request DeleteSSHKeyRequest) (SSHKeyActionResult, error) {
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodDelete, Path: sshKeyPath(request.Organization, request.ID)}, nil)
	if err != nil {
		return SSHKeyActionResult{}, err
	}
	return SSHKeyActionResult{ID: request.ID, Meta: meta}, nil
}

func sshKeyPath(organization string, keyID string) string {
	if strings.TrimSpace(organization) != "" {
		if strings.TrimSpace(keyID) != "" {
			return pathf("/oapi/v1/codeup/organizations/%s/keys/%s", organization, keyID)
		}
		return pathf("/oapi/v1/codeup/organizations/%s/keys", organization)
	}
	if strings.TrimSpace(keyID) != "" {
		return pathf("/oapi/v1/codeup/keys/%s", keyID)
	}
	return "/oapi/v1/codeup/keys"
}

func firstBranches(values ...[]Branch) []Branch {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstBranch(values ...Branch) Branch {
	for _, value := range values {
		if value.Name != "" {
			return value
		}
	}
	return Branch{}
}

func firstCommits(values ...[]Commit) []Commit {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstCommit(values ...Commit) Commit {
	for _, value := range values {
		if value.SHA != "" || value.Title != "" {
			return value
		}
	}
	return Commit{}
}

func firstFiles(values ...[]FileEntry) []FileEntry {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstFile(values ...FileEntry) FileEntry {
	for _, value := range values {
		if value.Path != "" || value.Content != "" {
			return value
		}
	}
	return FileEntry{}
}

func firstSSHKeys(values ...[]SSHKey) []SSHKey {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstSSHKey(values ...SSHKey) SSHKey {
	for _, value := range values {
		if value.ID != "" || value.Title != "" {
			return value
		}
	}
	return SSHKey{}
}
