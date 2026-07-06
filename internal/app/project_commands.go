package app

import (
	"fmt"

	"github.com/gouzil/yunxiao-cli/internal/output"
	"github.com/gouzil/yunxiao-cli/internal/terminal"
	"github.com/gouzil/yunxiao-cli/internal/yunxiao"
	"github.com/spf13/cobra"
)

func (r *Root) newProjectCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Manage Yunxiao projects"}
	cmd.AddCommand(
		r.newProjectListCommand(),
		r.newProjectViewCommand(),
		r.newProjectMemberCommand(),
		r.newProjectIterationCommand(),
		r.newProjectMilestoneCommand(),
		r.newProjectLabelCommand(),
	)
	return cmd
}

func (r *Root) newProjectListCommand() *cobra.Command {
	var request yunxiao.ListProjectsRequest
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			result, err := r.services.Project.ListProjects(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.ProjectTable(result.Projects), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	return cmd
}

func (r *Root) newProjectViewCommand() *cobra.Command {
	var web bool
	cmd := &cobra.Command{
		Use:   "view [project]",
		Short: "View a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := ""
			if len(args) > 0 {
				projectID = args[0]
			} else {
				var err error
				projectID, err = r.defaultProject()
				if err != nil {
					return err
				}
			}
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			result, err := r.services.Project.GetProject(cmd.Context(), yunxiao.GetProjectRequest{Organization: resolved.Organization.Value, ProjectID: projectID})
			if err != nil {
				return err
			}
			if web {
				return openURL(r, result.Project.WebURL)
			}
			project := result.Project
			detail := output.Detail{Title: fmt.Sprintf("Project %s: %s", project.ID, project.Name), Fields: []output.DetailField{
				{Name: "Status", Value: yunxiao.Unknown(project.Status)},
				{Name: "Owner", Value: yunxiao.Unknown(project.Owner)},
				{Name: "Members", Value: fmt.Sprint(project.MemberCount)},
				{Name: "Iterations", Value: fmt.Sprint(project.IterationCount)},
				{Name: "Web URL", Value: yunxiao.Unknown(project.WebURL)},
			}}
			return r.renderer.RenderDetail(result, detail, r.renderOptions())
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "Open project in a browser")
	return cmd
}

func (r *Root) newProjectMemberCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "member", Short: "Manage project members"}
	cmd.AddCommand(r.newProjectMemberListCommand())
	return cmd
}

func (r *Root) newProjectMemberListCommand() *cobra.Command {
	var request yunxiao.ListProjectMembersRequest
	cmd := &cobra.Command{
		Use:   "list [project]",
		Short: "List project members",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := projectArgOrDefault(r, args)
			if err != nil {
				return err
			}
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.ProjectID = projectID
			result, err := r.services.Project.ListProjectMembers(cmd.Context(), request)
			if err != nil {
				return err
			}
			rows := make([]output.Row, 0, len(result.Members))
			for _, member := range result.Members {
				rows = append(rows, output.Row{member.UserID, member.Name, member.Role, member.Joined})
			}
			return r.renderer.Render(result, output.Table{Headers: []string{"USER ID", "NAME", "ROLE", "JOINED"}, Rows: rows, Empty: "No members found."}, r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	return cmd
}

func (r *Root) newProjectIterationCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "iteration", Short: "Manage project iterations"}
	cmd.AddCommand(r.newProjectIterationListCommand(), r.newProjectIterationViewCommand())
	return cmd
}

