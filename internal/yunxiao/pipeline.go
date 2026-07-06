package yunxiao

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gouzil/yunxiao-cli/internal/api"
)

type PipelineService interface {
	ListPipelines(ctx context.Context, request ListPipelinesRequest) (PipelineListResult, error)
	GetPipeline(ctx context.Context, request GetPipelineRequest) (PipelineResult, error)
	RunPipeline(ctx context.Context, request RunPipelineRequest) (PipelineRunStartResult, error)
}

type RunService interface {
	ListRuns(ctx context.Context, request ListRunsRequest) (RunListResult, error)
	GetRun(ctx context.Context, request GetRunRequest) (RunResult, error)
	GetRunLog(ctx context.Context, request GetRunLogRequest) (RunLogResult, error)
	WatchRun(ctx context.Context, request WatchRunRequest) (RunResult, error)
	CancelRun(ctx context.Context, request RunActionRequest) (RunActionResult, error)
	RetryRun(ctx context.Context, request RunActionRequest) (RunActionResult, error)
	RetryTask(ctx context.Context, request RunTaskActionRequest) (RunActionResult, error)
	StopTask(ctx context.Context, request RunTaskActionRequest) (RunActionResult, error)
	SkipTask(ctx context.Context, request RunTaskActionRequest) (RunActionResult, error)
}

type Pipeline struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	Creator   string        `json:"creator"`
	UpdatedAt string        `json:"updatedAt"`
	WebURL    string        `json:"webUrl"`
	Jobs      []PipelineJob `json:"jobs,omitempty"`
}

func (p *Pipeline) UnmarshalJSON(data []byte) error {
	raw := pipelineJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = Pipeline{
		ID:        first(string(raw.ID), string(raw.PipelineID)),
		Name:      first(raw.Name, raw.PipelineName),
		Status:    raw.Status,
		Creator:   raw.Creator.Name,
		UpdatedAt: first(raw.UpdatedAt, millisTime(raw.UpdateTime), millisTime(raw.CreateTime)),
		WebURL:    raw.WebURL,
		Jobs:      raw.Jobs,
	}
	return nil
}

type pipelineJSON struct {
	ID           flexibleString `json:"id"`
	PipelineID   flexibleString `json:"pipelineId"`
	Name         string         `json:"name"`
	PipelineName string         `json:"pipelineName"`
	Status       string         `json:"status"`
	Creator      namedValue     `json:"creator"`
	UpdatedAt    string         `json:"updatedAt"`
	UpdateTime   int64          `json:"updateTime"`
	CreateTime   int64          `json:"createTime"`
	WebURL       string         `json:"webUrl"`
	Jobs         []PipelineJob  `json:"jobs,omitempty"`
}

type PipelineJob struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration"`
}

func (j *PipelineJob) UnmarshalJSON(data []byte) error {
	raw := pipelineJobJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*j = PipelineJob{
		ID:       string(raw.ID),
		Name:     raw.Name,
		Status:   raw.Status,
		Duration: raw.Duration,
	}
	return nil
}

type pipelineJobJSON struct {
	ID       flexibleString `json:"id"`
	Name     string         `json:"name"`
	Status   string         `json:"status"`
	Duration string         `json:"duration"`
}

type PipelineVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Run struct {
	ID          string        `json:"id"`
	PipelineID  string        `json:"pipelineId"`
	Pipeline    string        `json:"pipeline"`
	Branch      string        `json:"branch"`
	Status      string        `json:"status"`
	TriggerMode string        `json:"triggerMode"`
	TriggeredBy string        `json:"triggeredBy"`
	StartedAt   string        `json:"startedAt"`
	FinishedAt  string        `json:"finishedAt"`
	Duration    string        `json:"duration"`
	WebURL      string        `json:"webUrl"`
	Jobs        []PipelineJob `json:"jobs,omitempty"`
}

