## API 映射原则

本文档记录每个规划中的 CLI 命令需要调用的云效 API 及官方文档链接。实现阶段必须优先使用新版 API 参考；只有新版未覆盖的能力，才标注并使用旧版 API。

基础文档：

- API 参考入口：https://help.aliyun.com/zh/yunxiao/developer-reference/api-reference-standard-proprietary/
- 服务接入点：https://help.aliyun.com/zh/yunxiao/developer-reference/service-access-point-domain
- 获取个人访问令牌：https://help.aliyun.com/zh/yunxiao/developer-reference/obtain-personal-access-token
- 错误码中心：https://help.aliyun.com/zh/yunxiao/developer-reference/error-code-center

## 1. auth / config / api

| CLI | 所需 API / 文档 | 备注 |
| --- | --- | --- |
| `yunxiao auth login` | 获取个人访问令牌：https://help.aliyun.com/zh/yunxiao/developer-reference/obtain-personal-access-token；GetUserByToken：https://help.aliyun.com/zh/yunxiao/developer-reference/getuserbytoken | PAT 由用户在云效页面创建；CLI 使用 `GetUserByToken` 校验令牌并获取当前用户信息。 |
| `yunxiao auth status` | GetUserByToken：https://help.aliyun.com/zh/yunxiao/developer-reference/getuserbytoken | 用已保存 token 查询当前账号。 |
| `yunxiao auth logout` | 本地实现 | 删除本地凭据，不调用云效 API。 |
| `yunxiao config get/list/set` | 本地实现；服务接入点文档：https://help.aliyun.com/zh/yunxiao/developer-reference/service-access-point-domain | 管理 endpoint、organization、default repo 等本地配置。 |
| `yunxiao api <method> <path>` | 通用 HTTP 调用；API 参考入口：https://help.aliyun.com/zh/yunxiao/developer-reference/api-reference-standard-proprietary/ | 直接调用用户指定的云效 OpenAPI 路径。 |

## 2. repo / branch / commit / file / ssh-key

