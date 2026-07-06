package app

import (
	"fmt"

	"github.com/gouzil/yunxiao-cli/internal/output"
	"github.com/gouzil/yunxiao-cli/internal/terminal"
	"github.com/gouzil/yunxiao-cli/internal/yunxiao"
	"github.com/spf13/cobra"
)

func (r *Root) newMRCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mr",
		Short: "Manage Yunxiao merge requests",
	}
	cmd.AddCommand(
		r.newMRListCommand(),
		r.newMRViewCommand(),
		r.newMRCreateCommand(),
		r.newMREditCommand(),
		r.newMRDiffCommand(),
		r.newMRFilesCommand(),
		r.newMRCommentCommand(),
		r.newMRResolveCommentCommand("resolve-comment", true),
		r.newMRResolveCommentCommand("unresolve-comment", false),
		r.newMRReviewCommand("approve", yunxiao.ReviewDecisionApprove),
		r.newMRReviewCommand("changes-requested", yunxiao.ReviewDecisionChangesRequested),
		r.newMRActionCommand("merge"),
		r.newMRActionCommand("close"),
		r.newMRActionCommand("reopen"),
		r.newMRStatusCommand(),
	)
	return cmd
}

func (r *Root) newMRListCommand() *cobra.Command {
	var request yunxiao.ListMergeRequestsRequest
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List merge requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			request.RepositoryID = repo
			request.Organization = organization
			result, err := r.services.MR.ListMergeRequests(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.MergeRequestTable(result.MergeRequests), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	cmd.Flags().StringVar(&request.State, "state", "", "Merge request state")
	return cmd
}

func (r *Root) newMRViewCommand() *cobra.Command {
	var web bool
	cmd := &cobra.Command{
		Use:   "view <mr>",
		Short: "View a merge request",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			result, err := r.services.MR.GetMergeRequest(cmd.Context(), yunxiao.GetMergeRequestRequest{Organization: organization, RepositoryID: repo, MergeRequestID: args[0]})
			if err != nil {
				return err
			}
			if web {
				return openURL(r, result.MergeRequest.WebURL)
			}
			renderedDescription, err := terminal.NewMarkdownRenderer(r.flags.Plain).Render(result.MergeRequest.Description)
			if err != nil {
				return err
			}
			return r.renderer.RenderDetail(result, output.MergeRequestDetail(result.MergeRequest, renderedDescription), r.renderOptions())
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "Open merge request in a browser")
	return cmd
}

func (r *Root) newMRCreateCommand() *cobra.Command {
	var request yunxiao.CreateMergeRequestRequest
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a merge request",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			request.RepositoryID = repo
			request.Organization = organization
			if request.SourceBranch == "" || request.TargetBranch == "" || request.Title == "" {
				return fmt.Errorf("--source, --target and --title are required")
			}
			result, err := r.services.MR.CreateMergeRequest(cmd.Context(), request)
			if err != nil {
				return err
			}
			mr := result.MergeRequest
			fmt.Fprintf(r.io.Out, "✓ Created merge request !%s\n- Title: %s\n- Source: %s\n- Target: %s\n- Web URL: %s\n", mrID(mr), mr.Title, mr.SourceBranch, mr.TargetBranch, yunxiao.Unknown(mr.WebURL))
			return nil
		},
	}
	cmd.Flags().StringVar(&request.SourceBranch, "source", "", "Source branch")
	cmd.Flags().StringVar(&request.TargetBranch, "target", "", "Target branch")
	cmd.Flags().StringVar(&request.Title, "title", "", "Merge request title")
	cmd.Flags().StringVarP(&request.Description, "body", "b", "", "Merge request description")
	cmd.Flags().BoolVar(&request.Draft, "draft", false, "Create as draft")
	return cmd
}