func (r *Run) UnmarshalJSON(data []byte) error {
	raw := runJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = Run{
		ID:          first(string(raw.ID), string(raw.PipelineRunID)),
		PipelineID:  string(raw.PipelineID),
		Pipeline:    raw.Pipeline,
		Branch:      raw.Branch,
		Status:      raw.Status,
		TriggerMode: string(raw.TriggerMode),
		TriggeredBy: raw.TriggeredBy.Name,
		StartedAt:   first(raw.StartedAt, millisTime(raw.StartTime), millisTime(raw.CreateTime)),
		FinishedAt:  first(raw.FinishedAt, millisTime(raw.EndTime), millisTime(raw.UpdateTime)),
		Duration:    raw.Duration,
		WebURL:      raw.WebURL,
		Jobs:        firstPipelineJobs(raw.Jobs, stageJobs(raw.Stages)),
	}
	return nil
}

type runJSON struct {
	ID            flexibleString `json:"id"`
	PipelineRunID flexibleString `json:"pipelineRunId"`
	PipelineID    flexibleString `json:"pipelineId"`
	Pipeline      string         `json:"pipeline"`
	Branch        string         `json:"branch"`
	Status        string         `json:"status"`
	TriggerMode   flexibleString `json:"triggerMode"`
	TriggeredBy   namedValue     `json:"triggeredBy"`
	StartedAt     string         `json:"startedAt"`
	FinishedAt    string         `json:"finishedAt"`
	StartTime     int64          `json:"startTime"`
	EndTime       int64          `json:"endTime"`
	CreateTime    int64          `json:"createTime"`
	UpdateTime    int64          `json:"updateTime"`
	Duration      string         `json:"duration"`
	WebURL        string         `json:"webUrl"`
	Jobs          []PipelineJob  `json:"jobs,omitempty"`
	Stages        []runStage     `json:"stages,omitempty"`
}

type runStage struct {
	Jobs      []PipelineJob `json:"jobs,omitempty"`
	StageInfo runStageInfo  `json:"stageInfo"`
}

type runStageInfo struct {
	Jobs []PipelineJob `json:"jobs,omitempty"`
}

type ListPipelinesRequest struct {
	Organization string      `json:"organization"`
	ProjectID    string      `json:"projectId"`
	Options      ListOptions `json:"options"`
}

type GetPipelineRequest struct {
	Organization string `json:"organization"`
	PipelineID   string `json:"pipelineId"`
}

type RunPipelineRequest struct {
	Organization string             `json:"organization"`
	PipelineID   string             `json:"pipelineId"`
	Branch       string             `json:"branch"`
	Variables    []PipelineVariable `json:"variables,omitempty"`
}

type ListRunsRequest struct {
	Organization string      `json:"organization"`
	PipelineID   string      `json:"pipelineId"`
	Branch       string      `json:"branch"`
	Status       string      `json:"status"`
	Options      ListOptions `json:"options"`
}

type GetRunRequest struct {
	Organization string `json:"organization"`
	PipelineID   string `json:"pipelineId"`
	RunID        string `json:"runId"`
}

type GetRunLogRequest struct {
	Organization string `json:"organization"`
	PipelineID   string `json:"pipelineId"`
	RunID        string `json:"runId"`
	JobID        string `json:"jobId"`
	FailedOnly   bool   `json:"failedOnly"`
}

type WatchRunRequest struct {
	Organization string        `json:"organization"`
	PipelineID   string        `json:"pipelineId"`
	RunID        string        `json:"runId"`
	Interval     time.Duration `json:"interval"`
	Timeout      time.Duration `json:"timeout"`
}

type RunActionRequest struct {
	RunID string `json:"runId"`
	Yes   bool   `json:"yes"`
}

type RunTaskActionRequest struct {
	RunID string `json:"runId"`
	JobID string `json:"jobId"`
	Yes   bool   `json:"yes"`
}

type PipelineListResult struct {
	Pipelines []Pipeline       `json:"pipelines"`
	Meta      api.ResponseMeta `json:"meta"`
}

type PipelineResult struct {
	Pipeline Pipeline         `json:"pipeline"`
	Meta     api.ResponseMeta `json:"meta"`
}

type PipelineRunStartResult struct {
	Run  Run              `json:"run"`
	Meta api.ResponseMeta `json:"meta"`
}

