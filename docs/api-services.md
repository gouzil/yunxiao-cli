# Yunxiao CLI Service 接口

本文档来自 OpenSpec API 映射和第一版 Go 实现边界。

每个命令都会调用 `internal/yunxiao` 中的类型化 service 接口。命令 handler 不传递松散的 `dict` 或 `map[string]any` 请求/结果对象。

## 本地服务

| 命令 | 接口 | 请求 / 结果 |
| --- | --- | --- |
| `auth login/status/logout` | `AuthService`, `auth.Manager` | `GetUserByTokenRequest`, `GetUserByTokenResult`, `LoginRequest`, `LoginResult` |
| `config get/list/set` | `config.Store` | `config.Values`, `config.Resolved`, `config.Entry` |
| `api <method> <path>` | `RawAPIService` | `RawAPIRequest`, `RawAPIResult` |
| `extension install/list/exec/upgrade/remove/create` | `extension.Manager` | `Extension`, `InstallSource`, `DispatchRequest`, `ExecutionContext` |

## 代码服务

| 命令组 | 接口 | 类型化模型 |
| --- | --- | --- |
| `repo` | `RepositoryService` | `Repository`, `ListRepositoriesRequest`, `RepositoryResult`, `RepositoryActionResult` |
| `branch` | `BranchService` | `Branch`, `ListBranchesRequest`, `BranchResult` |
| `commit` | `CommitService` | `Commit`, `CommitStatus`, `ListCommitsRequest`, `CommitResult` |
| `file` | `FileService` | `FileEntry`, `GetFileRequest`, `FileResult`, `FileTreeResult` |
| `ssh-key` | `SSHKeyService` | `SSHKey`, `CreateSSHKeyRequest`, `SSHKeyResult` |

## 协作服务

| 命令组 | 接口 | 类型化模型 |
| --- | --- | --- |
| `mr` | `MergeRequestService` | `MergeRequest`, `MergeRequestFile`, `MergeCheck`, `CreateMergeRequestRequest`, `MergeRequestStatusResult` |
| `pipeline` | `PipelineService` | `Pipeline`, `PipelineJob`, `RunPipelineRequest`, `PipelineRunStartResult` |
| `run` | `RunService` | `Run`, `GetRunLogRequest`, `WatchRunRequest`, `RunActionResult` |
| `project` | `ProjectService` | `Project`, `ProjectMember`, `Iteration`, `Milestone`, `Label` |
| `workitem` | `WorkItemService` | `WorkItem`, `WorkItemActivity`, `CreateWorkItemRequest`, `WorkItemResult` |
| `search` | `SearchService` | `SearchRepositoriesRequest`, `CodeSearchMatch`, `CommitSearchMatch`, `MergeRequestSearchMatch`, `WorkItemSearchMatch` |

## 待确认 API

OpenSpec API map 把 `repo unarchive` 和遗留跨仓库代码搜索 API 标记为待确认或 legacy。当前实现会显式保留这些状态：

- `repo unarchive` 返回类型化 `RepositoryActionResult{Supported:false}`，并打印待确认提示。
- `search code` 在第一版使用类型化文件列表路径，并通过 `SearchCodeResult.Legacy` 为后续 legacy API 云上验证留出位置。
