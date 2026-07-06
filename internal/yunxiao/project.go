package yunxiao

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gouzi/yunxiao-cli/internal/api"
)

type ProjectService interface {
	ListProjects(ctx context.Context, request ListProjectsRequest) (ProjectListResult, error)
	GetProject(ctx context.Context, request GetProjectRequest) (ProjectResult, error)
	ListProjectMembers(ctx context.Context, request ListProjectMembersRequest) (ProjectMemberListResult, error)
	ListIterations(ctx context.Context, request ListIterationsRequest) (IterationListResult, error)
	GetIteration(ctx context.Context, request GetIterationRequest) (IterationResult, error)
	ListMilestones(ctx context.Context, request ListMilestonesRequest) (MilestoneListResult, error)
	GetMilestone(ctx context.Context, request GetMilestoneRequest) (MilestoneResult, error)
	ListLabels(ctx context.Context, request ListLabelsRequest) (LabelListResult, error)
}

type WorkItemService interface {
	ListWorkItems(ctx context.Context, request ListWorkItemsRequest) (WorkItemListResult, error)
	GetWorkItem(ctx context.Context, request GetWorkItemRequest) (WorkItemResult, error)
	CreateWorkItem(ctx context.Context, request CreateWorkItemRequest) (WorkItemResult, error)
	UpdateWorkItem(ctx context.Context, request UpdateWorkItemRequest) (WorkItemResult, error)
	DeleteWorkItem(ctx context.Context, request DeleteWorkItemRequest) (WorkItemActionResult, error)
	ListWorkItemActivities(ctx context.Context, request ListWorkItemActivitiesRequest) (WorkItemActivityListResult, error)
}

type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	Owner          string `json:"owner"`
	MemberCount    int    `json:"memberCount"`
	IterationCount int    `json:"iterationCount"`
	WebURL         string `json:"webUrl"`
	UpdatedAt      string `json:"updatedAt"`
}

func (p *Project) UnmarshalJSON(data []byte) error {
	raw := projectJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = Project{
		ID:             string(raw.ID),
		Name:           raw.Name,
		Status:         first(raw.Status.Name, raw.Status.Value),
		Owner:          first(raw.Owner, raw.Creator.Name),
		MemberCount:    raw.MemberCount,
		IterationCount: raw.IterationCount,
		WebURL:         raw.WebURL,
		UpdatedAt:      first(raw.UpdatedAt, millisTime(raw.GMTModified), millisTime(raw.GMTCreate)),
	}
	return nil
}

type projectJSON struct {
	ID             flexibleString `json:"id"`
	Name           string         `json:"name"`
	Status         namedValue     `json:"status"`
	Owner          string         `json:"owner"`
	Creator        namedValue     `json:"creator"`
	MemberCount    int            `json:"memberCount"`
	IterationCount int            `json:"iterationCount"`
	WebURL         string         `json:"webUrl"`
	UpdatedAt      string         `json:"updatedAt"`
	GMTModified    int64          `json:"gmtModified"`
	GMTCreate      int64          `json:"gmtCreate"`
}

type namedValue struct {
	ID    flexibleString `json:"id"`
	Name  string         `json:"name"`
	Value string         `json:"value"`
}

func (v *namedValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	text := ""
	if err := json.Unmarshal(data, &text); err == nil {
		v.Name = text
		return nil
	}
	type object namedValue
	value := object{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = namedValue(value)
	return nil
}

func millisTime(value int64) string {
	if value <= 0 {
		return ""
	}
	return time.UnixMilli(value).Format(time.RFC3339)
}

func millisStringTime(value flexibleString) string {
	parsed, err := strconv.ParseInt(string(value), 10, 64)
	if err != nil {
		return string(value)
	}
	return millisTime(parsed)
}

type ProjectMember struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Joined string `json:"joined"`
}

func (m *ProjectMember) UnmarshalJSON(data []byte) error {
	raw := projectMemberJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*m = ProjectMember{
		UserID: string(raw.UserID),
		Name:   first(raw.Name, raw.UserName),
		Role:   first(raw.Role, raw.RoleName),
		Joined: first(raw.Joined, millisTime(raw.GMTCreate)),
	}
	return nil
}

type projectMemberJSON struct {
	UserID    flexibleString `json:"userId"`
	Name      string         `json:"name"`
	UserName  string         `json:"userName"`
	Role      string         `json:"role"`
	RoleName  string         `json:"roleName"`
	Joined    string         `json:"joined"`
	GMTCreate int64          `json:"gmtCreate"`
}

type Iteration struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type Milestone struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	DueDate string `json:"dueDate"`
}

type Label struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