func (r *Root) newMREditCommand() *cobra.Command {
	var request yunxiao.UpdateMergeRequestRequest
	cmd := &cobra.Command{
		Use:   "edit <mr>",
		Short: "Edit a merge request",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			request.RepositoryID = repo
			request.Organization = organization
			request.MergeRequestID = args[0]
			result, err := r.services.MR.UpdateMergeRequest(cmd.Context(), request)
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Updated merge request !%s\n", firstNonEmpty(mrID(result.MergeRequest), args[0]))
			return nil
		},
	}
	cmd.Flags().StringVar(&request.Title, "title", "", "Merge request title")
	cmd.Flags().StringVarP(&request.Description, "body", "b", "", "Merge request description")
	cmd.Flags().StringVar(&request.TargetBranch, "target", "", "Target branch")
	return cmd
}

func (r *Root) newMRDiffCommand() *cobra.Command {
	var version string
	cmd := &cobra.Command{
		Use:   "diff <mr>",
		Short: "Show merge request diff",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			result, err := r.services.MR.GetMergeRequestDiff(cmd.Context(), yunxiao.GetMergeRequestDiffRequest{Organization: organization, RepositoryID: repo, MergeRequestID: args[0], Version: version})
			if err != nil {
				return err
			}
			return r.renderer.RenderText(result, result.Diff, r.renderOptions())
		},
	}
	cmd.Flags().StringVar(&version, "version", "", "Merge request version")
	return cmd
}

func (r *Root) newMRFilesCommand() *cobra.Command {
	var version string
	cmd := &cobra.Command{
		Use:   "files <mr>",
		Short: "List changed files",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			result, err := r.services.MR.ListMergeRequestFiles(cmd.Context(), yunxiao.ListMergeRequestFilesRequest{Organization: organization, RepositoryID: repo, MergeRequestID: args[0], Version: version})
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.MRFileTable(result.Files), r.renderOptions())
		},
	}
	cmd.Flags().StringVar(&version, "version", "", "Merge request version")
	return cmd
}

func (r *Root) newMRCommentCommand() *cobra.Command {
	var request yunxiao.CommentMergeRequestRequest
	var bodyFile string
	cmd := &cobra.Command{
		Use:   "comment <mr>",
		Short: "Comment on a merge request",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			value, err := readFileOrValue(bodyFile, request.Body)
			if err != nil {
				return err
			}
			if value == "" {
				return fmt.Errorf("comment body is required")
			}
			request.Organization = organization
			request.RepositoryID = repo
			request.MergeRequestID = args[0]
			request.Body = value
			result, err := r.services.MR.CommentMergeRequest(cmd.Context(), request)
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Commented on merge request !%s\n- Comment ID: %s\n", args[0], result.ID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&request.Body, "body", "b", "", "Comment body")
	cmd.Flags().StringVarP(&bodyFile, "body-file", "F", "", "Read comment body from file")
	cmd.Flags().BoolVar(&request.Draft, "draft", false, "Create a draft comment")
	cmd.Flags().BoolVar(&request.Resolved, "resolved", false, "Mark comment as resolved")
	cmd.Flags().StringVar(&request.FilePath, "file", "", "Changed file path for an inline comment")
	cmd.Flags().IntVar(&request.LineNumber, "line", 0, "Changed file line number for an inline comment")
	cmd.Flags().StringVar(&request.PatchSetID, "patch-set", "", "Patch set ID for the comment")
	cmd.Flags().StringVar(&request.FromPatchSetID, "from-patch-set", "", "From patch set ID for an inline comment")
	cmd.Flags().StringVar(&request.ToPatchSetID, "to-patch-set", "", "To patch set ID for an inline comment")
	return cmd
}

func (r *Root) newMRResolveCommentCommand(name string, resolved bool) *cobra.Command {
	return &cobra.Command{
		Use:   name + " <mr> <comment>",
		Short: name + " a merge request comment",
		Args:  requireArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			result, err := r.services.MR.ResolveMergeRequestComment(cmd.Context(), yunxiao.ResolveMergeRequestCommentRequest{Organization: organization, RepositoryID: repo, MergeRequestID: args[0], CommentID: args[1], Resolved: resolved})
			if err != nil {
				return err
			}
			if resolved {
				fmt.Fprintf(r.io.Out, "✓ Resolved comment %s on merge request !%s\n", result.ID, args[0])
			} else {
				fmt.Fprintf(r.io.Out, "✓ Marked comment %s as unresolved on merge request !%s\n", result.ID, args[0])
			}
			return nil
		},
	}
}