type RunListResult struct {
	Runs []Run            `json:"runs"`
	Meta api.ResponseMeta `json:"meta"`
}

type RunResult struct {
	Run  Run              `json:"run"`
	Meta api.ResponseMeta `json:"meta"`
}

type RunLogResult struct {
	RunID   string           `json:"runId"`
	JobID   string           `json:"jobId"`
	JobName string           `json:"jobName"`
	Lines   string           `json:"lines"`
	Meta    api.ResponseMeta `json:"meta"`
}

type RunActionResult struct {
	RunID     string           `json:"runId"`
	JobID     string           `json:"jobId"`
	Action    string           `json:"action"`
	NewStatus string           `json:"newStatus"`
	Meta      api.ResponseMeta `json:"meta"`
}

type pipelineListResponse struct {
	Data    []Pipeline `json:"data"`
	Result  []Pipeline `json:"result"`
	Items   []Pipeline `json:"items"`
	Records []Pipeline `json:"records"`
}

func (r *pipelineListResponse) UnmarshalJSON(data []byte) error {
	pipelines := []Pipeline{}
	if err := json.Unmarshal(data, &pipelines); err == nil {
		r.Data = pipelines
		return nil
	}
	type wrapped pipelineListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = pipelineListResponse(value)
	return nil
}

type pipelineResponse struct {
	Data     Pipeline `json:"data"`
	Pipeline Pipeline `json:"pipeline"`
}

func (r *pipelineResponse) UnmarshalJSON(data []byte) error {
	pipeline := Pipeline{}
	if err := json.Unmarshal(data, &pipeline); err == nil && (pipeline.ID != "" || pipeline.Name != "") {
		r.Data = pipeline
		return nil
	}
	type wrapped pipelineResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = pipelineResponse(value)
	return nil
}

type runListResponse struct {
	Data    []Run `json:"data"`
	Result  []Run `json:"result"`
	Items   []Run `json:"items"`
	Records []Run `json:"records"`
}

func (r *runListResponse) UnmarshalJSON(data []byte) error {
	runs := []Run{}
	if err := json.Unmarshal(data, &runs); err == nil {
		r.Data = runs
		return nil
	}
	type wrapped runListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = runListResponse(value)
	return nil
}

type runResponse struct {
	Data Run `json:"data"`
	Run  Run `json:"run"`
}

func (r *runResponse) UnmarshalJSON(data []byte) error {
	run := Run{}
	if err := json.Unmarshal(data, &run); err == nil && (run.ID != "" || run.Status != "") {
		r.Data = run
		return nil
	}
	type wrapped runResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = runResponse(value)
	return nil
}

type runLogResponse struct {
	Data    string `json:"data"`
	Log     string `json:"log"`
	Content string `json:"content"`
}

func (s ClientServices) ListPipelines(ctx context.Context, request ListPipelinesRequest) (PipelineListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listPipelinesForAllOrganizations(ctx, request)
	}
	return s.listPipelinesForOrganization(ctx, request)
}

func (s ClientServices) listPipelinesForAllOrganizations(ctx context.Context, request ListPipelinesRequest) (PipelineListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return PipelineListResult{}, err
	}
	pipelines := []Pipeline{}
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listPipelinesForOrganization(ctx, ListPipelinesRequest{Organization: organization.ID, ProjectID: request.ProjectID, Options: request.Options})
		if err != nil {
			return PipelineListResult{}, err
		}
		pipelines = append(pipelines, result.Pipelines...)
		meta = result.Meta
	}
	return PipelineListResult{Pipelines: pipelines, Meta: meta}, nil
}

func (s ClientServices) listPipelinesForOrganization(ctx context.Context, request ListPipelinesRequest) (PipelineListResult, error) {
	params := queryParams(request.Options)
	params = addParam(params, "projectId", request.ProjectID)
	response := pipelineListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: flowPipelinePath(request.Organization, "", ""), Query: params}, &response)
	if err != nil {
		return PipelineListResult{}, err
	}
	return PipelineListResult{Pipelines: firstPipelines(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) GetPipeline(ctx context.Context, request GetPipelineRequest) (PipelineResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.getPipelineForAllOrganizations(ctx, request)
	}
	return s.getPipelineForOrganization(ctx, request)
}