type WorkItem struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	TypeID      string `json:"typeId"`
	State       string `json:"state"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Assignee    string `json:"assignee"`
	Reporter    string `json:"reporter"`
	ProjectID   string `json:"projectId"`
	ProjectName string `json:"projectName"`
	Iteration   string `json:"iteration"`
	Priority    string `json:"priority"`
	UpdatedAt   string `json:"updatedAt"`
	WebURL      string `json:"webUrl"`
}

func (w *WorkItem) UnmarshalJSON(data []byte) error {
	raw := workItemJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*w = WorkItem{
		ID:          first(string(raw.ID), raw.SerialNumber),
		Type:        first(raw.Type, raw.WorkItemType.Name, raw.CategoryID),
		TypeID:      string(raw.WorkItemType.ID),
		State:       first(raw.State, raw.Status.Name),
		Title:       first(raw.Title, raw.Subject),
		Description: first(raw.Description, raw.Body),
		Assignee:    first(raw.Assignee, raw.AssignedTo.Name),
		Reporter:    first(raw.Reporter, raw.Creator.Name),
		ProjectID:   first(raw.ProjectID, string(raw.Space.ID)),
		ProjectName: first(raw.ProjectName, raw.Space.Name),
		Iteration:   first(raw.Iteration, raw.Sprint.Name),
		Priority:    first(raw.Priority, raw.PriorityValue.Name),
		UpdatedAt:   first(raw.UpdatedAt, millisTime(raw.GMTModified)),
		WebURL:      raw.WebURL,
	}
	return nil
}

type workItemJSON struct {
	ID            flexibleString `json:"id"`
	SerialNumber  string         `json:"serialNumber"`
	Type          string         `json:"type"`
	CategoryID    string         `json:"categoryId"`
	WorkItemType  namedValue     `json:"workitemType"`
	State         string         `json:"state"`
	Status        namedValue     `json:"status"`
	Title         string         `json:"title"`
	Subject       string         `json:"subject"`
	Description   string         `json:"description"`
	Body          string         `json:"body"`
	Assignee      string         `json:"assignee"`
	AssignedTo    namedValue     `json:"assignedTo"`
	Reporter      string         `json:"reporter"`
	Creator       namedValue     `json:"creator"`
	ProjectID     string         `json:"projectId"`
	ProjectName   string         `json:"projectName"`
	Space         idNameValue    `json:"space"`
	Iteration     string         `json:"iteration"`
	Sprint        namedValue     `json:"sprint"`
	Priority      string         `json:"priority"`
	PriorityValue namedValue     `json:"priorityValue"`
	UpdatedAt     string         `json:"updatedAt"`
	GMTModified   int64          `json:"gmtModified"`
	WebURL        string         `json:"webUrl"`
}

type idNameValue struct {
	ID   flexibleString `json:"id"`
	Name string         `json:"name"`
}

type WorkItemActivity struct {
	Time   string `json:"time"`
	Actor  string `json:"actor"`
	Action string `json:"action"`
}

func (a *WorkItemActivity) UnmarshalJSON(data []byte) error {
	raw := workItemActivityJSON{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*a = WorkItemActivity{
		Time:   first(raw.Time, millisStringTime(raw.EventTime)),
		Actor:  first(raw.Actor, raw.Operator.Name),
		Action: first(raw.Action, raw.ActionType, raw.EventType),
	}
	return nil
}

type workItemActivityJSON struct {
	Time       string         `json:"time"`
	EventTime  flexibleString `json:"eventTime"`
	Actor      string         `json:"actor"`
	Operator   namedValue     `json:"operator"`
	Action     string         `json:"action"`
	ActionType string         `json:"actionType"`
	EventType  string         `json:"eventType"`
}

type ListProjectsRequest struct {
	Organization string      `json:"organization"`
	Options      ListOptions `json:"options"`
}

type GetProjectRequest struct {
	Organization string `json:"organization"`
	ProjectID    string `json:"projectId"`
}

type ListProjectMembersRequest struct {
	Organization string      `json:"organization"`
	ProjectID    string      `json:"projectId"`
	Options      ListOptions `json:"options"`
}

type ListIterationsRequest struct {
	Organization string      `json:"organization"`
	ProjectID    string      `json:"projectId"`
	Options      ListOptions `json:"options"`
}

type GetIterationRequest struct {
	Organization string `json:"organization"`
	ProjectID    string `json:"projectId"`
	IterationID  string `json:"iterationId"`
}

type ListMilestonesRequest struct {
	Organization string      `json:"organization"`
	ProjectID    string      `json:"projectId"`
	Options      ListOptions `json:"options"`
}

type GetMilestoneRequest struct {
	Organization string `json:"organization"`
	ProjectID    string `json:"projectId"`
	MilestoneID  string `json:"milestoneId"`
}

type ListLabelsRequest struct {
	Organization string      `json:"organization"`
	ProjectID    string      `json:"projectId"`
	Options      ListOptions `json:"options"`
}

type ListWorkItemsRequest struct {
	Organization string      `json:"organization"`
	ProjectID    string      `json:"projectId"`
	Category     string      `json:"category"`
	State        string      `json:"state"`
	Assignee     string      `json:"assignee"`
	Options      ListOptions `json:"options"`
}

type GetWorkItemRequest struct {
	Organization string `json:"organization"`
	ID           string `json:"id"`
}

type CreateWorkItemRequest struct {
	Organization string `json:"organization"`
	ProjectID    string `json:"projectId"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	Assignee     string `json:"assignee"`
}

