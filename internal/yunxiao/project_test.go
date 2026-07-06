package yunxiao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gouzil/yunxiao-cli/internal/api"
)

func TestListProjectsUsesOrganizationSearchAndDecodesCurrentShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/oapi/v1/projex/organizations/org-1/projects:search" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("perPage") != "30" {
			t.Fatalf("query = %q", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id":          "259be5fb2d81a291e0883d6b50",
			"name":        "敏捷研发示例项目",
			"gmtModified": float64(1783231082000),
			"creator":     map[string]any{"name": "gouziya"},
			"status":      map[string]any{"name": "进行中"},
		}})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListProjects(context.Background(), ListProjectsRequest{
		Organization: "org-1",
		Options:      ListOptions{Page: 1, PerPage: 30},
	})
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}
	if len(result.Projects) != 1 {
		t.Fatalf("projects = %#v", result.Projects)
	}
	project := result.Projects[0]
	if project.ID != "259be5fb2d81a291e0883d6b50" || project.Status != "进行中" || project.Owner != "gouziya" || project.UpdatedAt == "" {
		t.Fatalf("project = %#v", project)
	}
}

func TestListProjectsSearchesAllOrganizationsWhenOrganizationMissing(t *testing.T) {
	paths := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/oapi/v1/platform/organizations":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": "org-1", "name": "Org One"},
				{"id": "org-2", "name": "Org Two"},
			})
		case "/oapi/v1/projex/organizations/org-1/projects:search":
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		case "/oapi/v1/projex/organizations/org-2/projects:search":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "project-1", "name": "Demo", "status": map[string]any{"name": "进行中"}}})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListProjects(context.Background(), ListProjectsRequest{})
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}
	if len(result.Projects) != 1 || result.Projects[0].ID != "project-1" {
		t.Fatalf("projects = %#v", result.Projects)
	}
	if len(paths) != 3 || paths[0] != "/oapi/v1/platform/organizations" {
		t.Fatalf("paths = %#v", paths)
	}
}

func TestGetProjectTriesOrganizationsAndDecodesBareObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/platform/organizations":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": "org-1", "name": "Org One"},
				{"id": "org-2", "name": "Org Two"},
			})
		case "/oapi/v1/projex/organizations/org-1/projects/project-1":
			http.Error(w, `{"errorCode":"NotFound","errorMessage":"Not Found"}`, http.StatusNotFound)
		case "/oapi/v1/projex/organizations/org-2/projects/project-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "project-1",
				"name":        "Demo",
				"status":      map[string]any{"name": "进行中"},
				"creator":     map[string]any{"name": "gouzi"},
				"gmtModified": float64(1783231082000),
			})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.GetProject(context.Background(), GetProjectRequest{ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("GetProject returned error: %v", err)
	}
	if result.Project.ID != "project-1" || result.Project.Owner != "gouzi" || result.Project.UpdatedAt == "" {
		t.Fatalf("project = %#v", result.Project)
	}
}

func TestProjectSubresourcesUseOrganizationPathAndDecodeBareArrays(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/projex/organizations/org-1/projects/project-1/members":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"userId": "u1", "userName": "gouzi", "roleName": "admin"}})
		case "/oapi/v1/projex/organizations/org-1/projects/project-1/labels":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "label-1", "name": "backend", "color": "#fff"}})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	members, err := service.ListProjectMembers(context.Background(), ListProjectMembersRequest{Organization: "org-1", ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("ListProjectMembers returned error: %v", err)
	}
	if len(members.Members) != 1 || members.Members[0].Name != "gouzi" || members.Members[0].Role != "admin" {
		t.Fatalf("members = %#v", members.Members)
	}

	labels, err := service.ListLabels(context.Background(), ListLabelsRequest{Organization: "org-1", ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("ListLabels returned error: %v", err)
	}
	if len(labels.Labels) != 1 || labels.Labels[0].ID != "label-1" {
		t.Fatalf("labels = %#v", labels.Labels)
	}
}

func TestListWorkItemsUsesOrganizationSearchAndDecodesCurrentShape(t *testing.T) {
	categories := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/oapi/v1/projex/organizations/org-1/workitems:search" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["spaceId"] != "project-1" {
			t.Fatalf("body = %#v", body)
		}
		category, _ := body["category"].(string)
		categories = append(categories, category)
		if category != "Req" {
			_ = json.NewEncoder(w).Encode([]map[string]any{})
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id":           "9e88dfd0bf395ae1148374c5d9",
			"serialNumber": "DEMO-26",
			"subject":      "Fix login",
			"workitemType": map[string]any{"name": "需求"},
			"status":       map[string]any{"name": "进行中"},
			"assignedTo":   map[string]any{"name": "alice"},
			"creator":      map[string]any{"name": "bob"},
			"space":        map[string]any{"id": "project-1", "name": "Demo"},
			"gmtModified":  float64(1783231082000),
		}})
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.ListWorkItems(context.Background(), ListWorkItemsRequest{Organization: "org-1", ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("ListWorkItems returned error: %v", err)
	}
	if len(result.WorkItems) != 1 {
		t.Fatalf("workItems = %#v", result.WorkItems)
	}
	item := result.WorkItems[0]
	if item.ID != "9e88dfd0bf395ae1148374c5d9" || item.Title != "Fix login" || item.State != "进行中" || item.ProjectName != "Demo" || item.UpdatedAt == "" {
		t.Fatalf("workItem = %#v", item)
	}
	if len(categories) != 3 || categories[0] != "Req" || categories[1] != "Task" || categories[2] != "Bug" {
		t.Fatalf("categories = %#v", categories)
	}
}