func (s ClientServices) getPipelineForAllOrganizations(ctx context.Context, request GetPipelineRequest) (PipelineResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return PipelineResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.getPipelineForOrganization(ctx, GetPipelineRequest{Organization: organization.ID, PipelineID: request.PipelineID})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return PipelineResult{}, err
	}
	if lastErr != nil {
		return PipelineResult{}, lastErr
	}
	return PipelineResult{Meta: meta}, nil
}

func (s ClientServices) getPipelineForOrganization(ctx context.Context, request GetPipelineRequest) (PipelineResult, error) {
	response := pipelineResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: flowPipelinePath(request.Organization, request.PipelineID, "")}, &response)
	if err != nil {
		return PipelineResult{}, err
	}
	return PipelineResult{Pipeline: firstPipeline(response.Data, response.Pipeline), Meta: meta}, nil
}

func (s ClientServices) RunPipeline(ctx context.Context, request RunPipelineRequest) (PipelineRunStartResult, error) {
	body, err := api.EncodeJSONBody(request)
	if err != nil {
		return PipelineRunStartResult{}, err
	}
	response := runResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPost, Path: flowPipelinePath(request.Organization, request.PipelineID, "runs"), Body: body}, &response)
	if err != nil {
		return PipelineRunStartResult{}, err
	}
	return PipelineRunStartResult{Run: firstRun(response.Data, response.Run), Meta: meta}, nil
}

func (s ClientServices) ListRuns(ctx context.Context, request ListRunsRequest) (RunListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listRunsForAllOrganizations(ctx, request)
	}
	return s.listRunsForOrganization(ctx, request)
}

func (s ClientServices) listRunsForAllOrganizations(ctx context.Context, request ListRunsRequest) (RunListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return RunListResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listRunsForOrganization(ctx, ListRunsRequest{
			Organization: organization.ID,
			PipelineID:   request.PipelineID,
			Branch:       request.Branch,
			Status:       request.Status,
			Options:      request.Options,
		})
		if err == nil {
			return result, nil
		}
		var apiErr *api.Error
		if errors.As(err, &apiErr) && apiErr.HTTPStatus == http.StatusNotFound {
			lastErr = err
			continue
		}
		return RunListResult{}, err
	}
	if lastErr != nil {
		return RunListResult{}, lastErr
	}
	return RunListResult{Meta: meta}, nil
}