| CLI | 所需 API / 文档 | 备注 |
| --- | --- | --- |
| `yunxiao repo list` | ListRepositories：https://help.aliyun.com/zh/yunxiao/developer-reference/listrepositories-query-code-base-list | 列出代码库。 |
| `yunxiao repo view` | GetRepository：https://help.aliyun.com/zh/yunxiao/developer-reference/getrepository-query-the-code-base | 查看代码库详情。 |
| `yunxiao repo create` | CreateRepository：https://help.aliyun.com/zh/yunxiao/developer-reference/createreposition-creates-a-code-base | 创建代码库。 |
| `yunxiao repo update` | UpdateRepository：https://help.aliyun.com/zh/yunxiao/developer-reference/updatereposition-update-the-code-base | 更新代码库基础信息。 |
| `yunxiao repo archive` | ArchiveRepository：https://help.aliyun.com/zh/yunxiao/developer-reference/archivereposition-archive-the-code-base | 归档代码库。 |
| `yunxiao repo unarchive` | 待确认 | 新版目录只明确列出 `ArchiveRepository`，未看到取消归档接口；实现前需要确认是否由 `UpdateRepository` 或其他接口承载。 |
| `yunxiao repo delete` | DeleteRepository：https://help.aliyun.com/zh/yunxiao/developer-reference/deletereposition-delete-code-base | 删除代码库。 |
| `yunxiao repo transfer` | TransferRepository：https://help.aliyun.com/zh/yunxiao/developer-reference/transfer-repository-transfer-repository | 转移代码库，可作为后续命令。 |
| `yunxiao repo set-default` | 本地实现；GetRepository：https://help.aliyun.com/zh/yunxiao/developer-reference/getrepository-query-the-code-base | 本地保存默认代码库前可调用 `GetRepository` 校验。 |
| `yunxiao branch list` | ListBranches：https://help.aliyun.com/zh/yunxiao/developer-reference/listbranches-query-the-list-of-branches | 查询分支列表。 |
| `yunxiao branch view` | GetBranch：https://help.aliyun.com/zh/yunxiao/developer-reference/getbranch-query-branch-information | 查询分支详情。 |
| `yunxiao branch create` | CreateBranch：https://help.aliyun.com/zh/yunxiao/developer-reference/createbranch-create-branch | 可作为后续命令。 |
| `yunxiao branch delete` | DeleteBranch：https://help.aliyun.com/zh/yunxiao/developer-reference/deletebranch-delete-branch | 可作为后续命令。 |
| `yunxiao commit list` | ListCommits：https://help.aliyun.com/zh/yunxiao/developer-reference/listcommits-query-the-submission-list | 查询提交列表。 |
| `yunxiao commit view` | GetCommit：https://help.aliyun.com/zh/yunxiao/developer-reference/getcommit-query-commit-information；ListCommitStatuses：https://help.aliyun.com/zh/yunxiao/developer-reference/listcommitstatuses-query-the-submission-status-list | 查看提交详情及提交状态。 |
| `yunxiao commit comment` | CreateCommitComment：https://help.aliyun.com/zh/yunxiao/developer-reference/createcommitcomment | 可作为后续命令。 |
| `yunxiao file tree` | ListFiles：https://help.aliyun.com/zh/yunxiao/developer-reference/listfiles | 查询文件树。 |
| `yunxiao file view` | GetFileBlobs：https://help.aliyun.com/zh/yunxiao/developer-reference/getfileblobs | 查询文件内容。 |
| `yunxiao file blame` | GetFileBlame：https://help.aliyun.com/zh/yunxiao/developer-reference/getfileblame-get-file-blame-information | 可作为后续命令。 |
| `yunxiao file create/update/delete` | CreateFile：https://help.aliyun.com/zh/yunxiao/developer-reference/createfile；UpdateFile：https://help.aliyun.com/zh/yunxiao/developer-reference/updatefile；DeleteFile：https://help.aliyun.com/zh/yunxiao/developer-reference/deletefile | 可作为后续命令。 |
| `yunxiao file commit-multiple` | CommitMultipleFiles：https://help.aliyun.com/zh/yunxiao/developer-reference/commitmultiplefiles-multi-file-change-commit | 多文件变更提交。 |
| `yunxiao ssh-key list` | ListSSHKeys：https://help.aliyun.com/zh/yunxiao/developer-reference/listsshkeys-query-the-ssh-key-list；ListUserSSHKeys：https://help.aliyun.com/zh/yunxiao/developer-reference/listusersshkeys | 当前用户或指定用户 SSH Key 列表。 |
| `yunxiao ssh-key view` | GetSSHKey：https://help.aliyun.com/zh/yunxiao/developer-reference/getsshkey-query-the-ssh-key | 查询 SSH Key。 |
| `yunxiao ssh-key add` | CreateSSHKey：https://help.aliyun.com/zh/yunxiao/developer-reference/createsshkey-create-an-ssh-key | 创建 SSH Key。 |
| `yunxiao ssh-key delete` | DeleteSSHKey：https://help.aliyun.com/zh/yunxiao/developer-reference/deletesshkey-deletes-the-ssh-key | 删除 SSH Key。 |

## 3. mr