func TestCreateWorkItemUsesOrganizationPathAndResolvesCategoryType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/projex/organizations/org-1/workitems:search":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode search body: %v", err)
			}
			if body["category"] != "Task" || body["spaceId"] != "project-1" {
				t.Fatalf("search body = %#v", body)
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "existing", "subject": "Existing", "workitemType": map[string]any{"id": "type-task", "name": "任务"}}})
		case "/oapi/v1/projex/organizations/org-1/workitems":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			if body["spaceId"] != "project-1" || body["subject"] != "New task" || body["workitemTypeId"] != "type-task" || body["description"] != "body" {
				t.Fatalf("create body = %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "new-1", "subject": "New task", "workitemType": map[string]any{"id": "type-task", "name": "任务"}})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	result, err := service.CreateWorkItem(context.Background(), CreateWorkItemRequest{Organization: "org-1", ProjectID: "project-1", Type: "Task", Title: "New task", Body: "body"})
	if err != nil {
		t.Fatalf("CreateWorkItem returned error: %v", err)
	}
	if result.WorkItem.ID != "new-1" || result.WorkItem.Type != "任务" {
		t.Fatalf("workItem = %#v", result.WorkItem)
	}
}

func TestWorkItemDetailActionsUseOrganizationPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/projex/organizations/org-1/workitems/item-1":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "item-1", "subject": "New task", "status": map[string]any{"name": "待处理"}})
		case r.Method == http.MethodPut && r.URL.Path == "/oapi/v1/projex/organizations/org-1/workitems/item-1":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			if body["subject"] != "Updated task" || body["description"] != "updated body" || body["status"] != "100005" || body["assignedTo"] != "user-1" {
				t.Fatalf("update body = %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "item-1", "subject": "Updated task"})
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/projex/organizations/org-1/workitems/item-1/activities":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"eventTime": "2026-07-05T10:00:00Z", "actionType": "updated", "operator": map[string]any{"name": "gouzi"}}})
		case r.Method == http.MethodDelete && r.URL.Path == "/oapi/v1/projex/organizations/org-1/workitems/item-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewClientServices(api.NewClient(api.ClientOptions{Endpoint: server.URL}))
	item, err := service.GetWorkItem(context.Background(), GetWorkItemRequest{Organization: "org-1", ID: "item-1"})
	if err != nil {
		t.Fatalf("GetWorkItem returned error: %v", err)
	}
	if item.WorkItem.ID != "item-1" || item.WorkItem.Title != "New task" {
		t.Fatalf("item = %#v", item.WorkItem)
	}

	updated, err := service.UpdateWorkItem(context.Background(), UpdateWorkItemRequest{Organization: "org-1", ID: "item-1", Title: "Updated task", Body: "updated body", State: "100005", Assignee: "user-1"})
	if err != nil {
		t.Fatalf("UpdateWorkItem returned error: %v", err)
	}
	if updated.WorkItem.Title != "Updated task" {
		t.Fatalf("updated = %#v", updated.WorkItem)
	}

	activities, err := service.ListWorkItemActivities(context.Background(), ListWorkItemActivitiesRequest{Organization: "org-1", ID: "item-1"})
	if err != nil {
		t.Fatalf("ListWorkItemActivities returned error: %v", err)
	}
	if len(activities.Activities) != 1 || activities.Activities[0].Actor != "gouzi" || activities.Activities[0].Action != "updated" {
		t.Fatalf("activities = %#v", activities.Activities)
	}

	deleted, err := service.DeleteWorkItem(context.Background(), DeleteWorkItemRequest{Organization: "org-1", ID: "item-1"})
	if err != nil {
		t.Fatalf("DeleteWorkItem returned error: %v", err)
	}
	if deleted.ID != "item-1" {
		t.Fatalf("deleted = %#v", deleted)
	}
}
