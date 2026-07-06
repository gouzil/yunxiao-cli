package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gouzi/yunxiao-cli/internal/auth"
	"github.com/gouzi/yunxiao-cli/internal/config"
	"github.com/gouzi/yunxiao-cli/internal/extension"
	"github.com/gouzi/yunxiao-cli/internal/output"
	"github.com/gouzi/yunxiao-cli/internal/terminal"
	"github.com/gouzi/yunxiao-cli/internal/yunxiao"
)

func TestRepoListCommandUsesTypedService(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service})

	if err := root.Execute(context.Background(), []string{"repo", "list"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.listRepositoriesRequest.Organization != "org-1" {
		t.Fatalf("organization = %q", service.listRepositoriesRequest.Organization)
	}
	if !strings.Contains(out.String(), "repo-1") || !strings.Contains(out.String(), "api") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRepoListAllowsMissingOrganization(t *testing.T) {
	service := &fakeCommandService{}
	root, _, errOut := newCommandTestRootWithValues(t, yunxiao.ServiceSet{Repo: service}, nil, config.Values{}, config.Values{})

	if err := root.Execute(context.Background(), []string{"repo", "list"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if errOut.String() != "" {
		t.Fatalf("stderr = %q", errOut.String())
	}
	if !service.listRepositoriesCalled {
		t.Fatal("repository service was not called")
	}
}

func TestRepoViewResolvesRepositoryPathFromList(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service})

	if err := root.Execute(context.Background(), []string{"repo", "view", "6a49f26608f52788b13355d2/test-repo"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(out.String(), "test-repo") || !strings.Contains(out.String(), "master") || !strings.Contains(out.String(), "git@codeup.aliyun.com") {
		t.Fatalf("output = %q", out.String())
	}
	if !service.getRepositoryCalled {
		t.Fatal("repository detail service was not called")
	}
	if service.getRepositoryRequest.Organization != "6a49f26608f52788b13355d2" || service.getRepositoryRequest.RepositoryID != "repo-2" {
		t.Fatalf("get request = %#v", service.getRepositoryRequest)
	}
}

func TestRepoWriteCommandsResolveRepositoryOrganization(t *testing.T) {
	t.Run("update", func(t *testing.T) {
		service := &fakeCommandService{}
		root, _ := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service})

		if err := root.Execute(context.Background(), []string{"repo", "update", "6a49f26608f52788b13355d2/test-repo", "--name", "renamed"}); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
		if service.updateRepositoryRequest.Organization != "6a49f26608f52788b13355d2" || service.updateRepositoryRequest.RepositoryID != "repo-2" {
			t.Fatalf("update request = %#v", service.updateRepositoryRequest)
		}
	})

	t.Run("archive", func(t *testing.T) {
		service := &fakeCommandService{}
		root, _ := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service})

		if err := root.Execute(context.Background(), []string{"repo", "archive", "6a49f26608f52788b13355d2/test-repo"}); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
		if service.archiveRepositoryRequest.Organization != "6a49f26608f52788b13355d2" || service.archiveRepositoryRequest.RepositoryID != "repo-2" {
			t.Fatalf("archive request = %#v", service.archiveRepositoryRequest)
		}
	})

	t.Run("delete", func(t *testing.T) {
		service := &fakeCommandService{}
		root, _ := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service})

		if err := root.Execute(context.Background(), []string{"repo", "delete", "6a49f26608f52788b13355d2/test-repo", "--yes"}); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
		if service.deleteRepositoryRequest.Organization != "6a49f26608f52788b13355d2" || service.deleteRepositoryRequest.RepositoryID != "repo-2" {
			t.Fatalf("delete request = %#v", service.deleteRepositoryRequest)
		}
	})
}