| CLI | 所需 API / 文档 | 备注 |
| --- | --- | --- |
| `yunxiao mr list` | ListChangeRequests：https://help.aliyun.com/zh/yunxiao/developer-reference/listchangerequests-query-the-list-of-merge-requests | 查询合并请求列表。 |
| `yunxiao mr view` | GetChangeRequest：https://help.aliyun.com/zh/yunxiao/developer-reference/getchangerequest-query-merge-request；GetChangeRequestLabels：https://help.aliyun.com/zh/yunxiao/developer-reference/getchangerequestlabels；ListCheckRuns：https://help.aliyun.com/zh/yunxiao/developer-reference/listcheckruns | 详情页可聚合类标和运行检查。 |
| `yunxiao mr create` | CreateChangeRequest：https://help.aliyun.com/zh/yunxiao/developer-reference/createchangerequest-create-merge-request | 创建合并请求。 |
| `yunxiao mr edit` | UpdateChangeRequest：https://help.aliyun.com/zh/yunxiao/developer-reference/updatechangerequest-update-merge-request-basic-information；UpdateChangeRequestRelatedPerson：https://help.aliyun.com/zh/yunxiao/developer-reference/update-merge-request-stakeholders | 更新基础信息和干系人。 |
| `yunxiao mr close` | CloseChangeRequest：https://help.aliyun.com/zh/yunxiao/developer-reference/closechangerequest-close-merge-request | 关闭合并请求。 |
| `yunxiao mr reopen` | ReopenChangeRequest：https://help.aliyun.com/zh/yunxiao/developer-reference/reopenchangerequest-reopen-merge-request | 重新打开合并请求。 |
| `yunxiao mr merge` | MergeChangeRequest：https://help.aliyun.com/zh/yunxiao/developer-reference/mergechangerequest-merge-merge-request | 合并合并请求。 |
| `yunxiao mr files` | GetChangeRequestTree：https://help.aliyun.com/zh/yunxiao/developer-reference/getchangerequesttree-queries-the-change-file-tree-for-merge-requests | 查询变更文件树。 |
| `yunxiao mr diff` | GetCompare：https://help.aliyun.com/zh/yunxiao/developer-reference/getcompare；ListChangeRequestPatchSets：https://help.aliyun.com/zh/yunxiao/developer-reference/listchangerequestpatchsets-query-the-list-of-merge-request-versions | 先取 MR 版本，再用代码比较接口生成 diff。 |
| `yunxiao mr comment` | CreateChangeRequestComment：https://help.aliyun.com/zh/yunxiao/developer-reference/createchangerequestcomment；ListMergeRequestComments：https://help.aliyun.com/zh/yunxiao/developer-reference/listmergerequestcomments | 创建评论，必要时查询评论列表。 |
| `yunxiao mr comment edit/delete` | UpdateChangeRequestComment：https://help.aliyun.com/zh/yunxiao/developer-reference/updatechangerequestcomment；DeleteChangeRequestComment：https://help.aliyun.com/zh/yunxiao/developer-reference/deletechangerequestcomment | 可作为后续命令。 |
| `yunxiao mr approve` | ReviewChangeRequest：https://help.aliyun.com/zh/yunxiao/developer-reference/reviewchangerequest-review-merge-request | 通过评审。 |
| `yunxiao mr changes-requested` | ReviewChangeRequest：https://help.aliyun.com/zh/yunxiao/developer-reference/reviewchangerequest-review-merge-request | 要求修改。 |
| `yunxiao mr labels attach/list` | AttachLabelsToChangeRequest：https://help.aliyun.com/zh/yunxiao/developer-reference/attachlabelstochangerquest；GetChangeRequestLabels：https://help.aliyun.com/zh/yunxiao/developer-reference/getchangerequestlabels | 可作为后续命令。 |
| `yunxiao mr status` | GetChangeRequest：https://help.aliyun.com/zh/yunxiao/developer-reference/getchangerequest-query-merge-request；ListCheckRuns：https://help.aliyun.com/zh/yunxiao/developer-reference/listcheckruns；ListCommitStatuses：https://help.aliyun.com/zh/yunxiao/developer-reference/listcommitstatuses-query-the-submission-status-list | 聚合可合并、运行检查和提交状态。 |

## 4. pipeline / run