func (s ClientServices) listRunsForOrganization(ctx context.Context, request ListRunsRequest) (RunListResult, error) {
	params := queryParams(request.Options)
	params = addParam(params, "branch", request.Branch)
	params = addParam(params, "status", request.Status)
	response := runListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: flowPipelinePath(request.Organization, request.PipelineID, "runs"), Query: params}, &response)
	if err != nil {
		return RunListResult{}, err
	}
	return RunListResult{Runs: firstRuns(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func flowPipelinePath(organization string, pipelineID string, resource string) string {
	if strings.TrimSpace(organization) != "" {
		if strings.TrimSpace(pipelineID) != "" {
			if strings.TrimSpace(resource) != "" {
				return pathf("/oapi/v1/flow/organizations/%s/pipelines/%s/%s", organization, pipelineID, resource)
			}
			return pathf("/oapi/v1/flow/organizations/%s/pipelines/%s", organization, pipelineID)
		}
		return pathf("/oapi/v1/flow/organizations/%s/pipelines", organization)
	}
	if strings.TrimSpace(pipelineID) != "" {
		if strings.TrimSpace(resource) != "" {
			return pathf("/oapi/v1/flow/pipelines/%s/%s", pipelineID, resource)
		}
		return pathf("/oapi/v1/flow/pipelines/%s", pipelineID)
	}
	return "/oapi/v1/flow/pipelines"
}

func (s ClientServices) GetRun(ctx context.Context, request GetRunRequest) (RunResult, error) {
	location, err := s.findRun(ctx, request)
	if err != nil {
		return RunResult{}, err
	}
	return location.Result, nil
}

type runLocation struct {
	Organization string
	PipelineID   string
	Result       RunResult
}

func (s ClientServices) findRun(ctx context.Context, request GetRunRequest) (runLocation, error) {
	if strings.TrimSpace(request.Organization) != "" {
		return s.findRunInOrganization(ctx, request.Organization, request)
	}
	organizations, _, err := s.listOrganizations(ctx)
	if err != nil {
		return runLocation{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		location, err := s.findRunInOrganization(ctx, organization.ID, request)
		if err == nil {
			return location, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return runLocation{}, err
	}
	if lastErr != nil {
		return runLocation{}, lastErr
	}
	return runLocation{}, notFoundError()
}

func (s ClientServices) findRunInOrganization(ctx context.Context, organization string, request GetRunRequest) (runLocation, error) {
	if strings.TrimSpace(request.PipelineID) != "" {
		next := request
		next.Organization = organization
		result, err := s.getRunForOrganizationPipeline(ctx, next)
		if err != nil {
			return runLocation{}, err
		}
		return runLocation{Organization: organization, PipelineID: first(result.Run.PipelineID, request.PipelineID), Result: result}, nil
	}
	pipelines, err := s.listPipelinesForOrganization(ctx, ListPipelinesRequest{Organization: organization, Options: ListOptions{PerPage: 100}})
	if err != nil {
		return runLocation{}, err
	}
	var lastErr error
	for _, pipeline := range pipelines.Pipelines {
		if pipeline.ID == "" {
			continue
		}
		next := request
		next.Organization = organization
		next.PipelineID = pipeline.ID
		result, err := s.getRunForOrganizationPipeline(ctx, next)
		if err == nil {
			return runLocation{Organization: organization, PipelineID: first(result.Run.PipelineID, pipeline.ID), Result: result}, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return runLocation{}, err
	}
	if lastErr != nil {
		return runLocation{}, lastErr
	}
	return runLocation{}, notFoundError()
}

func (s ClientServices) getRunForOrganizationPipeline(ctx context.Context, request GetRunRequest) (RunResult, error) {
	response := runResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: flowPipelinePath(request.Organization, request.PipelineID, pathf("runs/%s", request.RunID))}, &response)
	if err != nil {
		return RunResult{}, err
	}
	return RunResult{Run: firstRun(response.Data, response.Run), Meta: meta}, nil
}

func (s ClientServices) GetRunLog(ctx context.Context, request GetRunLogRequest) (RunLogResult, error) {
	organization := request.Organization
	pipelineID := request.PipelineID
	jobID := request.JobID
	if strings.TrimSpace(organization) == "" || strings.TrimSpace(pipelineID) == "" || strings.TrimSpace(jobID) == "" {
		location, err := s.findRun(ctx, GetRunRequest{Organization: request.Organization, PipelineID: request.PipelineID, RunID: request.RunID})
		if err != nil {
			return RunLogResult{}, err
		}
		organization = location.Organization
		pipelineID = location.PipelineID
		if strings.TrimSpace(jobID) == "" {
			jobID = firstJobID(location.Result.Run.Jobs)
		}
	}
	params := []api.QueryParam{}
	if request.FailedOnly {
		params = addParam(params, "failedOnly", "true")
	}
	response := runLogResponse{}
	logPath := flowPipelinePath(organization, pipelineID, pathf("runs/%s/log", request.RunID))
	if strings.TrimSpace(jobID) != "" {
		logPath = flowPipelinePath(organization, pipelineID, pathf("runs/%s/job/%s/log", request.RunID, jobID))
	}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: logPath, Query: params}, &response)
	if err != nil {
		return RunLogResult{}, err
	}
	return RunLogResult{RunID: request.RunID, JobID: jobID, Lines: first(response.Log, response.Data, response.Content), Meta: meta}, nil
}

func (s ClientServices) WatchRun(ctx context.Context, request WatchRunRequest) (RunResult, error) {
	interval := request.Interval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ctxToUse := ctx
	cancel := func() {}
	if request.Timeout > 0 {
		ctxToUse, cancel = context.WithTimeout(ctx, request.Timeout)
	}
	defer cancel()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	next := GetRunRequest{Organization: request.Organization, PipelineID: request.PipelineID, RunID: request.RunID}
	for {
		location, err := s.findRun(ctxToUse, next)
		if err != nil {
			return RunResult{}, err
		}
		result := location.Result
		next.Organization = location.Organization
		next.PipelineID = location.PipelineID
		if terminalRunStatus(result.Run.Status) {
			return result, nil
		}
		select {
		case <-ctxToUse.Done():
			return result, ctxToUse.Err()
		case <-ticker.C:
		}
	}
}

func (s ClientServices) CancelRun(ctx context.Context, request RunActionRequest) (RunActionResult, error) {
	return s.runAction(ctx, request.RunID, "", "cancel")
}

func (s ClientServices) RetryRun(ctx context.Context, request RunActionRequest) (RunActionResult, error) {
	return s.runAction(ctx, request.RunID, "", "retry")
}

func (s ClientServices) RetryTask(ctx context.Context, request RunTaskActionRequest) (RunActionResult, error) {
	return s.runAction(ctx, request.RunID, request.JobID, "retry-task")
}

func (s ClientServices) StopTask(ctx context.Context, request RunTaskActionRequest) (RunActionResult, error) {
	return s.runAction(ctx, request.RunID, request.JobID, "stop-task")
}

func (s ClientServices) SkipTask(ctx context.Context, request RunTaskActionRequest) (RunActionResult, error) {
	return s.runAction(ctx, request.RunID, request.JobID, "skip-task")
}

func (s ClientServices) runAction(ctx context.Context, runID string, jobID string, action string) (RunActionResult, error) {
	payload := RunActionResult{RunID: runID, JobID: jobID, Action: action}
	body, err := api.EncodeJSONBody(payload)
	if err != nil {
		return RunActionResult{}, err
	}
	path := pathf("/oapi/v1/flow/runs/%s/%s", runID, action)
	if jobID != "" {
		path = pathf("/oapi/v1/flow/runs/%s/jobs/%s/%s", runID, jobID, action)
	}
	response := runResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPost, Path: path, Body: body}, &response)
	if err != nil {
		return RunActionResult{}, err
	}
	return RunActionResult{RunID: runID, JobID: jobID, Action: action, NewStatus: firstRun(response.Data, response.Run).Status, Meta: meta}, nil
}

