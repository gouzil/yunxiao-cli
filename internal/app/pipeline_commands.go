package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/gouzi/yunxiao-cli/internal/output"
	"github.com/gouzi/yunxiao-cli/internal/yunxiao"
	"github.com/spf13/cobra"
)

func (r *Root) newPipelineCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "pipeline", Short: "Manage Yunxiao pipelines"}
	cmd.AddCommand(r.newPipelineListCommand(), r.newPipelineViewCommand(), r.newPipelineRunCommand())
	return cmd
}

func (r *Root) newPipelineListCommand() *cobra.Command {
	var request yunxiao.ListPipelinesRequest
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipelines",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.ProjectID = resolved.Project.Value
			result, err := r.services.Pipeline.ListPipelines(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.PipelineTable(result.Pipelines), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	return cmd
}

func (r *Root) newPipelineViewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "view <pipeline-id>",
		Short: "View a pipeline",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			result, err := r.services.Pipeline.GetPipeline(cmd.Context(), yunxiao.GetPipelineRequest{Organization: resolved.Organization.Value, PipelineID: args[0]})
			if err != nil {
				return err
			}
			pipeline := result.Pipeline
			detail := output.Detail{Title: fmt.Sprintf("Pipeline %s: %s", pipeline.ID, pipeline.Name), Fields: []output.DetailField{
				{Name: "Status", Value: yunxiao.Unknown(pipeline.Status)},
				{Name: "Creator", Value: yunxiao.Unknown(pipeline.Creator)},
				{Name: "Updated", Value: yunxiao.FormatTime(pipeline.UpdatedAt)},
				{Name: "Web URL", Value: yunxiao.Unknown(pipeline.WebURL)},
			}}
			return r.renderer.RenderDetail(result, detail, r.renderOptions())
		},
	}
}

func (r *Root) newPipelineRunCommand() *cobra.Command {
	var request yunxiao.RunPipelineRequest
	var variables []string
	cmd := &cobra.Command{
		Use:   "run <pipeline-id>",
		Short: "Start a pipeline run",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.PipelineID = args[0]
			parsedVariables, err := parsePipelineVariables(variables)
			if err != nil {
				return err
			}
			request.Variables = parsedVariables
			result, err := r.services.Pipeline.RunPipeline(cmd.Context(), request)
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Started pipeline %s\n- Run ID: %s\n- Branch: %s\n- Web URL: %s\n", request.PipelineID, result.Run.ID, yunxiao.Unknown(result.Run.Branch), yunxiao.Unknown(result.Run.WebURL))
			return nil
		},
	}
	cmd.Flags().StringVar(&request.Branch, "branch", "", "Branch to run")
	cmd.Flags().StringArrayVarP(&variables, "var", "F", nil, "Pipeline variable in name=value form")
	return cmd
}

func (r *Root) newRunCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "run", Short: "Manage Yunxiao pipeline runs"}
	cmd.AddCommand(
		r.newRunListCommand(),
		r.newRunViewCommand(),
		r.newRunLogCommand(),
		r.newRunWatchCommand(),
		r.newRunActionCommand("cancel"),
		r.newRunActionCommand("retry"),
		r.newRunTaskActionCommand("retry-task"),
		r.newRunTaskActionCommand("stop-task"),
		r.newRunTaskActionCommand("skip-task"),
	)
	return cmd
}

func (r *Root) newRunListCommand() *cobra.Command {
	var request yunxiao.ListRunsRequest
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipeline runs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if request.PipelineID == "" {
				return fmt.Errorf("--pipeline is required")
			}
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			result, err := r.services.Run.ListRuns(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.RunTable(result.Runs), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	cmd.Flags().StringVar(&request.PipelineID, "pipeline", "", "Pipeline ID")
	cmd.Flags().StringVar(&request.Branch, "branch", "", "Branch")
	cmd.Flags().StringVar(&request.Status, "status", "", "Run status")
	return cmd
}

func (r *Root) newRunViewCommand() *cobra.Command {
	var request yunxiao.GetRunRequest
	cmd := &cobra.Command{
		Use:   "view <run-id>",
		Short: "View a run",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(request.PipelineID) == "" {
				return fmt.Errorf("--pipeline is required")
			}
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.RunID = args[0]
			result, err := r.services.Run.GetRun(cmd.Context(), request)
			if err != nil {
				return err
			}
			run := result.Run
			detail := output.Detail{Title: fmt.Sprintf("Run %s", run.ID), Fields: []output.DetailField{
				{Name: "Pipeline", Value: fmt.Sprintf("%s (%s)", yunxiao.Unknown(run.Pipeline), yunxiao.Unknown(run.PipelineID))},
				{Name: "Status", Value: yunxiao.Unknown(run.Status)},
				{Name: "Trigger", Value: yunxiao.Unknown(run.TriggerMode)},
				{Name: "Triggered by", Value: yunxiao.Unknown(run.TriggeredBy)},
				{Name: "Branch", Value: yunxiao.Unknown(run.Branch)},
				{Name: "Started", Value: yunxiao.FormatTime(run.StartedAt)},
				{Name: "Finished", Value: yunxiao.FormatTime(run.FinishedAt)},
				{Name: "Duration", Value: yunxiao.Unknown(run.Duration)},
				{Name: "Web URL", Value: yunxiao.Unknown(run.WebURL)},
			}}
			return r.renderer.RenderDetail(result, detail, r.renderOptions())
		},
	}
	cmd.Flags().StringVar(&request.PipelineID, "pipeline", "", "Pipeline ID")
	return cmd
}

