package output

import (
	"fmt"
	"strings"

	"github.com/gouzi/yunxiao-cli/internal/auth"
	"github.com/gouzi/yunxiao-cli/internal/config"
	"github.com/gouzi/yunxiao-cli/internal/yunxiao"
)

func AuthStatus(endpoint string, credential auth.Credential, loggedIn bool) Detail {
	if !loggedIn {
		return Detail{
			Title: endpoint,
			Fields: []DetailField{
				{Name: "Run", Value: "yunxiao auth login"},
			},
		}
	}
	return Detail{
		Title: endpoint,
		Fields: []DetailField{
			{Name: "Logged in", Value: fmt.Sprintf("%s account %s (%s)", endpoint, displayName(credential), credential.Location)},
			{Name: "Active account", Value: fmt.Sprint(credential.Active)},
			{Name: "User ID", Value: yunxiao.Unknown(credential.User.ID)},
			{Name: "Organization", Value: yunxiao.Unknown(credential.User.Organization)},
			{Name: "Git operations protocol", Value: "https"},
			{Name: "Token", Value: auth.MaskToken(credential.Token)},
			{Name: "Token scopes", Value: scopes(credential.Scopes)},
		},
	}
}

func LoginSuccess(result auth.LoginResult) Detail {
	credential := result.Credential
	return Detail{
		Title: fmt.Sprintf("✓ Logged in to %s account %s (%s)", credential.Endpoint, displayName(credential), result.Location),
		Fields: []DetailField{
			{Name: "Active account", Value: "true"},
			{Name: "Organization", Value: yunxiao.Unknown(credential.User.Organization)},
			{Name: "Git operations protocol", Value: "https"},
			{Name: "Token", Value: auth.MaskToken(credential.Token)},
			{Name: "Token scopes", Value: scopes(credential.Scopes)},
		},
	}
}

func ConfigTable(entries []config.Entry) Table {
	rows := make([]Row, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, Row{string(entry.Key), entry.Value, string(entry.Source)})
	}
	return Table{Headers: []string{"KEY", "VALUE", "SOURCE"}, Rows: rows}
}

func RepositoryTable(repositories []yunxiao.Repository) Table {
	rows := make([]Row, 0, len(repositories))
	for _, repo := range repositories {
		rows = append(rows, Row{repo.ID, repo.Name, repo.Path, repo.DefaultBranch, repo.Visibility, fmt.Sprint(repo.Archived), yunxiao.FormatTime(repo.UpdatedAt)})
	}
	return Table{Headers: []string{"ID", "NAME", "PATH", "DEFAULT BRANCH", "VISIBILITY", "ARCHIVED", "UPDATED"}, Rows: rows, Empty: "No repositories found."}
}

func RepositoryDetail(repo yunxiao.Repository) Detail {
	title := repo.Path
	if title == "" {
		title = repo.Name
	}
	return Detail{Title: title, Fields: []DetailField{
		{Name: "ID", Value: yunxiao.Unknown(repo.ID)},
		{Name: "Name", Value: yunxiao.Unknown(repo.Name)},
		{Name: "Path", Value: yunxiao.Unknown(repo.Path)},
		{Name: "Default branch", Value: yunxiao.Unknown(repo.DefaultBranch)},
		{Name: "Visibility", Value: yunxiao.Unknown(repo.Visibility)},
		{Name: "Archived", Value: fmt.Sprint(repo.Archived)},
		{Name: "Created", Value: yunxiao.FormatTime(repo.CreatedAt)},
		{Name: "Updated", Value: yunxiao.FormatTime(repo.UpdatedAt)},
		{Name: "Web URL", Value: yunxiao.Unknown(repo.WebURL)},
		{Name: "SSH URL", Value: yunxiao.Unknown(repo.SSHURL)},
		{Name: "HTTP URL", Value: yunxiao.Unknown(repo.HTTPURL)},
	}}
}

func BranchTable(branches []yunxiao.Branch) Table {
	rows := make([]Row, 0, len(branches))
	for _, branch := range branches {
		rows = append(rows, Row{branch.Name, fmt.Sprint(branch.Default), yunxiao.YesNoUnknown(branch.Protected), short(branch.CommitSHA), yunxiao.FormatTime(branch.UpdatedAt)})
	}
	return Table{Headers: []string{"NAME", "DEFAULT", "PROTECTED", "COMMIT", "UPDATED"}, Rows: rows, Empty: "No branches found."}
}