func terminalRunStatus(status string) bool {
	switch strings.ToLower(status) {
	case "success", "failed", "canceled", "cancelled", "skipped", "error":
		return true
	default:
		return false
	}
}

func isNotFound(err error) bool {
	var apiErr *api.Error
	return errors.As(err, &apiErr) && apiErr.HTTPStatus == http.StatusNotFound
}

func notFoundError() error {
	return &api.Error{Summary: "Yunxiao API request failed", HTTPStatus: http.StatusNotFound, Code: "NotFound", Message: "Not Found"}
}

func firstPipelines(values ...[]Pipeline) []Pipeline {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstPipeline(values ...Pipeline) Pipeline {
	for _, value := range values {
		if value.ID != "" || value.Name != "" {
			return value
		}
	}
	return Pipeline{}
}

func firstPipelineJobs(values ...[]PipelineJob) []PipelineJob {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func stageJobs(stages []runStage) []PipelineJob {
	jobs := []PipelineJob{}
	for _, stage := range stages {
		jobs = append(jobs, stage.Jobs...)
		jobs = append(jobs, stage.StageInfo.Jobs...)
	}
	return jobs
}

func firstRuns(values ...[]Run) []Run {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstJobID(jobs []PipelineJob) string {
	for _, job := range jobs {
		if strings.TrimSpace(job.ID) != "" {
			return job.ID
		}
	}
	return ""
}

func firstRun(values ...Run) Run {
	for _, value := range values {
		if value.ID != "" || value.Status != "" {
			return value
		}
	}
	return Run{}
}