func (r *Root) newProjectIterationListCommand() *cobra.Command {
	var request yunxiao.ListIterationsRequest
	cmd := &cobra.Command{
		Use:   "list [project]",
		Short: "List project iterations",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := projectArgOrDefault(r, args)
			if err != nil {
				return err
			}
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.ProjectID = projectID
			result, err := r.services.Project.ListIterations(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.IterationTable(result.Iterations), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	return cmd
}

func (r *Root) newProjectIterationViewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "view <project> <iteration>",
		Short: "View an iteration",
		Args:  requireArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			result, err := r.services.Project.GetIteration(cmd.Context(), yunxiao.GetIterationRequest{Organization: resolved.Organization.Value, ProjectID: args[0], IterationID: args[1]})
			if err != nil {
				return err
			}
			iteration := result.Iteration
			detail := output.Detail{Title: fmt.Sprintf("Iteration %s: %s", iteration.ID, iteration.Name), Fields: []output.DetailField{
				{Name: "Status", Value: yunxiao.Unknown(iteration.Status)},
				{Name: "Start", Value: yunxiao.Unknown(iteration.StartDate)},
				{Name: "End", Value: yunxiao.Unknown(iteration.EndDate)},
				{Name: "Project", Value: yunxiao.Unknown(iteration.ProjectID)},
			}}
			return r.renderer.RenderDetail(result, detail, r.renderOptions())
		},
	}
}

func (r *Root) newProjectMilestoneCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "milestone", Short: "Manage project milestones"}
	cmd.AddCommand(r.newProjectMilestoneListCommand(), r.newProjectMilestoneViewCommand())
	return cmd
}

func (r *Root) newProjectMilestoneListCommand() *cobra.Command {
	var request yunxiao.ListMilestonesRequest
	cmd := &cobra.Command{
		Use:   "list [project]",
		Short: "List project milestones",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := projectArgOrDefault(r, args)
			if err != nil {
				return err
			}
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.ProjectID = projectID
			result, err := r.services.Project.ListMilestones(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.MilestoneTable(result.Milestones), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	return cmd
}

func (r *Root) newProjectMilestoneViewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "view <project> <milestone>",
		Short: "View a project milestone",
		Args:  requireArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			result, err := r.services.Project.GetMilestone(cmd.Context(), yunxiao.GetMilestoneRequest{Organization: resolved.Organization.Value, ProjectID: args[0], MilestoneID: args[1]})
			if err != nil {
				return err
			}
			milestone := result.Milestone
			detail := output.Detail{Title: fmt.Sprintf("Milestone %s: %s", milestone.ID, milestone.Name), Fields: []output.DetailField{
				{Name: "Status", Value: yunxiao.Unknown(milestone.Status)},
				{Name: "Due", Value: yunxiao.Unknown(milestone.DueDate)},
				{Name: "Project", Value: args[0]},
			}}
			return r.renderer.RenderDetail(result, detail, r.renderOptions())
		},
	}
}

func (r *Root) newProjectLabelCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "label", Short: "Manage project labels"}
	cmd.AddCommand(r.newProjectLabelListCommand())
	return cmd
}

func (r *Root) newProjectLabelListCommand() *cobra.Command {
	var request yunxiao.ListLabelsRequest
	cmd := &cobra.Command{
		Use:   "list [project]",
		Short: "List project labels",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := projectArgOrDefault(r, args)
			if err != nil {
				return err
			}
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.ProjectID = projectID
			result, err := r.services.Project.ListLabels(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.LabelTable(result.Labels), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	return cmd
}

func (r *Root) newWorkItemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workitem",
		Short: "Manage Yunxiao work items for requirements, defects and tasks",
	}
	cmd.AddCommand(
		r.newWorkItemListCommand(),
		r.newWorkItemViewCommand(),
		r.newWorkItemCreateCommand(),
		r.newWorkItemEditCommand(),
		r.newWorkItemDeleteCommand(),
		r.newWorkItemActivityCommand(),
	)
	return cmd
}

func (r *Root) newWorkItemListCommand() *cobra.Command {
	var request yunxiao.ListWorkItemsRequest
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work items",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			if request.ProjectID == "" {
				request.ProjectID = resolved.Project.Value
			}
			result, err := r.services.WorkItem.ListWorkItems(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.WorkItemTable(result.WorkItems), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	cmd.Flags().StringVar(&request.ProjectID, "project-id", "", "Project ID")
	cmd.Flags().StringVar(&request.State, "state", "", "Work item state")
	cmd.Flags().StringVar(&request.Assignee, "assignee", "", "Assignee")
	return cmd
}