func CommitTable(commits []yunxiao.Commit) Table {
	rows := make([]Row, 0, len(commits))
	for _, commit := range commits {
		rows = append(rows, Row{short(firstNonEmpty(commit.ShortSHA, commit.SHA)), commit.AuthorName, yunxiao.FormatTime(commit.Date), commit.Title})
	}
	return Table{Headers: []string{"SHA", "AUTHOR", "DATE", "TITLE"}, Rows: rows, Empty: "No commits found."}
}

func FileTreeTable(files []yunxiao.FileEntry) Table {
	rows := make([]Row, 0, len(files))
	for _, file := range files {
		size := "-"
		if file.Type != "dir" && file.Size > 0 {
			size = fmt.Sprint(file.Size)
		}
		rows = append(rows, Row{file.Type, file.Path, size})
	}
	return Table{Headers: []string{"TYPE", "PATH", "SIZE"}, Rows: rows, Empty: "No files found."}
}

func SSHKeyTable(keys []yunxiao.SSHKey) Table {
	rows := make([]Row, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, Row{key.ID, key.Title, key.Fingerprint, yunxiao.FormatTime(key.CreatedAt)})
	}
	return Table{Headers: []string{"ID", "TITLE", "FINGERPRINT", "CREATED"}, Rows: rows, Empty: "No SSH keys found."}
}

func MergeRequestTable(mrs []yunxiao.MergeRequest) Table {
	rows := make([]Row, 0, len(mrs))
	for _, mr := range mrs {
		rows = append(rows, Row{mrNumber(mr), mr.State, mr.Title, mr.SourceBranch, mr.TargetBranch, mr.Author, yunxiao.FormatTime(mr.UpdatedAt)})
	}
	return Table{Headers: []string{"ID", "STATE", "TITLE", "SOURCE", "TARGET", "AUTHOR", "UPDATED"}, Rows: rows, Empty: "No merge requests found."}
}

func MergeRequestDetail(mr yunxiao.MergeRequest, description string) Detail {
	return Detail{
		Title: fmt.Sprintf("!%s %s", mrNumber(mr), mr.Title),
		Fields: []DetailField{
			{Name: "State", Value: yunxiao.Unknown(mr.State)},
			{Name: "Author", Value: yunxiao.Unknown(mr.Author)},
			{Name: "Source", Value: yunxiao.Unknown(mr.SourceBranch)},
			{Name: "Target", Value: yunxiao.Unknown(mr.TargetBranch)},
			{Name: "Review", Value: yunxiao.Unknown(mr.ReviewStatus)},
			{Name: "Mergeable", Value: yunxiao.YesNoUnknown(mr.Mergeable)},
			{Name: "Pipeline", Value: yunxiao.Unknown(mr.PipelineStatus)},
			{Name: "Labels", Value: labels(mr.Labels)},
			{Name: "Web URL", Value: yunxiao.Unknown(mr.WebURL)},
		},
		BodyHeading: "Description",
		Body:        description,
	}
}

func MRFileTable(files []yunxiao.MergeRequestFile) Table {
	rows := make([]Row, 0, len(files))
	for _, file := range files {
		rows = append(rows, Row{file.Status, fmt.Sprint(file.Additions), fmt.Sprint(file.Deletions), file.Path})
	}
	return Table{Headers: []string{"STATUS", "ADDITIONS", "DELETIONS", "PATH"}, Rows: rows, Empty: "No files found."}
}

func PipelineTable(pipelines []yunxiao.Pipeline) Table {
	rows := make([]Row, 0, len(pipelines))
	for _, pipeline := range pipelines {
		rows = append(rows, Row{pipeline.ID, pipeline.Name, pipeline.Status, yunxiao.FormatTime(pipeline.UpdatedAt)})
	}
	return Table{Headers: []string{"ID", "NAME", "STATUS", "UPDATED"}, Rows: rows, Empty: "No pipelines found."}
}

func RunTable(runs []yunxiao.Run) Table {
	rows := make([]Row, 0, len(runs))
	for _, run := range runs {
		rows = append(rows, Row{run.ID, run.Pipeline, run.Branch, run.Status, yunxiao.FormatTime(run.StartedAt), yunxiao.Unknown(run.Duration)})
	}
	return Table{Headers: []string{"ID", "PIPELINE", "BRANCH", "STATUS", "STARTED", "DURATION"}, Rows: rows, Empty: "No runs found."}
}

