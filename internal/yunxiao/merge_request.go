package yunxiao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gouzil/yunxiao-cli/internal/api"
)

type MergeRequestService interface {
	ListMergeRequests(ctx context.Context, request ListMergeRequestsRequest) (MergeRequestListResult, error)
	GetMergeRequest(ctx context.Context, request GetMergeRequestRequest) (MergeRequestResult, error)
	CreateMergeRequest(ctx context.Context, request CreateMergeRequestRequest) (MergeRequestResult, error)
	UpdateMergeRequest(ctx context.Context, request UpdateMergeRequestRequest) (MergeRequestResult, error)
	GetMergeRequestDiff(ctx context.Context, request GetMergeRequestDiffRequest) (MergeRequestDiffResult, error)
	ListMergeRequestFiles(ctx context.Context, request ListMergeRequestFilesRequest) (MergeRequestFileListResult, error)
	CommentMergeRequest(ctx context.Context, request CommentMergeRequestRequest) (CommentResult, error)
	ResolveMergeRequestComment(ctx context.Context, request ResolveMergeRequestCommentRequest) (CommentResult, error)
	ReviewMergeRequest(ctx context.Context, request ReviewMergeRequestRequest) (ReviewResult, error)
	MergeMergeRequest(ctx context.Context, request MergeRequestActionRequest) (MergeRequestActionResult, error)
	CloseMergeRequest(ctx context.Context, request MergeRequestActionRequest) (MergeRequestActionResult, error)
	ReopenMergeRequest(ctx context.Context, request MergeRequestActionRequest) (MergeRequestActionResult, error)
	GetMergeRequestStatus(ctx context.Context, request GetMergeRequestRequest) (MergeRequestStatusResult, error)
}

type MergeRequest struct {
	ID              string   `json:"id"`
	IID             string   `json:"iid"`
	ProjectID       string   `json:"projectId"`
	SourceProjectID string   `json:"sourceProjectId"`
	TargetProjectID string   `json:"targetProjectId"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	State           string   `json:"state"`
	Author          string   `json:"author"`
	SourceBranch    string   `json:"sourceBranch"`
	TargetBranch    string   `json:"targetBranch"`
	ReviewStatus    string   `json:"reviewStatus"`
	Mergeable       *bool    `json:"mergeable"`
	HasConflicts    *bool    `json:"hasConflicts"`
	PipelineStatus  string   `json:"pipelineStatus"`
	Labels          []string `json:"labels"`
	WebURL          string   `json:"webUrl"`
	UpdatedAt       string   `json:"updatedAt"`
}

func (m *MergeRequest) UnmarshalJSON(data []byte) error {
	raw := mergeRequestJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*m = MergeRequest{
		ID:              first(raw.ID, raw.MRBizID, raw.BizID),
		IID:             first(raw.IID, string(raw.LocalID)),
		ProjectID:       string(raw.ProjectID),
		SourceProjectID: string(raw.SourceProjectID),
		TargetProjectID: string(raw.TargetProjectID),
		Title:           raw.Title,
		Description:     raw.Description,
		State:           first(raw.State, raw.Status),
		Author:          raw.Author.Name,
		SourceBranch:    raw.SourceBranch,
		TargetBranch:    raw.TargetBranch,
		ReviewStatus:    raw.ReviewStatus,
		Mergeable:       raw.Mergeable,
		HasConflicts:    firstBool(raw.HasConflicts, raw.HasConflict),
		PipelineStatus:  raw.PipelineStatus,
		Labels:          raw.Labels,
		WebURL:          first(raw.WebURL, raw.DetailURL),
		UpdatedAt:       first(raw.UpdatedAt, raw.UpdateTime),
	}
	return nil
}

type mergeRequestJSON struct {
	ID              string             `json:"id"`
	IID             string             `json:"iid"`
	LocalID         flexibleString     `json:"localId"`
	BizID           string             `json:"bizId"`
	MRBizID         string             `json:"mrBizId"`
	ProjectID       flexibleString     `json:"projectId"`
	SourceProjectID flexibleString     `json:"sourceProjectId"`
	TargetProjectID flexibleString     `json:"targetProjectId"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	State           string             `json:"state"`
	Status          string             `json:"status"`
	Author          mergeRequestAuthor `json:"author"`
	SourceBranch    string             `json:"sourceBranch"`
	TargetBranch    string             `json:"targetBranch"`
	ReviewStatus    string             `json:"reviewStatus"`
	Mergeable       *bool              `json:"mergeable"`
	HasConflicts    *bool              `json:"hasConflicts"`
	HasConflict     *bool              `json:"hasConflict"`
	PipelineStatus  string             `json:"pipelineStatus"`
	Labels          []string           `json:"labels"`
	WebURL          string             `json:"webUrl"`
	DetailURL       string             `json:"detailUrl"`
	UpdatedAt       string             `json:"updatedAt"`
	UpdateTime      string             `json:"updateTime"`
}