func (r *Root) newWorkItemViewCommand() *cobra.Command {
	var web bool
	cmd := &cobra.Command{
		Use:   "view <workitem>",
		Short: "View a work item",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			result, err := r.services.WorkItem.GetWorkItem(cmd.Context(), yunxiao.GetWorkItemRequest{Organization: resolved.Organization.Value, ID: args[0]})
			if err != nil {
				return err
			}
			if web {
				return openURL(r, result.WorkItem.WebURL)
			}
			description, err := terminal.NewMarkdownRenderer(r.flags.Plain).Render(result.WorkItem.Description)
			if err != nil {
				return err
			}
			return r.renderer.RenderDetail(result, output.WorkItemDetail(result.WorkItem, description), r.renderOptions())
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "Open work item in a browser")
	return cmd
}

func (r *Root) newWorkItemCreateCommand() *cobra.Command {
	var request yunxiao.CreateWorkItemRequest
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a work item",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			if request.ProjectID == "" {
				projectID, err := r.defaultProject()
				if err != nil {
					return err
				}
				request.ProjectID = projectID
			}
			if request.Type == "" || request.Title == "" {
				return fmt.Errorf("--type and --title are required")
			}
			result, err := r.services.WorkItem.CreateWorkItem(cmd.Context(), request)
			if err != nil {
				return err
			}
			item := result.WorkItem
			fmt.Fprintf(r.io.Out, "✓ Created work item %s\n- Title: %s\n- Type: %s\n- Web URL: %s\n", item.ID, item.Title, item.Type, yunxiao.Unknown(item.WebURL))
			return nil
		},
	}
	cmd.Flags().StringVar(&request.ProjectID, "project-id", "", "Project ID")
	cmd.Flags().StringVar(&request.Type, "type", "", "Work item type")
	cmd.Flags().StringVar(&request.Title, "title", "", "Work item title")
	cmd.Flags().StringVarP(&request.Body, "body", "b", "", "Work item description")
	cmd.Flags().StringVar(&request.Assignee, "assignee", "", "Assignee")
	return cmd
}

func (r *Root) newWorkItemEditCommand() *cobra.Command {
	var request yunxiao.UpdateWorkItemRequest
	cmd := &cobra.Command{
		Use:   "edit <workitem>",
		Short: "Edit a work item",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.ID = args[0]
			result, err := r.services.WorkItem.UpdateWorkItem(cmd.Context(), request)
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Updated work item %s\n", result.WorkItem.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&request.Title, "title", "", "Work item title")
	cmd.Flags().StringVarP(&request.Body, "body", "b", "", "Work item description")
	cmd.Flags().StringVar(&request.State, "state", "", "Work item state")
	cmd.Flags().StringVar(&request.Assignee, "assignee", "", "Assignee")
	return cmd
}

func (r *Root) newWorkItemDeleteCommand() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <workitem>",
		Short: "Delete a work item",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			_, err = r.services.WorkItem.DeleteWorkItem(cmd.Context(), yunxiao.DeleteWorkItemRequest{Organization: resolved.Organization.Value, ID: args[0], Yes: yes})
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Deleted work item %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation")
	return cmd
}

func (r *Root) newWorkItemActivityCommand() *cobra.Command {
	var request yunxiao.ListWorkItemActivitiesRequest
	cmd := &cobra.Command{
		Use:   "activity <workitem>",
		Short: "List work item activity",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.ID = args[0]
			result, err := r.services.WorkItem.ListWorkItemActivities(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.ActivityTable(result.Activities), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	return cmd
}

func projectArgOrDefault(root *Root, args []string) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}
	return root.defaultProject()
}