func ProjectTable(projects []yunxiao.Project) Table {
	rows := make([]Row, 0, len(projects))
	for _, project := range projects {
		rows = append(rows, Row{project.ID, project.Name, project.Status, project.Owner, yunxiao.FormatTime(project.UpdatedAt)})
	}
	return Table{Headers: []string{"ID", "NAME", "STATUS", "OWNER", "UPDATED"}, Rows: rows, Empty: "No projects found."}
}

func WorkItemTable(items []yunxiao.WorkItem) Table {
	rows := make([]Row, 0, len(items))
	for _, item := range items {
		rows = append(rows, Row{item.ID, item.Type, item.State, item.Title, firstNonEmpty(item.Assignee, "unassigned"), yunxiao.FormatTime(item.UpdatedAt)})
	}
	return Table{Headers: []string{"ID", "TYPE", "STATE", "TITLE", "ASSIGNEE", "UPDATED"}, Rows: rows, Empty: "No work items found."}
}

func WorkItemDetail(item yunxiao.WorkItem, description string) Detail {
	return Detail{
		Title: fmt.Sprintf("Work item %s: %s", item.ID, item.Title),
		Fields: []DetailField{
			{Name: "Type", Value: yunxiao.Unknown(item.Type)},
			{Name: "State", Value: yunxiao.Unknown(item.State)},
			{Name: "Assignee", Value: firstNonEmpty(item.Assignee, "unassigned")},
			{Name: "Reporter", Value: yunxiao.Unknown(item.Reporter)},
			{Name: "Project", Value: yunxiao.Unknown(item.ProjectID)},
			{Name: "Iteration", Value: firstNonEmpty(item.Iteration, "none")},
			{Name: "Priority", Value: yunxiao.Unknown(item.Priority)},
			{Name: "Updated", Value: yunxiao.FormatTime(item.UpdatedAt)},
			{Name: "Web URL", Value: yunxiao.Unknown(item.WebURL)},
		},
		BodyHeading: "Description",
		Body:        description,
	}
}

func IterationTable(iterations []yunxiao.Iteration) Table {
	rows := make([]Row, 0, len(iterations))
	for _, iteration := range iterations {
		rows = append(rows, Row{iteration.ID, iteration.Name, iteration.Status, iteration.StartDate, iteration.EndDate})
	}
	return Table{Headers: []string{"ID", "NAME", "STATUS", "START", "END"}, Rows: rows, Empty: "No iterations found."}
}

func MilestoneTable(milestones []yunxiao.Milestone) Table {
	rows := make([]Row, 0, len(milestones))
	for _, milestone := range milestones {
		rows = append(rows, Row{milestone.ID, milestone.Name, milestone.Status, milestone.DueDate})
	}
	return Table{Headers: []string{"ID", "NAME", "STATUS", "DUE"}, Rows: rows, Empty: "No milestones found."}
}

func LabelTable(labels []yunxiao.Label) Table {
	rows := make([]Row, 0, len(labels))
	for _, label := range labels {
		rows = append(rows, Row{label.ID, label.Name, label.Color, label.Description})
	}
	return Table{Headers: []string{"ID", "NAME", "COLOR", "DESCRIPTION"}, Rows: rows, Empty: "No labels found."}
}

func ActivityTable(activities []yunxiao.WorkItemActivity) Table {
	rows := make([]Row, 0, len(activities))
	for _, activity := range activities {
		rows = append(rows, Row{yunxiao.FormatTime(activity.Time), activity.Actor, activity.Action})
	}
	return Table{Headers: []string{"TIME", "ACTOR", "ACTION"}, Rows: rows, Empty: "No activities found."}
}

func displayName(credential auth.Credential) string {
	return firstNonEmpty(credential.User.DisplayName, credential.User.Username, credential.User.ID, "unknown")
}

func scopes(values []string) string {
	if len(values) == 0 {
		return "unknown"
	}
	return strings.Join(values, ", ")
}

func labels(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}

func mrNumber(mr yunxiao.MergeRequest) string {
	return firstNonEmpty(mr.IID, mr.ID)
}

func short(value string) string {
	if len(value) > 7 {
		return value[:7]
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