| CLI | 所需 API / 文档 | 备注 |
| --- | --- | --- |
| `yunxiao pipeline list` | ListPipelines：https://help.aliyun.com/zh/yunxiao/developer-reference/listpipelines-get-a-list-of-pipelines | 获取流水线列表。 |
| `yunxiao pipeline view` | GetPipeline：https://help.aliyun.com/zh/yunxiao/developer-reference/getpipeline-get-pipeline-details；ListPipelineRelations：https://help.aliyun.com/zh/yunxiao/developer-reference/listpipelinerelations；ListPipelineJobs：https://help.aliyun.com/zh/yunxiao/developer-reference/listpipelinejobs | 获取流水线详情、关联和任务。 |
| `yunxiao pipeline create` | CreatePipeline：https://help.aliyun.com/zh/yunxiao/developer-reference/createpipeline-create-pipeline | 创建流水线。 |
| `yunxiao pipeline update` | UpdatePipeline：https://help.aliyun.com/zh/yunxiao/developer-reference/updatepipeline-update-pipeline；UpdatePipelineBaseInfo：https://help.aliyun.com/zh/yunxiao/developer-reference/updatepipelinebaseinfo | 更新流水线。 |
| `yunxiao pipeline delete` | DeletePipeline：https://help.aliyun.com/zh/yunxiao/developer-reference/deletepipeline-delete-pipeline | 删除流水线。 |
| `yunxiao pipeline run` | CreatePipelineRun：https://help.aliyun.com/zh/yunxiao/developer-reference/createpipelinerun | 运行流水线。 |
| `yunxiao pipeline artifact-url` | GetPipelineArtifactUrl：https://help.aliyun.com/zh/yunxiao/developer-reference/getpipelineartifacturl；GetPipelineEmasArtifactUrl：https://help.aliyun.com/zh/yunxiao/developer-reference/getpipelineemasartifacturl | 可作为后续命令。 |
| `yunxiao run list` | ListPipelineRuns：https://help.aliyun.com/zh/yunxiao/developer-reference/listpipelineruns | 获取运行实例列表。 |
| `yunxiao run view` | GetPipelineRun：https://help.aliyun.com/zh/yunxiao/developer-reference/getpipelinerun；GetLatestPipelineRun：https://help.aliyun.com/zh/yunxiao/developer-reference/getlatestpipelinerun | 查看运行实例。 |
| `yunxiao run cancel` | UpdatePipelineRun：https://help.aliyun.com/zh/yunxiao/developer-reference/updatepipelinerun | 终止流水线运行。 |
| `yunxiao run log` | GetPipelineJobRunLog：https://help.aliyun.com/zh/yunxiao/developer-reference/getpipelinejobrunlog；GetPipelineJobSteps：https://help.aliyun.com/zh/yunxiao/developer-reference/getpipelinejobsteps；GetPipelineJobStepLog：https://help.aliyun.com/zh/yunxiao/developer-reference/getpipelinejobsteplog；GetPipelineJobStepLogUrl：https://help.aliyun.com/zh/yunxiao/developer-reference/getpipelinejobsteplogurl | 支持任务日志、步骤日志和下载地址。 |
| `yunxiao run watch` | GetPipelineRun：https://help.aliyun.com/zh/yunxiao/developer-reference/getpipelinerun；GetPipelineJobSteps：https://help.aliyun.com/zh/yunxiao/developer-reference/getpipelinejobsteps | 本地轮询实现。 |
| `yunxiao run retry` / `retry-task` | RetryPipelineJobRun：https://help.aliyun.com/zh/yunxiao/developer-reference/retrypipelinejobrun；RerunPipelineJobRun：https://help.aliyun.com/zh/yunxiao/developer-reference/rerunpipelinejobrun-reruns-pipeline-tasks | 任务重试或重新运行。 |
| `yunxiao run stop-task` | StopPipelineJobRun：https://help.aliyun.com/zh/yunxiao/developer-reference/stoppipelinejobrun | 终止任务运行。 |
| `yunxiao run skip-task` | SkipPipelineJobRun：https://help.aliyun.com/zh/yunxiao/developer-reference/skippipelinejobrun | 跳过任务运行。 |
| `yunxiao run execute-task` | ExecutePipelineJobRun：https://help.aliyun.com/zh/yunxiao/developer-reference/executepipelinejobrun；ExecutePipelineJobAction：https://help.aliyun.com/zh/yunxiao/developer-reference/executepipipelinejobaction | 可作为后续命令。 |
| `yunxiao run validate pass/refuse` | PassPipelineValidate：https://help.aliyun.com/zh/yunxiao/developer-reference/passpipelinevalidate；RefusePipelineValidate：https://help.aliyun.com/zh/yunxiao/developer-reference/refusepipelinevalidate | 人工卡点处理。 |

