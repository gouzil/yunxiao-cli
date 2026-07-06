package yunxiao

import (
	"context"
)

type SearchService interface {
	SearchRepositories(ctx context.Context, request SearchRepositoriesRequest) (SearchRepositoryResult, error)
	SearchCode(ctx context.Context, request SearchCodeRequest) (SearchCodeResult, error)
	SearchCommits(ctx context.Context, request SearchCommitsRequest) (SearchCommitResult, error)
	SearchMergeRequests(ctx context.Context, request SearchMergeRequestsRequest) (SearchMergeRequestResult, error)
	SearchWorkItems(ctx context.Context, request SearchWorkItemsRequest) (SearchWorkItemResult, error)
}

type SearchRepositoriesRequest struct {
	Query        string      `json:"query"`
	Organization string      `json:"organization"`
	Options      ListOptions `json:"options"`
}

type SearchCodeRequest struct {
	Organization string      `json:"organization"`
	Query        string      `json:"query"`
	RepositoryID string      `json:"repositoryId"`
	Ref          string      `json:"ref"`
	Options      ListOptions `json:"options"`
}

type SearchCommitsRequest struct {
	Organization string      `json:"organization"`
	Query        string      `json:"query"`
	RepositoryID string      `json:"repositoryId"`
	Ref          string      `json:"ref"`
	Options      ListOptions `json:"options"`
}

type SearchMergeRequestsRequest struct {
	Organization string      `json:"organization"`
	Query        string      `json:"query"`
	RepositoryID string      `json:"repositoryId"`
	State        string      `json:"state"`
	Options      ListOptions `json:"options"`
}

type SearchWorkItemsRequest struct {
	Organization string      `json:"organization"`
	Query        string      `json:"query"`
	ProjectID    string      `json:"projectId"`
	State        string      `json:"state"`
	Options      ListOptions `json:"options"`
}

type CodeSearchMatch struct {
	Repository string `json:"repository"`
	Ref        string `json:"ref"`
	Path       string `json:"path"`
	Line       int    `json:"line"`
	Match      string `json:"match"`
}

type CommitSearchMatch struct {
	Repository string `json:"repository"`
	Commit     Commit `json:"commit"`
}

type MergeRequestSearchMatch struct {
	Repository   string       `json:"repository"`
	MergeRequest MergeRequest `json:"mergeRequest"`
}

type WorkItemSearchMatch struct {
	WorkItem WorkItem `json:"workItem"`
}

type SearchRepositoryResult struct {
	Repositories []Repository `json:"repositories"`
}

type SearchCodeResult struct {
	Matches []CodeSearchMatch `json:"matches"`
	Legacy  bool              `json:"legacy"`
}

type SearchCommitResult struct {
	Matches []CommitSearchMatch `json:"matches"`
}

type SearchMergeRequestResult struct {
	Matches []MergeRequestSearchMatch `json:"matches"`
}

type SearchWorkItemResult struct {
	Matches []WorkItemSearchMatch `json:"matches"`
}

func (s ClientServices) SearchRepositories(ctx context.Context, request SearchRepositoriesRequest) (SearchRepositoryResult, error) {
	options := request.Options
	options.Query = request.Query
	result, err := s.ListRepositories(ctx, ListRepositoriesRequest{Organization: request.Organization, Options: options})
	if err != nil {
		return SearchRepositoryResult{}, err
	}
	return SearchRepositoryResult{Repositories: result.Repositories}, nil
}

func (s ClientServices) SearchCode(ctx context.Context, request SearchCodeRequest) (SearchCodeResult, error) {
	files, err := s.ListFiles(ctx, ListFilesRequest{Organization: request.Organization, RepositoryID: request.RepositoryID, Ref: request.Ref, Options: request.Options})
	if err != nil {
		return SearchCodeResult{}, err
	}
	matches := make([]CodeSearchMatch, 0, len(files.Files))
	for _, file := range files.Files {
		matches = append(matches, CodeSearchMatch{Repository: request.RepositoryID, Ref: request.Ref, Path: file.Path, Match: request.Query})
	}
	return SearchCodeResult{Matches: matches, Legacy: false}, nil
}

func (s ClientServices) SearchCommits(ctx context.Context, request SearchCommitsRequest) (SearchCommitResult, error) {
	options := request.Options
	options.Query = request.Query
	result, err := s.ListCommits(ctx, ListCommitsRequest{Organization: request.Organization, RepositoryID: request.RepositoryID, Branch: request.Ref, Options: options})
	if err != nil {
		return SearchCommitResult{}, err
	}
	matches := make([]CommitSearchMatch, 0, len(result.Commits))
	for _, commit := range result.Commits {
		matches = append(matches, CommitSearchMatch{Repository: request.RepositoryID, Commit: commit})
	}
	return SearchCommitResult{Matches: matches}, nil
}

func (s ClientServices) SearchMergeRequests(ctx context.Context, request SearchMergeRequestsRequest) (SearchMergeRequestResult, error) {
	options := request.Options
	options.Query = request.Query
	result, err := s.ListMergeRequests(ctx, ListMergeRequestsRequest{Organization: request.Organization, RepositoryID: request.RepositoryID, State: request.State, Options: options})
	if err != nil {
		return SearchMergeRequestResult{}, err
	}
	matches := make([]MergeRequestSearchMatch, 0, len(result.MergeRequests))
	for _, mergeRequest := range result.MergeRequests {
		matches = append(matches, MergeRequestSearchMatch{Repository: request.RepositoryID, MergeRequest: mergeRequest})
	}
	return SearchMergeRequestResult{Matches: matches}, nil
}

func (s ClientServices) SearchWorkItems(ctx context.Context, request SearchWorkItemsRequest) (SearchWorkItemResult, error) {
	options := request.Options
	options.Query = request.Query
	result, err := s.ListWorkItems(ctx, ListWorkItemsRequest{Organization: request.Organization, ProjectID: request.ProjectID, State: request.State, Options: options})
	if err != nil {
		return SearchWorkItemResult{}, err
	}
	matches := make([]WorkItemSearchMatch, 0, len(result.WorkItems))
	for _, workItem := range result.WorkItems {
		matches = append(matches, WorkItemSearchMatch{WorkItem: workItem})
	}
	return SearchWorkItemResult{Matches: matches}, nil
}