func (r *Root) newMRReviewCommand(name string, decision yunxiao.ReviewDecision) *cobra.Command {
	var body string
	cmd := &cobra.Command{
		Use:   name + " <mr>",
		Short: name + " a merge request",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			result, err := r.services.MR.ReviewMergeRequest(cmd.Context(), yunxiao.ReviewMergeRequestRequest{Organization: organization, RepositoryID: repo, MergeRequestID: args[0], Decision: decision, Body: body})
			if err != nil {
				return err
			}
			if decision == yunxiao.ReviewDecisionApprove {
				fmt.Fprintf(r.io.Out, "✓ Approved merge request !%s\n", result.MergeRequestID)
			} else {
				fmt.Fprintf(r.io.Out, "✓ Requested changes on merge request !%s\n", result.MergeRequestID)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&body, "body", "b", "", "Review body")
	return cmd
}

func (r *Root) newMRActionCommand(action string) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   action + " <mr>",
		Short: action + " a merge request",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			request := yunxiao.MergeRequestActionRequest{Organization: organization, RepositoryID: repo, MergeRequestID: args[0], Yes: yes}
			var result yunxiao.MergeRequestActionResult
			switch action {
			case "merge":
				result, err = r.services.MR.MergeMergeRequest(cmd.Context(), request)
			case "close":
				result, err = r.services.MR.CloseMergeRequest(cmd.Context(), request)
			case "reopen":
				result, err = r.services.MR.ReopenMergeRequest(cmd.Context(), request)
			}
			if err != nil {
				return err
			}
			switch action {
			case "merge":
				fmt.Fprintf(r.io.Out, "✓ Merged merge request !%s\n- Merge commit: %s\n", mrID(result.MergeRequest), yunxiao.Unknown(result.MergeCommit))
			case "close":
				fmt.Fprintf(r.io.Out, "✓ Closed merge request !%s\n", mrID(result.MergeRequest))
			case "reopen":
				fmt.Fprintf(r.io.Out, "✓ Reopened merge request !%s\n", mrID(result.MergeRequest))
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation")
	return cmd
}

func (r *Root) newMRStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status <mr>",
		Short: "Show merge request status",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			result, err := r.services.MR.GetMergeRequestStatus(cmd.Context(), yunxiao.GetMergeRequestRequest{Organization: organization, RepositoryID: repo, MergeRequestID: args[0]})
			if err != nil {
				return err
			}
			mr := result.MergeRequest
			detail := output.Detail{Title: fmt.Sprintf("Merge request !%s", mrID(mr)), Fields: []output.DetailField{
				{Name: "State", Value: yunxiao.Unknown(mr.State)},
				{Name: "Mergeable", Value: yunxiao.YesNoUnknown(mr.Mergeable)},
				{Name: "Review", Value: yunxiao.Unknown(mr.ReviewStatus)},
				{Name: "Conflicts", Value: yunxiao.YesNoUnknown(mr.HasConflicts)},
				{Name: "Pipeline", Value: yunxiao.Unknown(mr.PipelineStatus)},
			}}
			return r.renderer.RenderDetail(result, detail, r.renderOptions())
		},
	}
}

func mrID(mr yunxiao.MergeRequest) string {
	if mr.IID != "" {
		return mr.IID
	}
	if mr.ID != "" {
		return mr.ID
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return "unknown"
}