type UpdateWorkItemRequest struct {
	Organization string `json:"organization"`
	ID           string `json:"id"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	State        string `json:"state"`
	Assignee     string `json:"assignee"`
}

type DeleteWorkItemRequest struct {
	Organization string `json:"organization"`
	ID           string `json:"id"`
	Yes          bool   `json:"yes"`
}

type ListWorkItemActivitiesRequest struct {
	Organization string      `json:"organization"`
	ID           string      `json:"id"`
	Options      ListOptions `json:"options"`
}

type ProjectListResult struct {
	Projects []Project        `json:"projects"`
	Meta     api.ResponseMeta `json:"meta"`
}

type ProjectResult struct {
	Project Project          `json:"project"`
	Meta    api.ResponseMeta `json:"meta"`
}

type ProjectMemberListResult struct {
	Members []ProjectMember  `json:"members"`
	Meta    api.ResponseMeta `json:"meta"`
}

type IterationListResult struct {
	Iterations []Iteration      `json:"iterations"`
	Meta       api.ResponseMeta `json:"meta"`
}

type IterationResult struct {
	Iteration Iteration        `json:"iteration"`
	Meta      api.ResponseMeta `json:"meta"`
}

type MilestoneListResult struct {
	Milestones []Milestone      `json:"milestones"`
	Meta       api.ResponseMeta `json:"meta"`
}

type MilestoneResult struct {
	Milestone Milestone        `json:"milestone"`
	Meta      api.ResponseMeta `json:"meta"`
}

type LabelListResult struct {
	Labels []Label          `json:"labels"`
	Meta   api.ResponseMeta `json:"meta"`
}

type WorkItemListResult struct {
	WorkItems []WorkItem       `json:"workItems"`
	Meta      api.ResponseMeta `json:"meta"`
}

type WorkItemResult struct {
	WorkItem WorkItem         `json:"workItem"`
	Meta     api.ResponseMeta `json:"meta"`
}

type WorkItemActionResult struct {
	ID   string           `json:"id"`
	Meta api.ResponseMeta `json:"meta"`
}

type WorkItemActivityListResult struct {
	Activities []WorkItemActivity `json:"activities"`
	Meta       api.ResponseMeta   `json:"meta"`
}

type projectListResponse struct {
	Data    []Project `json:"data"`
	Result  []Project `json:"result"`
	Items   []Project `json:"items"`
	Records []Project `json:"records"`
}

func (r *projectListResponse) UnmarshalJSON(data []byte) error {
	projects := []Project{}
	if err := json.Unmarshal(data, &projects); err == nil {
		r.Data = projects
		return nil
	}
	type wrapped projectListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = projectListResponse(value)
	return nil
}

type projectResponse struct {
	Data    Project `json:"data"`
	Project Project `json:"project"`
}

func (r *projectResponse) UnmarshalJSON(data []byte) error {
	project := Project{}
	if err := json.Unmarshal(data, &project); err == nil && (project.ID != "" || project.Name != "") {
		r.Data = project
		return nil
	}
	type wrapped projectResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = projectResponse(value)
	return nil
}

type projectMemberListResponse struct {
	Data    []ProjectMember `json:"data"`
	Result  []ProjectMember `json:"result"`
	Items   []ProjectMember `json:"items"`
	Records []ProjectMember `json:"records"`
}

func (r *projectMemberListResponse) UnmarshalJSON(data []byte) error {
	members := []ProjectMember{}
	if err := json.Unmarshal(data, &members); err == nil {
		r.Data = members
		return nil
	}
	type wrapped projectMemberListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = projectMemberListResponse(value)
	return nil
}

type iterationListResponse struct {
	Data    []Iteration `json:"data"`
	Result  []Iteration `json:"result"`
	Items   []Iteration `json:"items"`
	Records []Iteration `json:"records"`
}

func (r *iterationListResponse) UnmarshalJSON(data []byte) error {
	iterations := []Iteration{}
	if err := json.Unmarshal(data, &iterations); err == nil {
		r.Data = iterations
		return nil
	}
	type wrapped iterationListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = iterationListResponse(value)
	return nil
}

type iterationResponse struct {
	Data      Iteration `json:"data"`
	Iteration Iteration `json:"iteration"`
}

func (r *iterationResponse) UnmarshalJSON(data []byte) error {
	iteration := Iteration{}
	if err := json.Unmarshal(data, &iteration); err == nil && (iteration.ID != "" || iteration.Name != "") {
		r.Data = iteration
		return nil
	}
	type wrapped iterationResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = iterationResponse(value)
	return nil
}

type milestoneListResponse struct {
	Data    []Milestone `json:"data"`
	Result  []Milestone `json:"result"`
	Items   []Milestone `json:"items"`
	Records []Milestone `json:"records"`
}

func (r *milestoneListResponse) UnmarshalJSON(data []byte) error {
	milestones := []Milestone{}
	if err := json.Unmarshal(data, &milestones); err == nil {
		r.Data = milestones
		return nil
	}
	type wrapped milestoneListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = milestoneListResponse(value)
	return nil
}

type milestoneResponse struct {
	Data      Milestone `json:"data"`
	Milestone Milestone `json:"milestone"`
}

func (r *milestoneResponse) UnmarshalJSON(data []byte) error {
	milestone := Milestone{}
	if err := json.Unmarshal(data, &milestone); err == nil && (milestone.ID != "" || milestone.Name != "") {
		r.Data = milestone
		return nil
	}
	type wrapped milestoneResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = milestoneResponse(value)
	return nil
}

type labelListResponse struct {
	Data    []Label `json:"data"`
	Result  []Label `json:"result"`
	Items   []Label `json:"items"`
	Records []Label `json:"records"`
}

func (r *labelListResponse) UnmarshalJSON(data []byte) error {
	labels := []Label{}
	if err := json.Unmarshal(data, &labels); err == nil {
		r.Data = labels
		return nil
	}
	type wrapped labelListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = labelListResponse(value)
	return nil
}

type workItemListResponse struct {
	Data    []WorkItem `json:"data"`
	Result  []WorkItem `json:"result"`
	Items   []WorkItem `json:"items"`
	Records []WorkItem `json:"records"`
}

func (r *workItemListResponse) UnmarshalJSON(data []byte) error {
	workItems := []WorkItem{}
	if err := json.Unmarshal(data, &workItems); err == nil {
		r.Data = workItems
		return nil
	}
	type wrapped workItemListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = workItemListResponse(value)
	return nil
}

type workItemResponse struct {
	Data     WorkItem `json:"data"`
	WorkItem WorkItem `json:"workItem"`
}

func (r *workItemResponse) UnmarshalJSON(data []byte) error {
	workItem := WorkItem{}
	if err := json.Unmarshal(data, &workItem); err == nil && (workItem.ID != "" || workItem.Title != "") {
		r.Data = workItem
		return nil
	}
	type wrapped workItemResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = workItemResponse(value)
	return nil
}

type workItemActivityListResponse struct {
	Data    []WorkItemActivity `json:"data"`
	Result  []WorkItemActivity `json:"result"`
	Items   []WorkItemActivity `json:"items"`
	Records []WorkItemActivity `json:"records"`
}

func (r *workItemActivityListResponse) UnmarshalJSON(data []byte) error {
	activities := []WorkItemActivity{}
	if err := json.Unmarshal(data, &activities); err == nil {
		r.Data = activities
		return nil
	}
	type wrapped workItemActivityListResponse
	value := wrapped{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = workItemActivityListResponse(value)
	return nil
}

func (s ClientServices) ListProjects(ctx context.Context, request ListProjectsRequest) (ProjectListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listProjectsForAllOrganizations(ctx, request)
	}
	return s.listProjectsForOrganization(ctx, request)
}

func (s ClientServices) listProjectsForAllOrganizations(ctx context.Context, request ListProjectsRequest) (ProjectListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return ProjectListResult{}, err
	}
	projects := []Project{}
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listProjectsForOrganization(ctx, ListProjectsRequest{Organization: organization.ID, Options: request.Options})
		if err != nil {
			return ProjectListResult{}, err
		}
		projects = append(projects, result.Projects...)
		meta = result.Meta
	}
	return ProjectListResult{Projects: projects, Meta: meta}, nil
}

func (s ClientServices) listProjectsForOrganization(ctx context.Context, request ListProjectsRequest) (ProjectListResult, error) {
	params := queryParams(request.Options)
	body, err := api.EncodeJSONBody(map[string]any{})
	if err != nil {
		return ProjectListResult{}, err
	}
	response := projectListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPost, Path: pathf("/oapi/v1/projex/organizations/%s/projects:search", request.Organization), Query: params, Body: body}, &response)
	if err != nil {
		return ProjectListResult{}, err
	}
	return ProjectListResult{Projects: firstProjects(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) GetProject(ctx context.Context, request GetProjectRequest) (ProjectResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.getProjectForAllOrganizations(ctx, request)
	}
	return s.getProjectForOrganization(ctx, request)
}

func (s ClientServices) getProjectForAllOrganizations(ctx context.Context, request GetProjectRequest) (ProjectResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return ProjectResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.getProjectForOrganization(ctx, GetProjectRequest{Organization: organization.ID, ProjectID: request.ProjectID})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return ProjectResult{}, err
	}
	if lastErr != nil {
		return ProjectResult{}, lastErr
	}
	return ProjectResult{Meta: meta}, nil
}

func (s ClientServices) getProjectForOrganization(ctx context.Context, request GetProjectRequest) (ProjectResult, error) {
	response := projectResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: projectPath(request.Organization, request.ProjectID, "")}, &response)
	if err != nil {
		return ProjectResult{}, err
	}
	return ProjectResult{Project: firstProject(response.Data, response.Project), Meta: meta}, nil
}

func (s ClientServices) ListProjectMembers(ctx context.Context, request ListProjectMembersRequest) (ProjectMemberListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listProjectMembersForAllOrganizations(ctx, request)
	}
	return s.listProjectMembersForOrganization(ctx, request)
}

func (s ClientServices) listProjectMembersForAllOrganizations(ctx context.Context, request ListProjectMembersRequest) (ProjectMemberListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return ProjectMemberListResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listProjectMembersForOrganization(ctx, ListProjectMembersRequest{Organization: organization.ID, ProjectID: request.ProjectID, Options: request.Options})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return ProjectMemberListResult{}, err
	}
	if lastErr != nil {
		return ProjectMemberListResult{}, lastErr
	}
	return ProjectMemberListResult{Meta: meta}, nil
}

func (s ClientServices) listProjectMembersForOrganization(ctx context.Context, request ListProjectMembersRequest) (ProjectMemberListResult, error) {
	response := projectMemberListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: projectPath(request.Organization, request.ProjectID, "members"), Query: queryParams(request.Options)}, &response)
	if err != nil {
		return ProjectMemberListResult{}, err
	}
	return ProjectMemberListResult{Members: firstProjectMembers(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) ListIterations(ctx context.Context, request ListIterationsRequest) (IterationListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listIterationsForAllOrganizations(ctx, request)
	}
	return s.listIterationsForOrganization(ctx, request)
}

func (s ClientServices) listIterationsForAllOrganizations(ctx context.Context, request ListIterationsRequest) (IterationListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return IterationListResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listIterationsForOrganization(ctx, ListIterationsRequest{Organization: organization.ID, ProjectID: request.ProjectID, Options: request.Options})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return IterationListResult{}, err
	}
	if lastErr != nil {
		return IterationListResult{}, lastErr
	}
	return IterationListResult{Meta: meta}, nil
}

func (s ClientServices) listIterationsForOrganization(ctx context.Context, request ListIterationsRequest) (IterationListResult, error) {
	response := iterationListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: projectPath(request.Organization, request.ProjectID, "sprints"), Query: queryParams(request.Options)}, &response)
	if err != nil {
		return IterationListResult{}, err
	}
	return IterationListResult{Iterations: firstIterations(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) GetIteration(ctx context.Context, request GetIterationRequest) (IterationResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.getIterationForAllOrganizations(ctx, request)
	}
	return s.getIterationForOrganization(ctx, request)
}

func (s ClientServices) getIterationForAllOrganizations(ctx context.Context, request GetIterationRequest) (IterationResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return IterationResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.getIterationForOrganization(ctx, GetIterationRequest{Organization: organization.ID, ProjectID: request.ProjectID, IterationID: request.IterationID})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return IterationResult{}, err
	}
	if lastErr != nil {
		return IterationResult{}, lastErr
	}
	return IterationResult{Meta: meta}, nil
}

func (s ClientServices) getIterationForOrganization(ctx context.Context, request GetIterationRequest) (IterationResult, error) {
	response := iterationResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: projectPath(request.Organization, request.ProjectID, pathf("sprints/%s", request.IterationID))}, &response)
	if err != nil {
		return IterationResult{}, err
	}
	return IterationResult{Iteration: firstIteration(response.Data, response.Iteration), Meta: meta}, nil
}

func (s ClientServices) ListMilestones(ctx context.Context, request ListMilestonesRequest) (MilestoneListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listMilestonesForAllOrganizations(ctx, request)
	}
	return s.listMilestonesForOrganization(ctx, request)
}

func (s ClientServices) listMilestonesForAllOrganizations(ctx context.Context, request ListMilestonesRequest) (MilestoneListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return MilestoneListResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listMilestonesForOrganization(ctx, ListMilestonesRequest{Organization: organization.ID, ProjectID: request.ProjectID, Options: request.Options})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return MilestoneListResult{}, err
	}
	if lastErr != nil {
		return MilestoneListResult{}, lastErr
	}
	return MilestoneListResult{Meta: meta}, nil
}

func (s ClientServices) listMilestonesForOrganization(ctx context.Context, request ListMilestonesRequest) (MilestoneListResult, error) {
	response := milestoneListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: projectPath(request.Organization, request.ProjectID, "milestones"), Query: queryParams(request.Options)}, &response)
	if err != nil {
		return MilestoneListResult{}, err
	}
	return MilestoneListResult{Milestones: firstMilestones(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) GetMilestone(ctx context.Context, request GetMilestoneRequest) (MilestoneResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.getMilestoneForAllOrganizations(ctx, request)
	}
	return s.getMilestoneForOrganization(ctx, request)
}

func (s ClientServices) getMilestoneForAllOrganizations(ctx context.Context, request GetMilestoneRequest) (MilestoneResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return MilestoneResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.getMilestoneForOrganization(ctx, GetMilestoneRequest{Organization: organization.ID, ProjectID: request.ProjectID, MilestoneID: request.MilestoneID})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return MilestoneResult{}, err
	}
	if lastErr != nil {
		return MilestoneResult{}, lastErr
	}
	return MilestoneResult{Meta: meta}, nil
}

func (s ClientServices) getMilestoneForOrganization(ctx context.Context, request GetMilestoneRequest) (MilestoneResult, error) {
	response := milestoneResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: projectPath(request.Organization, request.ProjectID, pathf("milestones/%s", request.MilestoneID))}, &response)
	if err != nil {
		return MilestoneResult{}, err
	}
	return MilestoneResult{Milestone: firstMilestone(response.Data, response.Milestone), Meta: meta}, nil
}

func (s ClientServices) ListLabels(ctx context.Context, request ListLabelsRequest) (LabelListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listLabelsForAllOrganizations(ctx, request)
	}
	return s.listLabelsForOrganization(ctx, request)
}

func (s ClientServices) listLabelsForAllOrganizations(ctx context.Context, request ListLabelsRequest) (LabelListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return LabelListResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listLabelsForOrganization(ctx, ListLabelsRequest{Organization: organization.ID, ProjectID: request.ProjectID, Options: request.Options})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return LabelListResult{}, err
	}
	if lastErr != nil {
		return LabelListResult{}, lastErr
	}
	return LabelListResult{Meta: meta}, nil
}

func (s ClientServices) listLabelsForOrganization(ctx context.Context, request ListLabelsRequest) (LabelListResult, error) {
	response := labelListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: projectPath(request.Organization, request.ProjectID, "labels"), Query: queryParams(request.Options)}, &response)
	if err != nil {
		return LabelListResult{}, err
	}
	return LabelListResult{Labels: firstLabels(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func (s ClientServices) ListWorkItems(ctx context.Context, request ListWorkItemsRequest) (WorkItemListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listWorkItemsForAllOrganizations(ctx, request)
	}
	return s.listWorkItemsForOrganization(ctx, request)
}

func (s ClientServices) listWorkItemsForAllOrganizations(ctx context.Context, request ListWorkItemsRequest) (WorkItemListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return WorkItemListResult{}, err
	}
	workItems := []WorkItem{}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listWorkItemsForOrganization(ctx, ListWorkItemsRequest{
			Organization: organization.ID,
			ProjectID:    request.ProjectID,
			State:        request.State,
			Assignee:     request.Assignee,
			Options:      request.Options,
		})
		if err == nil {
			workItems = append(workItems, result.WorkItems...)
			meta = result.Meta
			continue
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return WorkItemListResult{}, err
	}
	if len(workItems) > 0 || lastErr == nil {
		return WorkItemListResult{WorkItems: workItems, Meta: meta}, nil
	}
	return WorkItemListResult{}, lastErr
}

func (s ClientServices) listWorkItemsForOrganization(ctx context.Context, request ListWorkItemsRequest) (WorkItemListResult, error) {
	workItems := []WorkItem{}
	var meta api.ResponseMeta
	for _, category := range workItemCategories(request.Category) {
		next := request
		next.Category = category
		result, err := s.searchWorkItemsForOrganization(ctx, next)
		if err != nil {
			return WorkItemListResult{}, err
		}
		workItems = append(workItems, result.WorkItems...)
		meta = result.Meta
	}
	return WorkItemListResult{WorkItems: workItems, Meta: meta}, nil
}

func (s ClientServices) searchWorkItemsForOrganization(ctx context.Context, request ListWorkItemsRequest) (WorkItemListResult, error) {
	body, err := api.EncodeJSONBody(workItemSearchPayload(request))
	if err != nil {
		return WorkItemListResult{}, err
	}
	response := workItemListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPost, Path: pathf("/oapi/v1/projex/organizations/%s/workitems:search", request.Organization), Body: body}, &response)
	if err != nil {
		return WorkItemListResult{}, err
	}
	return WorkItemListResult{WorkItems: firstWorkItems(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func workItemCategories(category string) []string {
	if strings.TrimSpace(category) != "" {
		return []string{category}
	}
	return []string{"Req", "Task", "Bug"}
}

func projectPath(organization string, projectID string, resource string) string {
	base := pathf("/oapi/v1/projex/organizations/%s/projects/%s", organization, projectID)
	if strings.TrimSpace(resource) == "" {
		return base
	}
	return base + "/" + strings.Trim(resource, "/")
}

func workItemPath(organization string, resource string) string {
	base := pathf("/oapi/v1/projex/organizations/%s/workitems", organization)
	if strings.TrimSpace(resource) == "" {
		return base
	}
	return base + "/" + strings.Trim(resource, "/")
}

func workItemSearchPayload(request ListWorkItemsRequest) map[string]any {
	payload := map[string]any{}
	if strings.TrimSpace(request.ProjectID) != "" {
		payload["spaceId"] = request.ProjectID
	}
	if strings.TrimSpace(request.Category) != "" {
		payload["category"] = request.Category
	}
	if strings.TrimSpace(request.State) != "" {
		payload["status"] = request.State
	}
	if strings.TrimSpace(request.Assignee) != "" {
		payload["assignedTo"] = request.Assignee
	}
	if strings.TrimSpace(request.Options.Query) != "" {
		payload["keyword"] = request.Options.Query
		payload["search"] = request.Options.Query
	}
	if request.Options.Page > 0 {
		payload["page"] = request.Options.Page
	}
	if request.Options.PerPage > 0 {
		payload["perPage"] = request.Options.PerPage
	}
	return payload
}

func (s ClientServices) GetWorkItem(ctx context.Context, request GetWorkItemRequest) (WorkItemResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.getWorkItemForAllOrganizations(ctx, request)
	}
	return s.getWorkItemForOrganization(ctx, request)
}

func (s ClientServices) getWorkItemForAllOrganizations(ctx context.Context, request GetWorkItemRequest) (WorkItemResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return WorkItemResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.getWorkItemForOrganization(ctx, GetWorkItemRequest{Organization: organization.ID, ID: request.ID})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return WorkItemResult{}, err
	}
	if lastErr != nil {
		return WorkItemResult{}, lastErr
	}
	return WorkItemResult{Meta: meta}, nil
}

func (s ClientServices) getWorkItemForOrganization(ctx context.Context, request GetWorkItemRequest) (WorkItemResult, error) {
	response := workItemResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: workItemPath(request.Organization, request.ID)}, &response)
	if err != nil {
		return WorkItemResult{}, err
	}
	return WorkItemResult{WorkItem: firstWorkItem(response.Data, response.WorkItem), Meta: meta}, nil
}

func (s ClientServices) CreateWorkItem(ctx context.Context, request CreateWorkItemRequest) (WorkItemResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.createWorkItemForAllOrganizations(ctx, request)
	}
	return s.createWorkItemForOrganization(ctx, request)
}

func (s ClientServices) createWorkItemForAllOrganizations(ctx context.Context, request CreateWorkItemRequest) (WorkItemResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return WorkItemResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		next := request
		next.Organization = organization.ID
		result, err := s.createWorkItemForOrganization(ctx, next)
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return WorkItemResult{}, err
	}
	if lastErr != nil {
		return WorkItemResult{}, lastErr
	}
	return WorkItemResult{Meta: meta}, nil
}

func (s ClientServices) createWorkItemForOrganization(ctx context.Context, request CreateWorkItemRequest) (WorkItemResult, error) {
	typeID, err := s.resolveWorkItemTypeID(ctx, request)
	if err != nil {
		return WorkItemResult{}, err
	}
	body, err := api.EncodeJSONBody(createWorkItemPayload(request, typeID))
	if err != nil {
		return WorkItemResult{}, err
	}
	response := workItemResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPost, Path: workItemPath(request.Organization, ""), Body: body}, &response)
	if err != nil {
		return WorkItemResult{}, err
	}
	return WorkItemResult{WorkItem: firstWorkItem(response.Data, response.WorkItem), Meta: meta}, nil
}

func (s ClientServices) resolveWorkItemTypeID(ctx context.Context, request CreateWorkItemRequest) (string, error) {
	if !isWorkItemCategory(request.Type) {
		return request.Type, nil
	}
	result, err := s.searchWorkItemsForOrganization(ctx, ListWorkItemsRequest{
		Organization: request.Organization,
		ProjectID:    request.ProjectID,
		Category:     request.Type,
		Options:      ListOptions{Page: 1, PerPage: 1},
	})
	if err != nil {
		return "", err
	}
	for _, item := range result.WorkItems {
		if item.TypeID != "" {
			return item.TypeID, nil
		}
	}
	return request.Type, nil
}

func isWorkItemCategory(value string) bool {
	switch value {
	case "Req", "Task", "Bug":
		return true
	default:
		return false
	}
}

func createWorkItemPayload(request CreateWorkItemRequest, typeID string) map[string]any {
	payload := map[string]any{
		"spaceId":        request.ProjectID,
		"subject":        request.Title,
		"workitemTypeId": typeID,
	}
	if strings.TrimSpace(request.Body) != "" {
		payload["description"] = request.Body
		payload["formatType"] = "MARKDOWN"
	}
	if strings.TrimSpace(request.Assignee) != "" {
		payload["assignedTo"] = request.Assignee
	}
	return payload
}

func updateWorkItemPayload(request UpdateWorkItemRequest) map[string]any {
	payload := map[string]any{}
	if strings.TrimSpace(request.Title) != "" {
		payload["subject"] = request.Title
	}
	if strings.TrimSpace(request.Body) != "" {
		payload["description"] = request.Body
		payload["formatType"] = "MARKDOWN"
	}
	if strings.TrimSpace(request.State) != "" {
		payload["status"] = request.State
	}
	if strings.TrimSpace(request.Assignee) != "" {
		payload["assignedTo"] = request.Assignee
	}
	return payload
}

func (s ClientServices) UpdateWorkItem(ctx context.Context, request UpdateWorkItemRequest) (WorkItemResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.updateWorkItemForAllOrganizations(ctx, request)
	}
	return s.updateWorkItemForOrganization(ctx, request)
}

func (s ClientServices) updateWorkItemForAllOrganizations(ctx context.Context, request UpdateWorkItemRequest) (WorkItemResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return WorkItemResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		next := request
		next.Organization = organization.ID
		result, err := s.updateWorkItemForOrganization(ctx, next)
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return WorkItemResult{}, err
	}
	if lastErr != nil {
		return WorkItemResult{}, lastErr
	}
	return WorkItemResult{Meta: meta}, nil
}

func (s ClientServices) updateWorkItemForOrganization(ctx context.Context, request UpdateWorkItemRequest) (WorkItemResult, error) {
	body, err := api.EncodeJSONBody(updateWorkItemPayload(request))
	if err != nil {
		return WorkItemResult{}, err
	}
	response := workItemResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodPut, Path: workItemPath(request.Organization, request.ID), Body: body}, &response)
	if err != nil {
		return WorkItemResult{}, err
	}
	return WorkItemResult{WorkItem: firstWorkItem(response.Data, response.WorkItem), Meta: meta}, nil
}

func (s ClientServices) DeleteWorkItem(ctx context.Context, request DeleteWorkItemRequest) (WorkItemActionResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.deleteWorkItemForAllOrganizations(ctx, request)
	}
	return s.deleteWorkItemForOrganization(ctx, request)
}

func (s ClientServices) deleteWorkItemForAllOrganizations(ctx context.Context, request DeleteWorkItemRequest) (WorkItemActionResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return WorkItemActionResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.deleteWorkItemForOrganization(ctx, DeleteWorkItemRequest{Organization: organization.ID, ID: request.ID, Yes: request.Yes})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return WorkItemActionResult{}, err
	}
	if lastErr != nil {
		return WorkItemActionResult{}, lastErr
	}
	return WorkItemActionResult{ID: request.ID, Meta: meta}, nil
}

func (s ClientServices) deleteWorkItemForOrganization(ctx context.Context, request DeleteWorkItemRequest) (WorkItemActionResult, error) {
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodDelete, Path: workItemPath(request.Organization, request.ID)}, nil)
	if err != nil {
		return WorkItemActionResult{}, err
	}
	return WorkItemActionResult{ID: request.ID, Meta: meta}, nil
}

func (s ClientServices) ListWorkItemActivities(ctx context.Context, request ListWorkItemActivitiesRequest) (WorkItemActivityListResult, error) {
	if strings.TrimSpace(request.Organization) == "" {
		return s.listWorkItemActivitiesForAllOrganizations(ctx, request)
	}
	return s.listWorkItemActivitiesForOrganization(ctx, request)
}

func (s ClientServices) listWorkItemActivitiesForAllOrganizations(ctx context.Context, request ListWorkItemActivitiesRequest) (WorkItemActivityListResult, error) {
	organizations, meta, err := s.listOrganizations(ctx)
	if err != nil {
		return WorkItemActivityListResult{}, err
	}
	var lastErr error
	for _, organization := range organizations {
		if organization.ID == "" {
			continue
		}
		result, err := s.listWorkItemActivitiesForOrganization(ctx, ListWorkItemActivitiesRequest{Organization: organization.ID, ID: request.ID, Options: request.Options})
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			lastErr = err
			continue
		}
		return WorkItemActivityListResult{}, err
	}
	if lastErr != nil {
		return WorkItemActivityListResult{}, lastErr
	}
	return WorkItemActivityListResult{Meta: meta}, nil
}

func (s ClientServices) listWorkItemActivitiesForOrganization(ctx context.Context, request ListWorkItemActivitiesRequest) (WorkItemActivityListResult, error) {
	response := workItemActivityListResponse{}
	meta, err := s.client.Do(ctx, api.Request{Method: http.MethodGet, Path: workItemPath(request.Organization, pathf("%s/activities", request.ID)), Query: queryParams(request.Options)}, &response)
	if err != nil {
		return WorkItemActivityListResult{}, err
	}
	return WorkItemActivityListResult{Activities: firstWorkItemActivities(response.Data, response.Result, response.Items, response.Records), Meta: meta}, nil
}

func firstProjects(values ...[]Project) []Project {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstProject(values ...Project) Project {
	for _, value := range values {
		if value.ID != "" || value.Name != "" {
			return value
		}
	}
	return Project{}
}

func firstProjectMembers(values ...[]ProjectMember) []ProjectMember {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstIterations(values ...[]Iteration) []Iteration {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstIteration(values ...Iteration) Iteration {
	for _, value := range values {
		if value.ID != "" || value.Name != "" {
			return value
		}
	}
	return Iteration{}
}

func firstMilestones(values ...[]Milestone) []Milestone {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstMilestone(values ...Milestone) Milestone {
	for _, value := range values {
		if value.ID != "" || value.Name != "" {
			return value
		}
	}
	return Milestone{}
}

func firstLabels(values ...[]Label) []Label {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstWorkItems(values ...[]WorkItem) []WorkItem {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstWorkItem(values ...WorkItem) WorkItem {
	for _, value := range values {
		if value.ID != "" || value.Title != "" {
			return value
		}
	}
	return WorkItem{}
}

func firstWorkItemActivities(values ...[]WorkItemActivity) []WorkItemActivity {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}