## 5. project / workitem

| CLI | 所需 API / 文档 | 备注 |
| --- | --- | --- |
| `yunxiao project list` | SearchProjects：https://help.aliyun.com/zh/yunxiao/developer-reference/searchprojects | 项目列表按搜索接口实现。 |
| `yunxiao project view` | GetProject：https://help.aliyun.com/zh/yunxiao/developer-reference/getproject | 查看项目。 |
| `yunxiao project create/update/delete` | CreateProject：https://help.aliyun.com/zh/yunxiao/developer-reference/createproject；UpdateProject：https://help.aliyun.com/zh/yunxiao/developer-reference/updateproject；DeleteProject：https://help.aliyun.com/zh/yunxiao/developer-reference/deleteproject | 可作为后续命令。 |
| `yunxiao project member list` | ListProjectMembers：https://help.aliyun.com/zh/yunxiao/developer-reference/listprojectmembers | 项目成员列表。 |
| `yunxiao project member add/delete` | CreateProjectMember：https://help.aliyun.com/zh/yunxiao/developer-reference/createprojectmember；DeleteProjectMember：https://help.aliyun.com/zh/yunxiao/developer-reference/deleteprojectmember | 可作为后续命令。 |
| `yunxiao project iteration list/view` | ListSprints：https://help.aliyun.com/zh/yunxiao/developer-reference/listsprints；GetSprint：https://help.aliyun.com/zh/yunxiao/developer-reference/getsprint | 云效迭代 API 命名为 sprint。 |
| `yunxiao project iteration create/update` | CreateSprint：https://help.aliyun.com/zh/yunxiao/developer-reference/createsprint；UpdateSprint：https://help.aliyun.com/zh/yunxiao/developer-reference/updatesprint | 可作为后续命令。 |
| `yunxiao project milestone list` | ListMilestones：https://help.aliyun.com/zh/yunxiao/developer-reference/listmilestones | 获取里程碑列表。 |
| `yunxiao project milestone create/update/delete` | CreateMilestone：https://help.aliyun.com/zh/yunxiao/developer-reference/createmilestone；UpdateMilestone：https://help.aliyun.com/zh/yunxiao/developer-reference/updatemilestone；DeleteMilestone：https://help.aliyun.com/zh/yunxiao/developer-reference/deletemilestone | 可作为后续命令。 |
| `yunxiao project label list` | ListLabels：https://help.aliyun.com/zh/yunxiao/developer-reference/listlabels | 项目协作标签。 |
| `yunxiao project label create/update` | CreateLabel：https://help.aliyun.com/zh/yunxiao/developer-reference/createlabel；UpdateLabel：https://help.aliyun.com/zh/yunxiao/developer-reference/updatelabel | 可作为后续命令。 |
| `yunxiao project version list` | ListVersions：https://help.aliyun.com/zh/yunxiao/developer-reference/listversions | 可作为后续命令。 |
| `yunxiao workitem list` | SearchWorkitems：https://help.aliyun.com/zh/yunxiao/developer-reference/searchworkitems | 搜索/列表工作项。 |
| `yunxiao workitem view` | GetWorkitem：https://help.aliyun.com/zh/yunxiao/developer-reference/getworkitem | 获取工作项。 |
| `yunxiao workitem create` | CreateWorkitem：https://help.aliyun.com/zh/yunxiao/developer-reference/createworkitem | 创建工作项。 |
| `yunxiao workitem edit` | UpdateWorkitem：https://help.aliyun.com/zh/yunxiao/developer-reference/updateworkitem | 更新工作项。 |
| `yunxiao workitem delete` | DeleteWorkitem：https://help.aliyun.com/zh/yunxiao/developer-reference/deleteworkitem | 删除工作项。 |
| `yunxiao workitem activity` | ListWorkitemActivities：https://help.aliyun.com/zh/yunxiao/developer-reference/listworkitemactivities | 获取工作项动态。 |
| `yunxiao workitem comments` | ListWorkitemComments：https://help.aliyun.com/zh/yunxiao/developer-reference/listworkitemcomments | 获取工作项评论。 |
| `yunxiao workitem comment` | CreateWorkitemComment：https://help.aliyun.com/zh/yunxiao/developer-reference/createworkitemcomment | 创建工作项评论。 |