func (r *Root) newRunLogCommand() *cobra.Command {
	var request yunxiao.GetRunLogRequest
	cmd := &cobra.Command{
		Use:   "log <run-id>",
		Short: "Show run log",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(request.PipelineID) == "" {
				return fmt.Errorf("--pipeline is required")
			}
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.RunID = args[0]
			result, err := r.services.Run.GetRunLog(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.RenderText(result, result.Lines, r.renderOptions())
		},
	}
	cmd.Flags().StringVar(&request.PipelineID, "pipeline", "", "Pipeline ID")
	cmd.Flags().StringVar(&request.JobID, "job", "", "Job ID")
	cmd.Flags().BoolVar(&request.FailedOnly, "failed", false, "Show failed job logs")
	return cmd
}

func (r *Root) newRunWatchCommand() *cobra.Command {
	var request yunxiao.WatchRunRequest
	var interval time.Duration
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "watch <run-id>",
		Short: "Watch a run until terminal status",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(request.PipelineID) == "" {
				return fmt.Errorf("--pipeline is required")
			}
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			request.RunID = args[0]
			request.Interval = interval
			request.Timeout = timeout
			fmt.Fprintf(r.io.Out, "Refreshing run status every %.0fs. Press Ctrl+C to quit.\n\n", interval.Seconds())
			result, err := r.services.Run.WatchRun(cmd.Context(), request)
			if err != nil {
				return err
			}
			if strings.EqualFold(result.Run.Status, "success") {
				fmt.Fprintf(r.io.Out, "✓ Run %s completed with status %s\n", result.Run.ID, result.Run.Status)
			} else {
				fmt.Fprintf(r.io.Out, "X Run %s completed with status %s\n", result.Run.ID, result.Run.Status)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&request.PipelineID, "pipeline", "", "Pipeline ID")
	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "Polling interval")
	cmd.Flags().DurationVar(&timeout, "timeout", 0, "Watch timeout")
	return cmd
}

func (r *Root) newRunActionCommand(action string) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   action + " <run-id>",
		Short: action + " a run",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			request := yunxiao.RunActionRequest{RunID: args[0], Yes: yes}
			var result yunxiao.RunActionResult
			var err error
			if action == "cancel" {
				result, err = r.services.Run.CancelRun(cmd.Context(), request)
			} else {
				result, err = r.services.Run.RetryRun(cmd.Context(), request)
			}
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ %s run %s\n", pastRunAction(action), result.RunID)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation")
	return cmd
}

func (r *Root) newRunTaskActionCommand(action string) *cobra.Command {
	var request yunxiao.RunTaskActionRequest
	cmd := &cobra.Command{
		Use:   action + " <run-id>",
		Short: action + " a run job",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			request.RunID = args[0]
			if request.JobID == "" {
				return fmt.Errorf("--job is required")
			}
			var result yunxiao.RunActionResult
			var err error
			switch action {
			case "retry-task":
				result, err = r.services.Run.RetryTask(cmd.Context(), request)
			case "stop-task":
				result, err = r.services.Run.StopTask(cmd.Context(), request)
			case "skip-task":
				result, err = r.services.Run.SkipTask(cmd.Context(), request)
			}
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ %s job %s in run %s\n", pastRunAction(action), result.JobID, result.RunID)
			return nil
		},
	}
	cmd.Flags().StringVar(&request.JobID, "job", "", "Job ID")
	cmd.Flags().BoolVarP(&request.Yes, "yes", "y", false, "Skip confirmation")
	return cmd
}

func pastRunAction(action string) string {
	switch action {
	case "cancel":
		return "Canceled"
	case "retry":
		return "Retried"
	case "retry-task":
		return "Retried"
	case "stop-task":
		return "Stopped"
	case "skip-task":
		return "Skipped"
	default:
		return action
	}
}

func parsePipelineVariables(values []string) ([]yunxiao.PipelineVariable, error) {
	variables := make([]yunxiao.PipelineVariable, 0, len(values))
	for _, value := range values {
		name, variableValue, ok := strings.Cut(value, "=")
		if !ok || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("pipeline variable must use name=value form")
		}
		variables = append(variables, yunxiao.PipelineVariable{Name: strings.TrimSpace(name), Value: variableValue})
	}
	return variables, nil
}