func TestBranchListResolvesRepositoryOrganization(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service, Branch: service})

	if err := root.Execute(context.Background(), []string{"--repo", "repo-2", "branch", "list"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.listBranchesRequest.Organization != "6a49f26608f52788b13355d2" || service.listBranchesRequest.RepositoryID != "repo-2" {
		t.Fatalf("request = %#v", service.listBranchesRequest)
	}
	if !strings.Contains(out.String(), "master") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestCommitListUsesDefaultBranchFromResolvedRepository(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service, Commit: service})

	if err := root.Execute(context.Background(), []string{"--repo", "repo-2", "commit", "list"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.listCommitsRequest.Organization != "6a49f26608f52788b13355d2" || service.listCommitsRequest.RepositoryID != "repo-2" || service.listCommitsRequest.Branch != "master" {
		t.Fatalf("request = %#v", service.listCommitsRequest)
	}
	if !strings.Contains(out.String(), "Initial commit") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestFileViewDecodesBase64Content(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service, File: service})

	if err := root.Execute(context.Background(), []string{"--repo", "repo-2", "file", "view", "README.md", "--ref", "master"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.getFileRequest.Organization != "6a49f26608f52788b13355d2" || service.getFileRequest.RepositoryID != "repo-2" {
		t.Fatalf("request = %#v", service.getFileRequest)
	}
	if got := out.String(); got != "## 测试仓库" {
		t.Fatalf("output = %q", got)
	}
}

func TestSearchCodeResolvesRepositoryOrganizationAndDefaultRef(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service, Search: service})

	if err := root.Execute(context.Background(), []string{"search", "code", "README", "--repo", "repo-2"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.searchCodeRequest.Organization != "6a49f26608f52788b13355d2" || service.searchCodeRequest.RepositoryID != "repo-2" || service.searchCodeRequest.Ref != "master" {
		t.Fatalf("request = %#v", service.searchCodeRequest)
	}
	if !strings.Contains(out.String(), "README") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestSearchCommitResolvesRepositoryOrganizationAndDefaultRef(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service, Search: service})

	if err := root.Execute(context.Background(), []string{"search", "commit", "README", "--repo", "repo-2"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.searchCommitsRequest.Organization != "6a49f26608f52788b13355d2" || service.searchCommitsRequest.RepositoryID != "repo-2" || service.searchCommitsRequest.Ref != "master" {
		t.Fatalf("request = %#v", service.searchCommitsRequest)
	}
	if !strings.Contains(out.String(), "README") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestSearchMRResolvesRepositoryOrganization(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service, Search: service})

	if err := root.Execute(context.Background(), []string{"search", "mr", "README", "--repo", "repo-2"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.searchMergeRequestsRequest.Organization != "6a49f26608f52788b13355d2" || service.searchMergeRequestsRequest.RepositoryID != "repo-2" {
		t.Fatalf("request = %#v", service.searchMergeRequestsRequest)
	}
	if !strings.Contains(out.String(), "README") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestMRListCommandUsesDefaultRepo(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service, MR: service})

	if err := root.Execute(context.Background(), []string{"mr", "list"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.listMergeRequestsRequest.RepositoryID != "repo-1" {
		t.Fatalf("repository = %q", service.listMergeRequestsRequest.RepositoryID)
	}
	if !strings.Contains(out.String(), "Add CLI") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestMRCommentCommandParsesInlineFlags(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service, MR: service})

	if err := root.Execute(context.Background(), []string{"mr", "comment", "3", "--body", "inline", "--file", "main.go", "--line", "3"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.commentMergeRequestRequest.MergeRequestID != "3" || service.commentMergeRequestRequest.Body != "inline" || service.commentMergeRequestRequest.FilePath != "main.go" || service.commentMergeRequestRequest.LineNumber != 3 {
		t.Fatalf("request = %#v", service.commentMergeRequestRequest)
	}
	if !strings.Contains(out.String(), "Comment ID: comment-1") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestMRResolveCommentCommandParsesIDs(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service, MR: service})

	if err := root.Execute(context.Background(), []string{"mr", "resolve-comment", "3", "comment-inline"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.resolveMergeRequestCommentRequest.MergeRequestID != "3" || service.resolveMergeRequestCommentRequest.CommentID != "comment-inline" || !service.resolveMergeRequestCommentRequest.Resolved {
		t.Fatalf("request = %#v", service.resolveMergeRequestCommentRequest)
	}
	if !strings.Contains(out.String(), "Resolved comment comment-inline") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestMRUnresolveCommentCommandParsesIDs(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Repo: service, MR: service})

	if err := root.Execute(context.Background(), []string{"mr", "unresolve-comment", "3", "comment-inline"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.resolveMergeRequestCommentRequest.MergeRequestID != "3" || service.resolveMergeRequestCommentRequest.CommentID != "comment-inline" || service.resolveMergeRequestCommentRequest.Resolved {
		t.Fatalf("request = %#v", service.resolveMergeRequestCommentRequest)
	}
	if !strings.Contains(out.String(), "Marked comment comment-inline as unresolved") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestPipelineRunCommandParsesVariables(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Pipeline: service})

	err := root.Execute(context.Background(), []string{"pipeline", "run", "pipe-1", "--branch", "master", "--var", "env=prod"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.runPipelineRequest.PipelineID != "pipe-1" || service.runPipelineRequest.Branch != "master" {
		t.Fatalf("request = %#v", service.runPipelineRequest)
	}
	if len(service.runPipelineRequest.Variables) != 1 || service.runPipelineRequest.Variables[0].Name != "env" || service.runPipelineRequest.Variables[0].Value != "prod" {
		t.Fatalf("variables = %#v", service.runPipelineRequest.Variables)
	}
	if !strings.Contains(out.String(), "Run ID: run-1") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestPipelineListCommandUsesDefaultOrganizationAndProject(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Pipeline: service})

	if err := root.Execute(context.Background(), []string{"pipeline", "list"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.listPipelinesRequest.Organization != "org-1" || service.listPipelinesRequest.ProjectID != "project-1" {
		t.Fatalf("request = %#v", service.listPipelinesRequest)
	}
	if !strings.Contains(out.String(), "pipe-1") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunListCommandUsesDefaultOrganizationAndRequiresPipeline(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Run: service})

	if err := root.Execute(context.Background(), []string{"run", "list"}); err == nil || !strings.Contains(err.Error(), "--pipeline is required") {
		t.Fatalf("error = %v", err)
	}
	if err := root.Execute(context.Background(), []string{"run", "list", "--pipeline", "pipe-1"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.listRunsRequest.Organization != "org-1" || service.listRunsRequest.PipelineID != "pipe-1" {
		t.Fatalf("request = %#v", service.listRunsRequest)
	}
	if !strings.Contains(out.String(), "run-1") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunDetailCommandsUseDefaultOrganizationAndRequirePipeline(t *testing.T) {
	service := &fakeCommandService{}
	root, _ := newCommandTestRoot(t, yunxiao.ServiceSet{Run: service})

	for _, args := range [][]string{
		{"run", "view", "13"},
		{"run", "log", "13"},
		{"run", "watch", "13", "--timeout", "1ms"},
	} {
		if err := root.Execute(context.Background(), args); err == nil || !strings.Contains(err.Error(), "--pipeline is required") {
			t.Fatalf("%v error = %v", args, err)
		}
	}

	if err := root.Execute(context.Background(), []string{"run", "view", "13", "--pipeline", "pipe-1"}); err != nil {
		t.Fatalf("run view returned error: %v", err)
	}
	if service.getRunRequest.Organization != "org-1" || service.getRunRequest.PipelineID != "pipe-1" || service.getRunRequest.RunID != "13" {
		t.Fatalf("getRunRequest = %#v", service.getRunRequest)
	}

	if err := root.Execute(context.Background(), []string{"run", "log", "13", "--pipeline", "pipe-1"}); err != nil {
		t.Fatalf("run log returned error: %v", err)
	}
	if service.getRunLogRequest.Organization != "org-1" || service.getRunLogRequest.PipelineID != "pipe-1" || service.getRunLogRequest.RunID != "13" {
		t.Fatalf("getRunLogRequest = %#v", service.getRunLogRequest)
	}

	if err := root.Execute(context.Background(), []string{"run", "watch", "13", "--pipeline", "pipe-1", "--timeout", "1ms"}); err != nil {
		t.Fatalf("run watch returned error: %v", err)
	}
	if service.watchRunRequest.Organization != "org-1" || service.watchRunRequest.PipelineID != "pipe-1" || service.watchRunRequest.RunID != "13" {
		t.Fatalf("watchRunRequest = %#v", service.watchRunRequest)
	}
}

func TestSSHKeyListCommandUsesDefaultOrganization(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{SSHKey: service})

	if err := root.Execute(context.Background(), []string{"ssh-key", "list"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.listSSHKeysRequest.Organization != "org-1" {
		t.Fatalf("request = %#v", service.listSSHKeysRequest)
	}
	if !strings.Contains(out.String(), "a100 vm") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestWorkItemListCommandUsesDefaultProject(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{WorkItem: service})

	if err := root.Execute(context.Background(), []string{"workitem", "list"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.listWorkItemsRequest.ProjectID != "project-1" {
		t.Fatalf("project = %q", service.listWorkItemsRequest.ProjectID)
	}
	if !strings.Contains(out.String(), "WI-1") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestWorkItemDeleteAcceptsYesFlag(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{WorkItem: service})

	if err := root.Execute(context.Background(), []string{"workitem", "delete", "WI-1", "--yes"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.deleteWorkItemRequest.Organization != "org-1" || service.deleteWorkItemRequest.ID != "WI-1" || !service.deleteWorkItemRequest.Yes {
		t.Fatalf("deleteWorkItemRequest = %#v", service.deleteWorkItemRequest)
	}
	if !strings.Contains(out.String(), "Deleted work item WI-1") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestSearchRepoCommandUsesTypedSearchService(t *testing.T) {
	service := &fakeCommandService{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{Search: service})

	if err := root.Execute(context.Background(), []string{"search", "repo", "api"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if service.searchRepositoriesRequest.Query != "api" || service.searchRepositoriesRequest.Organization != "org-1" {
		t.Fatalf("request = %#v", service.searchRepositoriesRequest)
	}
	if !strings.Contains(out.String(), "repo-1") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestAPICommandOutputsRawJSONForExtensions(t *testing.T) {
	service := &fakeCommandService{}
	root, out, errOut := newCommandTestRootWithBuffers(t, yunxiao.ServiceSet{RawAPI: service}, nil)

	err := root.Execute(context.Background(), []string{"--verbose", "--jq", ".data.id", "api", "GET", "/resource"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got := out.String(); got != "\"42\"\n" {
		t.Fatalf("stdout = %q", got)
	}
	if !strings.Contains(errOut.String(), "> GET /resource") || !strings.Contains(errOut.String(), "< HTTP 200") {
		t.Fatalf("stderr = %q", errOut.String())
	}
	if service.rawAPIRequest.Path != "/resource" {
		t.Fatalf("request = %#v", service.rawAPIRequest)
	}
}

func TestAuthLoginPromptShowsTokenCreationLink(t *testing.T) {
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{})
	root.io.In = strings.NewReader("\nsecret-token\n")
	root.authManager = auth.NewManager(&fakeCredentialStore{}, fakeTokenVerifier{})

	if err := root.Execute(context.Background(), []string{"auth", "login"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(out.String(), "https://account-devops.aliyun.com/settings/personalAccessToken") {
		t.Fatalf("output = %q", out.String())
	}
	if !strings.Contains(out.String(), "组织管理 / 用户 / 只读") {
		t.Fatalf("output = %q", out.String())
	}
	if !strings.Contains(out.String(), "代码管理 / 用户资源、代码仓库、分支、提交、文件 / 只读") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestAuthLoginWithTokenAndLogout(t *testing.T) {
	store := &storingCredentialStore{}
	verifier := &recordingTokenVerifier{}
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{})
	root.io.In = strings.NewReader("secret-token\n")
	root.authManager = auth.NewManager(store, verifier)

	if err := root.Execute(context.Background(), []string{"--endpoint", "http://127.0.0.1:9999", "auth", "login", "--with-token"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}
	if verifier.endpoint != "http://127.0.0.1:9999" || verifier.token != "secret-token" {
		t.Fatalf("verifier endpoint/token = %q/%q", verifier.endpoint, verifier.token)
	}
	if store.saved.Token != "secret-token" {
		t.Fatalf("saved token = %q", store.saved.Token)
	}

	out.Reset()
	if err := root.Execute(context.Background(), []string{"--endpoint", "http://127.0.0.1:9999", "auth", "logout"}); err != nil {
		t.Fatalf("logout returned error: %v", err)
	}
	if store.saved.Endpoint != "" {
		t.Fatalf("credential was not deleted: %#v", store.saved)
	}
	if !strings.Contains(out.String(), "Logged out") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestConfigGetReturnsEffectiveDefaultValue(t *testing.T) {
	root, out := newCommandTestRoot(t, yunxiao.ServiceSet{})

	if err := root.Execute(context.Background(), []string{"config", "get", "endpoint"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got := out.String(); got != "openapi-rdc.aliyuncs.com\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestExtensionDirectDispatchPreservesArgs(t *testing.T) {
	manager := &fakeExtensionManager{extensions: []extension.Extension{{Name: "hello", FullName: "yunxiao-hello", Kind: extension.KindLocal}}}
	root, out, _ := newCommandTestRootWithBuffers(t, yunxiao.ServiceSet{}, manager)

	if err := root.Execute(context.Background(), []string{"extension", "list"}); err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	if strings.Contains(out.String(), "core command") {
		t.Fatalf("non-conflicting extension marked as conflict: %q", out.String())
	}
	err := root.Execute(context.Background(), []string{"hello", "--name", "world"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if manager.dispatchedName != "hello" {
		t.Fatalf("dispatched name = %q", manager.dispatchedName)
	}
	if strings.Join(manager.dispatchedArgs, " ") != "--name world" {
		t.Fatalf("args = %#v", manager.dispatchedArgs)
	}
	if manager.dispatchedContext.Repo != "repo-1" || manager.dispatchedContext.Project != "project-1" {
		t.Fatalf("context = %#v", manager.dispatchedContext)
	}
}

func TestExtensionConflictUsesExplicitExec(t *testing.T) {
	manager := &fakeExtensionManager{extensions: []extension.Extension{{Name: "repo", FullName: "yunxiao-repo", Kind: extension.KindLocal}}}
	root, out, _ := newCommandTestRootWithBuffers(t, yunxiao.ServiceSet{}, manager)

	err := root.Execute(context.Background(), []string{"extension", "list"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(out.String(), "core command") {
		t.Fatalf("list output = %q", out.String())
	}
	err = root.Execute(context.Background(), []string{"extension", "exec", "repo", "--dry-run"})
	if err != nil {
		t.Fatalf("exec returned error: %v", err)
	}
	if manager.dispatchedName != "repo" || strings.Join(manager.dispatchedArgs, " ") != "--dry-run" {
		t.Fatalf("dispatch = %q %#v", manager.dispatchedName, manager.dispatchedArgs)
	}
}

func newCommandTestRoot(t *testing.T, services yunxiao.ServiceSet) (*Root, *bytes.Buffer) {
	root, out, _ := newCommandTestRootWithBuffers(t, services, nil)
	return root, out
}

func newCommandTestRootWithBuffers(t *testing.T, services yunxiao.ServiceSet, manager extension.Manager) (*Root, *bytes.Buffer, *bytes.Buffer) {
	return newCommandTestRootWithValues(t, services, manager, config.Values{Organization: "org-1"}, config.Values{Repo: "repo-1", Project: "project-1"})
}

func newCommandTestRootWithValues(t *testing.T, services yunxiao.ServiceSet, manager extension.Manager, globalValues config.Values, repoValues config.Values) (*Root, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	dir := t.TempDir()
	store := config.NewStore(config.Paths{
		Global: filepath.Join(dir, "global.json"),
		Repo:   filepath.Join(dir, "repo.json"),
	}, nil)
	if err := store.Save(config.ScopeGlobal, config.File{Values: globalValues}); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(config.ScopeRepo, config.File{Values: repoValues}); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	root := NewRoot(RootOptions{
		Build:       BuildInfo{Version: "test"},
		IO:          terminal.IOStreams{In: strings.NewReader(""), Out: out, ErrOut: errOut},
		ConfigStore: store,
		Extensions:  manager,
		Services:    services,
		Renderer:    output.NewRenderer(out),
	})
	return root, out, errOut
}

type fakeCommandService struct {
	listRepositoriesRequest           yunxiao.ListRepositoriesRequest
	listRepositoriesCalled            bool
	getRepositoryCalled               bool
	getRepositoryRequest              yunxiao.GetRepositoryRequest
	createRepositoryRequest           yunxiao.CreateRepositoryRequest
	updateRepositoryRequest           yunxiao.UpdateRepositoryRequest
	archiveRepositoryRequest          yunxiao.RepositoryActionRequest
	deleteRepositoryRequest           yunxiao.RepositoryActionRequest
	listBranchesRequest               yunxiao.ListBranchesRequest
	listCommitsRequest                yunxiao.ListCommitsRequest
	getFileRequest                    yunxiao.GetFileRequest
	listMergeRequestsRequest          yunxiao.ListMergeRequestsRequest
	listPipelinesRequest              yunxiao.ListPipelinesRequest
	runPipelineRequest                yunxiao.RunPipelineRequest
	listRunsRequest                   yunxiao.ListRunsRequest
	getRunRequest                     yunxiao.GetRunRequest
	getRunLogRequest                  yunxiao.GetRunLogRequest
	watchRunRequest                   yunxiao.WatchRunRequest
	listSSHKeysRequest                yunxiao.ListSSHKeysRequest
	listWorkItemsRequest              yunxiao.ListWorkItemsRequest
	deleteWorkItemRequest             yunxiao.DeleteWorkItemRequest
	searchRepositoriesRequest         yunxiao.SearchRepositoriesRequest
	searchCodeRequest                 yunxiao.SearchCodeRequest
	searchCommitsRequest              yunxiao.SearchCommitsRequest
	searchMergeRequestsRequest        yunxiao.SearchMergeRequestsRequest
	commentMergeRequestRequest        yunxiao.CommentMergeRequestRequest
	resolveMergeRequestCommentRequest yunxiao.ResolveMergeRequestCommentRequest
	rawAPIRequest                     yunxiao.RawAPIRequest
}

type fakeCredentialStore struct{}

func (*fakeCredentialStore) Get(endpoint string) (auth.Credential, error) {
	return auth.Credential{Endpoint: endpoint}, nil
}

func (*fakeCredentialStore) Save(credential auth.Credential) (auth.CredentialLocation, error) {
	return auth.LocationFile, nil
}

func (*fakeCredentialStore) Delete(string) error {
	return nil
}

type fakeTokenVerifier struct{}

func (fakeTokenVerifier) VerifyToken(ctx context.Context, endpoint string, token string) (auth.AuthenticatedUser, []string, error) {
	return auth.AuthenticatedUser{ID: "user-1", DisplayName: "Test User"}, nil, nil
}

type storingCredentialStore struct {
	saved auth.Credential
}

func (s *storingCredentialStore) Get(endpoint string) (auth.Credential, error) {
	if s.saved.Endpoint == endpoint {
		return s.saved, nil
	}
	return auth.Credential{}, auth.ErrNotLoggedIn{Endpoint: endpoint}
}

func (s *storingCredentialStore) Save(credential auth.Credential) (auth.CredentialLocation, error) {
	s.saved = credential
	return auth.LocationFile, nil
}

func (s *storingCredentialStore) Delete(endpoint string) error {
	if s.saved.Endpoint != endpoint {
		return auth.ErrNotLoggedIn{Endpoint: endpoint}
	}
	s.saved = auth.Credential{}
	return nil
}

type recordingTokenVerifier struct {
	endpoint string
	token    string
}

func (v *recordingTokenVerifier) VerifyToken(ctx context.Context, endpoint string, token string) (auth.AuthenticatedUser, []string, error) {
	v.endpoint = endpoint
	v.token = token
	return auth.AuthenticatedUser{ID: "user-1", DisplayName: "Test User"}, nil, nil
}

func (f *fakeCommandService) ListRepositories(ctx context.Context, request yunxiao.ListRepositoriesRequest) (yunxiao.RepositoryListResult, error) {
	f.listRepositoriesCalled = true
	f.listRepositoriesRequest = request
	return yunxiao.RepositoryListResult{Repositories: []yunxiao.Repository{
		{ID: "repo-1", Name: "api", Path: "cro/api", DefaultBranch: "master", Visibility: "private"},
		{ID: "repo-2", Name: "test-repo", Path: "6a49f26608f52788b13355d2/test-repo", Visibility: "internal"},
	}}, nil
}

func (f *fakeCommandService) GetRepository(ctx context.Context, request yunxiao.GetRepositoryRequest) (yunxiao.RepositoryResult, error) {
	f.getRepositoryCalled = true
	f.getRepositoryRequest = request
	return yunxiao.RepositoryResult{Repository: yunxiao.Repository{
		ID:            request.RepositoryID,
		Name:          "test-repo",
		Path:          request.Organization + "/test-repo",
		DefaultBranch: "master",
		SSHURL:        "git@codeup.aliyun.com:" + request.Organization + "/test-repo.git",
		HTTPURL:       "https://codeup.aliyun.com/" + request.Organization + "/test-repo.git",
	}}, nil
}

func (f *fakeCommandService) CreateRepository(ctx context.Context, request yunxiao.CreateRepositoryRequest) (yunxiao.RepositoryResult, error) {
	f.createRepositoryRequest = request
	return yunxiao.RepositoryResult{Repository: yunxiao.Repository{ID: "repo-new", Name: request.Name}}, nil
}

func (f *fakeCommandService) UpdateRepository(ctx context.Context, request yunxiao.UpdateRepositoryRequest) (yunxiao.RepositoryResult, error) {
	f.updateRepositoryRequest = request
	return yunxiao.RepositoryResult{Repository: yunxiao.Repository{ID: request.RepositoryID, Name: request.Name}}, nil
}

func (f *fakeCommandService) ArchiveRepository(ctx context.Context, request yunxiao.RepositoryActionRequest) (yunxiao.RepositoryActionResult, error) {
	f.archiveRepositoryRequest = request
	return yunxiao.RepositoryActionResult{Repository: yunxiao.Repository{ID: request.RepositoryID}, Supported: true}, nil
}

func (f *fakeCommandService) UnarchiveRepository(ctx context.Context, request yunxiao.RepositoryActionRequest) (yunxiao.RepositoryActionResult, error) {
	return yunxiao.RepositoryActionResult{Supported: false}, nil
}

func (f *fakeCommandService) DeleteRepository(ctx context.Context, request yunxiao.RepositoryActionRequest) (yunxiao.RepositoryActionResult, error) {
	f.deleteRepositoryRequest = request
	return yunxiao.RepositoryActionResult{Repository: yunxiao.Repository{ID: request.RepositoryID}, Supported: true}, nil
}

func (f *fakeCommandService) ListBranches(ctx context.Context, request yunxiao.ListBranchesRequest) (yunxiao.BranchListResult, error) {
	f.listBranchesRequest = request
	return yunxiao.BranchListResult{Branches: []yunxiao.Branch{{Name: "master", Default: true}}}, nil
}

func (f *fakeCommandService) GetBranch(ctx context.Context, request yunxiao.GetBranchRequest) (yunxiao.BranchResult, error) {
	return yunxiao.BranchResult{Branch: yunxiao.Branch{Name: request.Branch, Default: request.Branch == "master"}}, nil
}

func (f *fakeCommandService) ListCommits(ctx context.Context, request yunxiao.ListCommitsRequest) (yunxiao.CommitListResult, error) {
	f.listCommitsRequest = request
	return yunxiao.CommitListResult{Commits: []yunxiao.Commit{{SHA: "abc123", ShortSHA: "abc123", AuthorName: "gouziya", Date: "2026-07-05T14:00:02+08:00", Title: "Initial commit"}}}, nil
}

func (f *fakeCommandService) GetCommit(ctx context.Context, request yunxiao.GetCommitRequest) (yunxiao.CommitResult, error) {
	return yunxiao.CommitResult{Commit: yunxiao.Commit{SHA: request.SHA, Title: "Initial commit"}}, nil
}

func (f *fakeCommandService) GetFile(ctx context.Context, request yunxiao.GetFileRequest) (yunxiao.FileResult, error) {
	f.getFileRequest = request
	return yunxiao.FileResult{File: yunxiao.FileEntry{Path: request.Path, Ref: request.Ref, Encoding: "base64", Content: "IyMg5rWL6K+V5LuT5bqT"}}, nil
}

func (f *fakeCommandService) ListFiles(ctx context.Context, request yunxiao.ListFilesRequest) (yunxiao.FileTreeResult, error) {
	return yunxiao.FileTreeResult{Files: []yunxiao.FileEntry{{Type: "blob", Path: "README.md"}}}, nil
}

func (f *fakeCommandService) ListMergeRequests(ctx context.Context, request yunxiao.ListMergeRequestsRequest) (yunxiao.MergeRequestListResult, error) {
	f.listMergeRequestsRequest = request
	return yunxiao.MergeRequestListResult{MergeRequests: []yunxiao.MergeRequest{{ID: "1", IID: "1", Title: "Add CLI", State: "open", SourceBranch: "feature", TargetBranch: "master"}}}, nil
}

func (f *fakeCommandService) GetMergeRequest(ctx context.Context, request yunxiao.GetMergeRequestRequest) (yunxiao.MergeRequestResult, error) {
	return yunxiao.MergeRequestResult{MergeRequest: yunxiao.MergeRequest{ID: request.MergeRequestID, IID: request.MergeRequestID, Title: "Add CLI"}}, nil
}

func (f *fakeCommandService) CreateMergeRequest(ctx context.Context, request yunxiao.CreateMergeRequestRequest) (yunxiao.MergeRequestResult, error) {
	return yunxiao.MergeRequestResult{MergeRequest: yunxiao.MergeRequest{ID: "1", IID: "1", Title: request.Title, SourceBranch: request.SourceBranch, TargetBranch: request.TargetBranch}}, nil
}

func (f *fakeCommandService) UpdateMergeRequest(ctx context.Context, request yunxiao.UpdateMergeRequestRequest) (yunxiao.MergeRequestResult, error) {
	return yunxiao.MergeRequestResult{MergeRequest: yunxiao.MergeRequest{ID: request.MergeRequestID, IID: request.MergeRequestID, Title: request.Title}}, nil
}

func (f *fakeCommandService) GetMergeRequestDiff(ctx context.Context, request yunxiao.GetMergeRequestDiffRequest) (yunxiao.MergeRequestDiffResult, error) {
	return yunxiao.MergeRequestDiffResult{Diff: "diff --git a b\n"}, nil
}

func (f *fakeCommandService) ListMergeRequestFiles(ctx context.Context, request yunxiao.ListMergeRequestFilesRequest) (yunxiao.MergeRequestFileListResult, error) {
	return yunxiao.MergeRequestFileListResult{Files: []yunxiao.MergeRequestFile{{Path: "main.go", Status: "modified"}}}, nil
}

func (f *fakeCommandService) CommentMergeRequest(ctx context.Context, request yunxiao.CommentMergeRequestRequest) (yunxiao.CommentResult, error) {
	f.commentMergeRequestRequest = request
	return yunxiao.CommentResult{ID: "comment-1"}, nil
}

func (f *fakeCommandService) ResolveMergeRequestComment(ctx context.Context, request yunxiao.ResolveMergeRequestCommentRequest) (yunxiao.CommentResult, error) {
	f.resolveMergeRequestCommentRequest = request
	return yunxiao.CommentResult{ID: request.CommentID}, nil
}

func (f *fakeCommandService) ReviewMergeRequest(ctx context.Context, request yunxiao.ReviewMergeRequestRequest) (yunxiao.ReviewResult, error) {
	return yunxiao.ReviewResult{MergeRequestID: request.MergeRequestID, ReviewStatus: string(request.Decision)}, nil
}

func (f *fakeCommandService) MergeMergeRequest(ctx context.Context, request yunxiao.MergeRequestActionRequest) (yunxiao.MergeRequestActionResult, error) {
	return yunxiao.MergeRequestActionResult{MergeRequest: yunxiao.MergeRequest{ID: request.MergeRequestID, IID: request.MergeRequestID}}, nil
}

func (f *fakeCommandService) CloseMergeRequest(ctx context.Context, request yunxiao.MergeRequestActionRequest) (yunxiao.MergeRequestActionResult, error) {
	return yunxiao.MergeRequestActionResult{MergeRequest: yunxiao.MergeRequest{ID: request.MergeRequestID, IID: request.MergeRequestID}}, nil
}

func (f *fakeCommandService) ReopenMergeRequest(ctx context.Context, request yunxiao.MergeRequestActionRequest) (yunxiao.MergeRequestActionResult, error) {
	return yunxiao.MergeRequestActionResult{MergeRequest: yunxiao.MergeRequest{ID: request.MergeRequestID, IID: request.MergeRequestID}}, nil
}

func (f *fakeCommandService) GetMergeRequestStatus(ctx context.Context, request yunxiao.GetMergeRequestRequest) (yunxiao.MergeRequestStatusResult, error) {
	return yunxiao.MergeRequestStatusResult{MergeRequest: yunxiao.MergeRequest{ID: request.MergeRequestID, IID: request.MergeRequestID, State: "open"}}, nil
}

func (f *fakeCommandService) ListPipelines(ctx context.Context, request yunxiao.ListPipelinesRequest) (yunxiao.PipelineListResult, error) {
	f.listPipelinesRequest = request
	return yunxiao.PipelineListResult{Pipelines: []yunxiao.Pipeline{{ID: "pipe-1", Name: "build"}}}, nil
}

func (f *fakeCommandService) GetPipeline(ctx context.Context, request yunxiao.GetPipelineRequest) (yunxiao.PipelineResult, error) {
	return yunxiao.PipelineResult{Pipeline: yunxiao.Pipeline{ID: request.PipelineID, Name: "build"}}, nil
}

func (f *fakeCommandService) RunPipeline(ctx context.Context, request yunxiao.RunPipelineRequest) (yunxiao.PipelineRunStartResult, error) {
	f.runPipelineRequest = request
	return yunxiao.PipelineRunStartResult{Run: yunxiao.Run{ID: "run-1", Branch: request.Branch, WebURL: "https://example.test/run-1"}}, nil
}

func (f *fakeCommandService) ListRuns(ctx context.Context, request yunxiao.ListRunsRequest) (yunxiao.RunListResult, error) {
	f.listRunsRequest = request
	return yunxiao.RunListResult{Runs: []yunxiao.Run{{ID: "run-1", PipelineID: request.PipelineID, Pipeline: "build", Branch: "master", Status: "SUCCESS"}}}, nil
}

func (f *fakeCommandService) GetRun(ctx context.Context, request yunxiao.GetRunRequest) (yunxiao.RunResult, error) {
	f.getRunRequest = request
	return yunxiao.RunResult{Run: yunxiao.Run{ID: request.RunID, Status: "SUCCESS"}}, nil
}

func (f *fakeCommandService) GetRunLog(ctx context.Context, request yunxiao.GetRunLogRequest) (yunxiao.RunLogResult, error) {
	f.getRunLogRequest = request
	return yunxiao.RunLogResult{RunID: request.RunID, Lines: "ok"}, nil
}

func (f *fakeCommandService) WatchRun(ctx context.Context, request yunxiao.WatchRunRequest) (yunxiao.RunResult, error) {
	f.watchRunRequest = request
	return yunxiao.RunResult{Run: yunxiao.Run{ID: request.RunID, Status: "SUCCESS"}}, nil
}

func (f *fakeCommandService) CancelRun(ctx context.Context, request yunxiao.RunActionRequest) (yunxiao.RunActionResult, error) {
	return yunxiao.RunActionResult{RunID: request.RunID, Action: "cancel"}, nil
}

func (f *fakeCommandService) RetryRun(ctx context.Context, request yunxiao.RunActionRequest) (yunxiao.RunActionResult, error) {
	return yunxiao.RunActionResult{RunID: request.RunID, Action: "retry"}, nil
}

func (f *fakeCommandService) RetryTask(ctx context.Context, request yunxiao.RunTaskActionRequest) (yunxiao.RunActionResult, error) {
	return yunxiao.RunActionResult{RunID: request.RunID, JobID: request.JobID, Action: "retry-task"}, nil
}

func (f *fakeCommandService) StopTask(ctx context.Context, request yunxiao.RunTaskActionRequest) (yunxiao.RunActionResult, error) {
	return yunxiao.RunActionResult{RunID: request.RunID, JobID: request.JobID, Action: "stop-task"}, nil
}

func (f *fakeCommandService) SkipTask(ctx context.Context, request yunxiao.RunTaskActionRequest) (yunxiao.RunActionResult, error) {
	return yunxiao.RunActionResult{RunID: request.RunID, JobID: request.JobID, Action: "skip-task"}, nil
}

func (f *fakeCommandService) ListSSHKeys(ctx context.Context, request yunxiao.ListSSHKeysRequest) (yunxiao.SSHKeyListResult, error) {
	f.listSSHKeysRequest = request
	return yunxiao.SSHKeyListResult{Keys: []yunxiao.SSHKey{{ID: "key-1", Title: "a100 vm", Fingerprint: "fp", CreatedAt: "2026-07-05T14:00:02+08:00"}}}, nil
}

func (f *fakeCommandService) CreateSSHKey(ctx context.Context, request yunxiao.CreateSSHKeyRequest) (yunxiao.SSHKeyResult, error) {
	return yunxiao.SSHKeyResult{Key: yunxiao.SSHKey{ID: "key-1", Title: request.Title, Fingerprint: "fp"}}, nil
}

func (f *fakeCommandService) DeleteSSHKey(ctx context.Context, request yunxiao.DeleteSSHKeyRequest) (yunxiao.SSHKeyActionResult, error) {
	return yunxiao.SSHKeyActionResult{ID: request.ID}, nil
}

func (f *fakeCommandService) ListWorkItems(ctx context.Context, request yunxiao.ListWorkItemsRequest) (yunxiao.WorkItemListResult, error) {
	f.listWorkItemsRequest = request
	return yunxiao.WorkItemListResult{WorkItems: []yunxiao.WorkItem{{ID: "WI-1", Type: "Bug", State: "Open", Title: "Fix login"}}}, nil
}

func (f *fakeCommandService) GetWorkItem(ctx context.Context, request yunxiao.GetWorkItemRequest) (yunxiao.WorkItemResult, error) {
	return yunxiao.WorkItemResult{WorkItem: yunxiao.WorkItem{ID: request.ID, Title: "Fix login"}}, nil
}

func (f *fakeCommandService) CreateWorkItem(ctx context.Context, request yunxiao.CreateWorkItemRequest) (yunxiao.WorkItemResult, error) {
	return yunxiao.WorkItemResult{WorkItem: yunxiao.WorkItem{ID: "WI-1", Type: request.Type, Title: request.Title}}, nil
}

func (f *fakeCommandService) UpdateWorkItem(ctx context.Context, request yunxiao.UpdateWorkItemRequest) (yunxiao.WorkItemResult, error) {
	return yunxiao.WorkItemResult{WorkItem: yunxiao.WorkItem{ID: request.ID, Title: request.Title}}, nil
}

func (f *fakeCommandService) DeleteWorkItem(ctx context.Context, request yunxiao.DeleteWorkItemRequest) (yunxiao.WorkItemActionResult, error) {
	f.deleteWorkItemRequest = request
	return yunxiao.WorkItemActionResult{ID: request.ID}, nil
}

func (f *fakeCommandService) ListWorkItemActivities(ctx context.Context, request yunxiao.ListWorkItemActivitiesRequest) (yunxiao.WorkItemActivityListResult, error) {
	return yunxiao.WorkItemActivityListResult{Activities: []yunxiao.WorkItemActivity{{Actor: "gouzi", Action: "created"}}}, nil
}

func (f *fakeCommandService) SearchRepositories(ctx context.Context, request yunxiao.SearchRepositoriesRequest) (yunxiao.SearchRepositoryResult, error) {
	f.searchRepositoriesRequest = request
	return yunxiao.SearchRepositoryResult{Repositories: []yunxiao.Repository{{ID: "repo-1", Name: "api", Path: "cro/api"}}}, nil
}

func (f *fakeCommandService) SearchCode(ctx context.Context, request yunxiao.SearchCodeRequest) (yunxiao.SearchCodeResult, error) {
	f.searchCodeRequest = request
	return yunxiao.SearchCodeResult{Matches: []yunxiao.CodeSearchMatch{{Repository: request.RepositoryID, Path: "main.go", Match: request.Query}}}, nil
}

func (f *fakeCommandService) SearchCommits(ctx context.Context, request yunxiao.SearchCommitsRequest) (yunxiao.SearchCommitResult, error) {
	f.searchCommitsRequest = request
	return yunxiao.SearchCommitResult{Matches: []yunxiao.CommitSearchMatch{{Repository: request.RepositoryID, Commit: yunxiao.Commit{SHA: "abc1234", Title: request.Query}}}}, nil
}

func (f *fakeCommandService) SearchMergeRequests(ctx context.Context, request yunxiao.SearchMergeRequestsRequest) (yunxiao.SearchMergeRequestResult, error) {
	f.searchMergeRequestsRequest = request
	return yunxiao.SearchMergeRequestResult{Matches: []yunxiao.MergeRequestSearchMatch{{Repository: request.RepositoryID, MergeRequest: yunxiao.MergeRequest{ID: "1", Title: request.Query}}}}, nil
}

func (f *fakeCommandService) SearchWorkItems(ctx context.Context, request yunxiao.SearchWorkItemsRequest) (yunxiao.SearchWorkItemResult, error) {
	return yunxiao.SearchWorkItemResult{Matches: []yunxiao.WorkItemSearchMatch{{WorkItem: yunxiao.WorkItem{ID: "WI-1", Title: request.Query}}}}, nil
}

func (f *fakeCommandService) Request(ctx context.Context, request yunxiao.RawAPIRequest) (yunxiao.RawAPIResult, error) {
	f.rawAPIRequest = request
	return yunxiao.RawAPIResult{Method: request.Method, URL: request.Path, Status: 200, RequestID: "req-1", Body: `{"data":{"id":"42"},"success":true}`}, nil
}

type fakeExtensionManager struct {
	extensions        []extension.Extension
	dispatchedName    string
	dispatchedArgs    []string
	dispatchedContext extension.ExecutionContext
}

func (f *fakeExtensionManager) List() ([]extension.Extension, error) {
	return append([]extension.Extension(nil), f.extensions...), nil
}

func (f *fakeExtensionManager) Install(ctx context.Context, source extension.InstallSource) (extension.Extension, error) {
	return extension.Extension{Name: "installed", FullName: "yunxiao-installed", Kind: extension.KindGit}, nil
}

func (f *fakeExtensionManager) InstallLocal(ctx context.Context, dir string) (extension.Extension, error) {
	return extension.Extension{Name: "local", FullName: "yunxiao-local", Kind: extension.KindLocal}, nil
}

func (f *fakeExtensionManager) Upgrade(ctx context.Context, name string, force bool) error {
	return nil
}

func (f *fakeExtensionManager) Remove(name string) error {
	return nil
}

func (f *fakeExtensionManager) Dispatch(ctx context.Context, request extension.DispatchRequest) (bool, error) {
	f.dispatchedName = request.Name
	f.dispatchedArgs = append([]string(nil), request.Args...)
	f.dispatchedContext = request.Context
	for _, ext := range f.extensions {
		if ext.Name == request.Name {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeExtensionManager) Create(ctx context.Context, name string, template extension.TemplateType) error {
	return nil
}

func (f *fakeExtensionManager) UpdateDir(name string) string {
	return ""
}