## 6. search

| CLI | 所需 API / 文档 | 备注 |
| --- | --- | --- |
| `yunxiao search repo` | 新版：ListRepositories：https://help.aliyun.com/zh/yunxiao/developer-reference/listrepositories-query-code-base-list；旧版代码搜索目录：https://help.aliyun.com/zh/yunxiao/developer-reference/api-devops-2021-06-25-dir-code-search/；旧版 ListSearchRepository：https://help.aliyun.com/zh/yunxiao/developer-reference/api-devops-2021-06-25-listsearchrepository | 新版优先用代码库列表过滤；如需全文搜索语义，使用旧版搜索接口并在实现前验证可用性。 |
| `yunxiao search code` | 旧版代码搜索目录：https://help.aliyun.com/zh/yunxiao/developer-reference/api-devops-2021-06-25-dir-code-search/；ListSearchSourceCode：https://help.aliyun.com/zh/yunxiao/developer-reference/api-devops-2021-06-25-listsearchsourcecode；GetSearchCodePreview：https://help.aliyun.com/zh/yunxiao/developer-reference/api-devops-2021-06-25-getsearchcodepreview | 新版目录未见代码全文搜索接口；标注为旧版依赖，必须单独 smoke。 |
| `yunxiao search commit` | 新版：ListCommits：https://help.aliyun.com/zh/yunxiao/developer-reference/listcommits-query-the-submission-list；旧版 ListSearchCommit：https://help.aliyun.com/zh/yunxiao/developer-reference/api-devops-2021-06-25-listsearchcommit | 新版可做仓库内提交列表过滤；跨库搜索使用旧版接口。 |
| `yunxiao search mr` | ListChangeRequests：https://help.aliyun.com/zh/yunxiao/developer-reference/listchangerequests-query-the-list-of-merge-requests | 使用列表接口和过滤参数实现。 |
| `yunxiao search workitem` | SearchWorkitems：https://help.aliyun.com/zh/yunxiao/developer-reference/searchworkitems | 使用工作项搜索接口实现。 |
| `yunxiao search project` | SearchProjects：https://help.aliyun.com/zh/yunxiao/developer-reference/searchprojects；SearchPrograms：https://help.aliyun.com/zh/yunxiao/developer-reference/searchprograms-search-for-item-sets | 可作为后续命令。 |

## 7. output / completion / alias / browse

| CLI | 所需 API / 文档 | 备注 |
| --- | --- | --- |
| `--json` / `--jq` / `--template` | 本地实现 | 基于各命令返回的结构化模型渲染。 |
| `--web` | 本地实现；各资源 API 返回或拼接 Web URL | 无图形环境输出 URL。 |
| `yunxiao completion <shell>` | 本地实现 | 由 Cobra completion 生成。 |
| `yunxiao alias list/set/delete` | 本地实现 | 本地配置文件保存。 |

## 8. 暂不支持或后续单独规划

| gh 能力 | 云效对应情况 | 处理 |
| --- | --- | --- |
| `gh codespace` | 未在云效开发者 API 中发现对应产品面 | 不支持。 |
| `gh gist` | 未在云效开发者 API 中发现对应产品面 | 不支持。 |
| `gh release` | 云效有制品仓库、应用交付、流水线产物，但语义不同 | 后续按 `artifact`、`package` 或 `deploy` 单独规划。 |
| `gh attestation` | 未发现 GitHub Artifact Attestation 等价能力 | 不支持。 |
| `gh extension` / `gh skill` / `gh copilot` | 属于 GitHub CLI 或 GitHub 产品生态 | 不纳入云效服务端能力。 |