type mergeRequestAuthor struct {
	Name string `json:"name"`
}

func (a *mergeRequestAuthor) UnmarshalJSON(data []byte) error {
	name := ""
	if err := json.Unmarshal(data, &name); err == nil {
		a.Name = name
		return nil
	}
	type object mergeRequestAuthor
	value := object{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*a = mergeRequestAuthor(value)
	return nil
}

type MergeRequestFile struct {
	Path      string `json:"path"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

func (f *MergeRequestFile) UnmarshalJSON(data []byte) error {
	type rawFile struct {
		Path        string          `json:"path"`
		Status      string          `json:"status"`
		Additions   int             `json:"additions"`
		AddLines    int             `json:"addLines"`
		Deletions   int             `json:"deletions"`
		DelLines    int             `json:"delLines"`
		NewPath     string          `json:"newPath"`
		OldPath     string          `json:"oldPath"`
		NewFile     json.RawMessage `json:"newFile"`
		DeletedFile json.RawMessage `json:"deletedFile"`
		RenamedFile json.RawMessage `json:"renamedFile"`
	}
	raw := rawFile{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	status := raw.Status
	switch {
	case status != "":
	case rawBool(raw.NewFile):
		status = "added"
	case rawBool(raw.DeletedFile):
		status = "deleted"
	case rawBool(raw.RenamedFile):
		status = "renamed"
	default:
		status = "modified"
	}
	*f = MergeRequestFile{
		Path:      first(raw.Path, raw.NewPath, raw.OldPath),
		Status:    status,
		Additions: firstInt(raw.Additions, raw.AddLines),
		Deletions: firstInt(raw.Deletions, raw.DelLines),
	}
	return nil
}

type MergeCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Passed bool   `json:"passed"`
}

type ListMergeRequestsRequest struct {
	Organization string      `json:"organization"`
	RepositoryID string      `json:"repositoryId"`
	State        string      `json:"state"`
	Options      ListOptions `json:"options"`
}

type GetMergeRequestRequest struct {
	Organization   string `json:"organization"`
	RepositoryID   string `json:"repositoryId"`
	MergeRequestID string `json:"mergeRequestId"`
}

type CreateMergeRequestRequest struct {
	Organization string `json:"organization"`
	RepositoryID string `json:"repositoryId"`
	SourceBranch string `json:"sourceBranch"`
	TargetBranch string `json:"targetBranch"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Draft        bool   `json:"draft"`
}

type createMergeRequestPayload struct {
	Description        string         `json:"description,omitempty"`
	Draft              bool           `json:"draft,omitempty"`
	SourceBranch       string         `json:"sourceBranch"`
	SourceProjectID    mergeProjectID `json:"sourceProjectId"`
	TargetBranch       string         `json:"targetBranch"`
	TargetProjectID    mergeProjectID `json:"targetProjectId"`
	Title              string         `json:"title"`
	TriggerAIReviewRun bool           `json:"triggerAIReviewRun"`
}

type mergeProjectID string

func (id mergeProjectID) MarshalJSON() ([]byte, error) {
	if _, err := strconv.ParseInt(string(id), 10, 64); err == nil {
		return []byte(id), nil
	}
	return json.Marshal(string(id))
}

type UpdateMergeRequestRequest struct {
	Organization   string `json:"organization"`
	RepositoryID   string `json:"repositoryId"`
	MergeRequestID string `json:"mergeRequestId"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	TargetBranch   string `json:"targetBranch"`
}

type GetMergeRequestDiffRequest struct {
	Organization   string `json:"organization"`
	RepositoryID   string `json:"repositoryId"`
	MergeRequestID string `json:"mergeRequestId"`
	Version        string `json:"version"`
}

type ListMergeRequestFilesRequest struct {
	Organization   string `json:"organization"`
	RepositoryID   string `json:"repositoryId"`
	MergeRequestID string `json:"mergeRequestId"`
	Version        string `json:"version"`
}

type CommentMergeRequestRequest struct {
	Organization   string `json:"organization"`
	RepositoryID   string `json:"repositoryId"`
	MergeRequestID string `json:"mergeRequestId"`
	Body           string `json:"body"`
	Draft          bool   `json:"draft"`
	Resolved       bool   `json:"resolved"`
	PatchSetID     string `json:"patchSetId"`
	FilePath       string `json:"filePath"`
	LineNumber     int    `json:"lineNumber"`
	FromPatchSetID string `json:"fromPatchSetId"`
	ToPatchSetID   string `json:"toPatchSetId"`
}

type ResolveMergeRequestCommentRequest struct {
	Organization   string `json:"organization"`
	RepositoryID   string `json:"repositoryId"`
	MergeRequestID string `json:"mergeRequestId"`
	CommentID      string `json:"commentId"`
	Resolved       bool   `json:"resolved"`
}

type ReviewDecision string

const (
	ReviewDecisionApprove          ReviewDecision = "approved"
	ReviewDecisionChangesRequested ReviewDecision = "changes_requested"
)

type ReviewMergeRequestRequest struct {
	Organization   string         `json:"organization"`
	RepositoryID   string         `json:"repositoryId"`
	MergeRequestID string         `json:"mergeRequestId"`
	Decision       ReviewDecision `json:"decision"`
	Body           string         `json:"body"`
}

type MergeRequestActionRequest struct {
	Organization   string `json:"organization"`
	RepositoryID   string `json:"repositoryId"`
	MergeRequestID string `json:"mergeRequestId"`
	Yes            bool   `json:"yes"`
}

type MergeRequestListResult struct {
	MergeRequests []MergeRequest   `json:"mergeRequests"`
	Meta          api.ResponseMeta `json:"meta"`
}

type MergeRequestResult struct {
	MergeRequest MergeRequest     `json:"mergeRequest"`
	Meta         api.ResponseMeta `json:"meta"`
}

type MergeRequestDiffResult struct {
	Diff string           `json:"diff"`
	Meta api.ResponseMeta `json:"meta"`
}

type MergeRequestFileListResult struct {
	Files []MergeRequestFile `json:"files"`
	Meta  api.ResponseMeta   `json:"meta"`
}

type CommentResult struct {
	ID   string           `json:"id"`
	Meta api.ResponseMeta `json:"meta"`
}

type ReviewResult struct {
	MergeRequestID string           `json:"mergeRequestId"`
	ReviewStatus   string           `json:"reviewStatus"`
	Meta           api.ResponseMeta `json:"meta"`
}

type MergeRequestActionResult struct {
	MergeRequest MergeRequest     `json:"mergeRequest"`
	Action       string           `json:"action"`
	MergeCommit  string           `json:"mergeCommit"`
	Meta         api.ResponseMeta `json:"meta"`
}

type MergeRequestStatusResult struct {
	MergeRequest MergeRequest     `json:"mergeRequest"`
	Checks       []MergeCheck     `json:"checks"`
	Meta         api.ResponseMeta `json:"meta"`
}

type mergeRequestListResponse struct {
	Data    []MergeRequest `json:"data"`
	Result  []MergeRequest `json:"result"`
	Items   []MergeRequest `json:"items"`
	Records []MergeRequest `json:"records"`
}

func (r *mergeRequestListResponse) UnmarshalJSON(data []byte) error {
	mergeRequests := []MergeRequest{}
	if err := json.Unmarshal(data, &mergeRequests); err == nil {
		r.Data = mergeRequests
		return nil
	}
	type wrapped mergeRequestListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = mergeRequestListResponse(value)
	return nil
}

type mergeRequestResponse struct {
	Data           MergeRequest `json:"data"`
	MergeRequest   MergeRequest `json:"mergeRequest"`
	MergeCommit    string       `json:"mergeCommit"`
	MergedRevision string       `json:"mergedRevision"`
}

func (r *mergeRequestResponse) UnmarshalJSON(data []byte) error {
	type wrapped mergeRequestResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err == nil && firstMergeRequest(value.Data, value.MergeRequest).ID != "" {
		*r = mergeRequestResponse(value)
		return nil
	}
	mergeRequest := MergeRequest{}
	if err := json.Unmarshal(data, &mergeRequest); err != nil {
		return err
	}
	raw := struct {
		MergedRevision string `json:"mergedRevision"`
	}{}
	_ = json.Unmarshal(data, &raw)
	r.Data = mergeRequest
	r.MergedRevision = raw.MergedRevision
	return nil
}

type mergeRequestFileResponse struct {
	Data             []MergeRequestFile `json:"data"`
	Result           []MergeRequestFile `json:"result"`
	Items            []MergeRequestFile `json:"items"`
	Records          []MergeRequestFile `json:"records"`
	ChangedTreeItems []MergeRequestFile `json:"changedTreeItems"`
}

type mergeRequestDiffResponse struct {
	Data string `json:"data"`
	Diff string `json:"diff"`
}

type mergeRequestPatchSet struct {
	ID        string `json:"patchSetBizId"`
	ItemType  string `json:"relatedMergeItemType"`
	VersionNo int    `json:"versionNo"`
}

type mergeRequestPatchSetResponse struct {
	Data    []mergeRequestPatchSet `json:"data"`
	Result  []mergeRequestPatchSet `json:"result"`
	Items   []mergeRequestPatchSet `json:"items"`
	Records []mergeRequestPatchSet `json:"records"`
}

func (r *mergeRequestPatchSetResponse) UnmarshalJSON(data []byte) error {
	patches := []mergeRequestPatchSet{}
	if err := json.Unmarshal(data, &patches); err == nil {
		r.Data = patches
		return nil
	}
	type wrapped mergeRequestPatchSetResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = mergeRequestPatchSetResponse(value)
	return nil
}

type commentResponse struct {
	Data struct {
		ID           string `json:"id"`
		CommentBizID string `json:"comment_biz_id"`
	} `json:"data"`
	ID           string `json:"id"`
	CommentBizID string `json:"comment_biz_id"`
}

type commentPayload struct {
	CommentType       string `json:"comment_type"`
	Content           string `json:"content"`
	Draft             bool   `json:"draft"`
	Resolved          bool   `json:"resolved"`
	PatchSetBizID     string `json:"patchset_biz_id"`
	FilePath          string `json:"file_path,omitempty"`
	LineNumber        int    `json:"line_number,omitempty"`
	FromPatchSetBizID string `json:"from_patchset_biz_id,omitempty"`
	ToPatchSetBizID   string `json:"to_patchset_biz_id,omitempty"`
}

func (s ClientServices) ListMergeRequests(ctx context.Context, request ListMergeRequestsRequest) (MergeRequestListResult, error) {
	params := queryParams(request.Options)
	params = addParam(params, "state", request.State)
	response := mergeRequestListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: mergeRequestListPath(request.Organization), Query: params}, &response)
	if err != nil {
		return MergeRequestListResult{}, err
	}
	mergeRequests := firstMergeRequests(response.Data, response.Result, response.Items, response.Records)
	return MergeRequestListResult{MergeRequests: filterMergeRequestsByRepository(mergeRequests, request.RepositoryID), Meta: meta}, nil
}

func (s ClientServices) GetMergeRequest(ctx context.Context, request GetMergeRequestRequest) (MergeRequestResult, error) {
	response := mergeRequestResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: mergeRequestRepositoryPath(request.Organization, request.RepositoryID, pathf("changeRequests/%s", request.MergeRequestID))}, &response)
	if err != nil {
		return MergeRequestResult{}, err
	}
	return MergeRequestResult{MergeRequest: firstMergeRequest(response.Data, response.MergeRequest), Meta: meta}, nil
}

func (s ClientServices) CreateMergeRequest(ctx context.Context, request CreateMergeRequestRequest) (MergeRequestResult, error) {
	payload := createMergeRequestPayload{
		Description:        request.Description,
		Draft:              request.Draft,
		SourceBranch:       request.SourceBranch,
		SourceProjectID:    mergeProjectID(request.RepositoryID),
		TargetBranch:       request.TargetBranch,
		TargetProjectID:    mergeProjectID(request.RepositoryID),
		Title:              request.Title,
		TriggerAIReviewRun: false,
	}
	body, err := api.EncodeJSONBody(payload)
	if err != nil {
		return MergeRequestResult{}, err
	}
	response := mergeRequestResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPost, Path: mergeRequestRepositoryPath(request.Organization, request.RepositoryID, "changeRequests"), Body: body}, &response)
	if err != nil {
		return MergeRequestResult{}, err
	}
	return MergeRequestResult{MergeRequest: firstMergeRequest(response.Data, response.MergeRequest), Meta: meta}, nil
}

func (s ClientServices) UpdateMergeRequest(ctx context.Context, request UpdateMergeRequestRequest) (MergeRequestResult, error) {
	body, err := api.EncodeJSONBody(request)
	if err != nil {
		return MergeRequestResult{}, err
	}
	response := mergeRequestResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPut, Path: mergeRequestRepositoryPath(request.Organization, request.RepositoryID, pathf("changeRequests/%s", request.MergeRequestID)), Body: body}, &response)
	if err != nil {
		return MergeRequestResult{}, err
	}
	return MergeRequestResult{MergeRequest: firstMergeRequest(response.Data, response.MergeRequest), Meta: meta}, nil
}

func (s ClientServices) GetMergeRequestDiff(ctx context.Context, request GetMergeRequestDiffRequest) (MergeRequestDiffResult, error) {
	result, err := s.ListMergeRequestFiles(ctx, ListMergeRequestFilesRequest{
		Organization:   request.Organization,
		RepositoryID:   request.RepositoryID,
		MergeRequestID: request.MergeRequestID,
		Version:        request.Version,
	})
	if err != nil {
		return MergeRequestDiffResult{}, err
	}
	return MergeRequestDiffResult{Diff: mergeRequestDiffSummary(result.Files), Meta: result.Meta}, nil
}

func (s ClientServices) ListMergeRequestFiles(ctx context.Context, request ListMergeRequestFilesRequest) (MergeRequestFileListResult, error) {
	fromPatchSetID, toPatchSetID, err := s.mergeRequestPatchSetIDs(ctx, request.Organization, request.RepositoryID, request.MergeRequestID)
	if err != nil {
		return MergeRequestFileListResult{}, err
	}
	params := []api.QueryParam{
		{Name: "fromPatchSetId", Value: fromPatchSetID},
		{Name: "toPatchSetId", Value: toPatchSetID},
	}
	response := mergeRequestFileResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: mergeRequestRepositoryPath(request.Organization, request.RepositoryID, pathf("changeRequests/%s/diffs/changeTree", request.MergeRequestID)), Query: params}, &response)
	if err != nil {
		return MergeRequestFileListResult{}, err
	}
	return MergeRequestFileListResult{Files: firstMergeRequestFiles(response.ChangedTreeItems, response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) mergeRequestPatchSetIDs(ctx context.Context, organization string, repositoryID string, mergeRequestID string) (string, string, error) {
	response := mergeRequestPatchSetResponse{}
	_, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: mergeRequestRepositoryPath(organization, repositoryID, pathf("changeRequests/%s/diffs/patches", mergeRequestID))}, &response)
	if err != nil {
		return "", "", err
	}
	fromPatchSetID := ""
	toPatchSetID := ""
	for _, patch := range firstMergeRequestPatchSets(response.Data, response.Result, response.Items, response.Records) {
		switch strings.ToUpper(patch.ItemType) {
		case "MERGE_TARGET":
			fromPatchSetID = patch.ID
		case "MERGE_SOURCE":
			toPatchSetID = patch.ID
		}
	}
	if fromPatchSetID == "" || toPatchSetID == "" {
		return "", "", fmt.Errorf("merge request %s has no comparable patch sets", mergeRequestID)
	}
	return fromPatchSetID, toPatchSetID, nil
}

func (s ClientServices) CommentMergeRequest(ctx context.Context, request CommentMergeRequestRequest) (CommentResult, error) {
	payload := commentPayload{CommentType: "GLOBAL_COMMENT", Content: request.Body, Draft: request.Draft, Resolved: request.Resolved, PatchSetBizID: request.PatchSetID}
	if request.FilePath != "" || request.LineNumber > 0 {
		if request.FilePath == "" || request.LineNumber <= 0 {
			return CommentResult{}, fmt.Errorf("inline comment requires file path and line number")
		}
		fromPatchSetID := request.FromPatchSetID
		toPatchSetID := request.ToPatchSetID
		if fromPatchSetID == "" || toPatchSetID == "" {
			var err error
			fromPatchSetID, toPatchSetID, err = s.mergeRequestPatchSetIDs(ctx, request.Organization, request.RepositoryID, request.MergeRequestID)
			if err != nil {
				return CommentResult{}, err
			}
		}
		payload.CommentType = "INLINE_COMMENT"
		payload.FilePath = request.FilePath
		payload.LineNumber = request.LineNumber
		payload.FromPatchSetBizID = fromPatchSetID
		payload.ToPatchSetBizID = toPatchSetID
		if payload.PatchSetBizID == "" {
			payload.PatchSetBizID = toPatchSetID
		}
	} else if payload.PatchSetBizID == "" {
		_, toPatchSetID, err := s.mergeRequestPatchSetIDs(ctx, request.Organization, request.RepositoryID, request.MergeRequestID)
		if err != nil {
			return CommentResult{}, err
		}
		payload.PatchSetBizID = toPatchSetID
	}
	body, err := api.EncodeJSONBody(payload)
	if err != nil {
		return CommentResult{}, err
	}
	response := commentResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPost, Path: mergeRequestRepositoryPath(request.Organization, request.RepositoryID, pathf("changeRequests/%s/comments", request.MergeRequestID)), Body: body}, &response)
	if err != nil {
		return CommentResult{}, err
	}
	return CommentResult{ID: first(response.ID, response.CommentBizID, response.Data.ID, response.Data.CommentBizID), Meta: meta}, nil
}

func (s ClientServices) ResolveMergeRequestComment(ctx context.Context, request ResolveMergeRequestCommentRequest) (CommentResult, error) {
	body, err := api.EncodeJSONBody(map[string]bool{"resolved": request.Resolved})
	if err != nil {
		return CommentResult{}, err
	}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPut, Path: mergeRequestRepositoryPath(request.Organization, request.RepositoryID, pathf("changeRequests/%s/comments/%s", request.MergeRequestID, request.CommentID)), Body: body}, nil)
	if err != nil {
		return CommentResult{}, err
	}
	return CommentResult{ID: request.CommentID, Meta: meta}, nil
}

func (s ClientServices) ReviewMergeRequest(ctx context.Context, request ReviewMergeRequestRequest) (ReviewResult, error) {
	body, err := api.EncodeJSONBody(request)
	if err != nil {
		return ReviewResult{}, err
	}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPost, Path: mergeRequestRepositoryPath(request.Organization, request.RepositoryID, pathf("changeRequests/%s/review", request.MergeRequestID)), Body: body}, nil)
	if err != nil {
		return ReviewResult{}, err
	}
	return ReviewResult{MergeRequestID: request.MergeRequestID, ReviewStatus: string(request.Decision), Meta: meta}, nil
}

func (s ClientServices) MergeMergeRequest(ctx context.Context, request MergeRequestActionRequest) (MergeRequestActionResult, error) {
	body, err := api.EncodeJSONBody(map[string]string{"mergeType": "ff-only"})
	if err != nil {
		return MergeRequestActionResult{}, err
	}
	return s.mergeRequestAction(ctx, request.Organization, request.RepositoryID, request.MergeRequestID, http.MethodPost, "merge", body)
}

func (s ClientServices) CloseMergeRequest(ctx context.Context, request MergeRequestActionRequest) (MergeRequestActionResult, error) {
	return s.mergeRequestAction(ctx, request.Organization, request.RepositoryID, request.MergeRequestID, http.MethodPost, "close", nil)
}

func (s ClientServices) ReopenMergeRequest(ctx context.Context, request MergeRequestActionRequest) (MergeRequestActionResult, error) {
	return s.mergeRequestAction(ctx, request.Organization, request.RepositoryID, request.MergeRequestID, http.MethodPost, "reopen", nil)
}

func (s ClientServices) GetMergeRequestStatus(ctx context.Context, request GetMergeRequestRequest) (MergeRequestStatusResult, error) {
	result, err := s.GetMergeRequest(ctx, request)
	if err != nil {
		return MergeRequestStatusResult{}, err
	}
	checks := []MergeCheck{}
	if result.MergeRequest.PipelineStatus != "" {
		checks = append(checks, MergeCheck{Name: "pipeline", Status: result.MergeRequest.PipelineStatus, Passed: result.MergeRequest.PipelineStatus == "success"})
	}
	return MergeRequestStatusResult{MergeRequest: result.MergeRequest, Checks: checks, Meta: result.Meta}, nil
}

func (s ClientServices) mergeRequestAction(ctx context.Context, organization string, repositoryID string, mergeRequestID string, method string, action string, body []byte) (MergeRequestActionResult, error) {
	response := mergeRequestResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: method, Path: mergeRequestRepositoryPath(organization, repositoryID, pathf("changeRequests/%s/%s", mergeRequestID, action)), Body: body}, &response)
	if err != nil {
		return MergeRequestActionResult{}, err
	}
	return MergeRequestActionResult{MergeRequest: firstMergeRequest(response.Data, response.MergeRequest, MergeRequest{ID: mergeRequestID, IID: mergeRequestID}), Action: action, MergeCommit: first(response.MergeCommit, response.MergedRevision), Meta: meta}, nil
}

func mergeRequestListPath(organization string) string {
	if organization != "" {
		return pathf("/oapi/v1/codeup/organizations/%s/changeRequests", organization)
	}
	return "/oapi/v1/codeup/changeRequests"
}

func mergeRequestRepositoryPath(organization string, repositoryID string, resource string) string {
	if organization != "" {
		return pathf("/oapi/v1/codeup/organizations/%s/repositories/%s/%s", organization, repositoryID, resource)
	}
	return pathf("/oapi/v1/codeup/repositories/%s/%s", repositoryID, resource)
}

func filterMergeRequestsByRepository(mergeRequests []MergeRequest, repositoryID string) []MergeRequest {
	if repositoryID == "" {
		return mergeRequests
	}
	filtered := make([]MergeRequest, 0, len(mergeRequests))
	for _, mergeRequest := range mergeRequests {
		if mergeRequest.ProjectID == repositoryID || mergeRequest.SourceProjectID == repositoryID || mergeRequest.TargetProjectID == repositoryID {
			filtered = append(filtered, mergeRequest)
		}
	}
	return filtered
}

func firstBool(values ...*bool) *bool {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func rawBool(value json.RawMessage) bool {
	if len(value) == 0 {
		return false
	}
	boolean := false
	if err := json.Unmarshal(value, &boolean); err == nil {
		return boolean
	}
	text := ""
	if err := json.Unmarshal(value, &text); err == nil {
		return strings.EqualFold(text, "true")
	}
	return false
}

func firstInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func firstMergeRequests(values ...[]MergeRequest) []MergeRequest {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstMergeRequestPatchSets(values ...[]mergeRequestPatchSet) []mergeRequestPatchSet {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func mergeRequestDiffSummary(files []MergeRequestFile) string {
	lines := make([]string, 0, len(files))
	for _, file := range files {
		lines = append(lines, fmt.Sprintf("%s\t+%d -%d\t%s", file.Status, file.Additions, file.Deletions, file.Path))
	}
	return strings.Join(lines, "\n")
}

func firstMergeRequest(values ...MergeRequest) MergeRequest {
	for _, value := range values {
		if value.ID != "" || value.IID != "" || value.Title != "" {
			return value
		}
	}
	return MergeRequest{}
}

func firstMergeRequestFiles(values ...[]MergeRequestFile) []MergeRequestFile {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}
