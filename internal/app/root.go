package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gouzil/yunxiao-cli/internal/api"
	"github.com/gouzil/yunxiao-cli/internal/auth"
	"github.com/gouzil/yunxiao-cli/internal/config"
	"github.com/gouzil/yunxiao-cli/internal/extension"
	"github.com/gouzil/yunxiao-cli/internal/output"
	"github.com/gouzil/yunxiao-cli/internal/terminal"
	"github.com/gouzil/yunxiao-cli/internal/yunxiao"
	"github.com/spf13/cobra"
)

type RootOptions struct {
	Build       BuildInfo
	IO          terminal.IOStreams
	ConfigStore *config.Store
	AuthManager *auth.Manager
	Extensions  extension.Manager
	Services    yunxiao.ServiceSet
	Renderer    *output.Renderer
}

type Root struct {
	build       BuildInfo
	io          terminal.IOStreams
	configStore *config.Store
	authManager *auth.Manager
	extensions  extension.Manager
	services    yunxiao.ServiceSet
	renderer    *output.Renderer
	flags       GlobalFlags
	rootCmd     *cobra.Command
}

type GlobalFlags struct {
	Endpoint     string
	Organization string
	Project      string
	Repo         string
	JSONFields   string
	JQ           string
	Template     string
	Plain        bool
	Verbose      bool
}

func NewRoot(options RootOptions) *Root {
	root := &Root{
		build:       options.Build,
		io:          options.IO,
		configStore: options.ConfigStore,
		authManager: options.AuthManager,
		extensions:  options.Extensions,
		services:    options.Services,
		renderer:    options.Renderer,
	}
	root.rootCmd = root.buildRootCommand()
	return root
}

func (r *Root) Execute(ctx context.Context, args []string) error {
	r.rootCmd.SetContext(ctx)
	r.rootCmd.SetArgs(args)
	if err := r.rootCmd.Execute(); err != nil {
		return r.handleError(err)
	}
	return nil
}

func (r *Root) buildRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "yunxiao",
		Short:         "Yunxiao command line interface",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       r.version(),
	}
	cmd.PersistentFlags().StringVar(&r.flags.Endpoint, "endpoint", "", "Yunxiao OpenAPI endpoint")
	cmd.PersistentFlags().StringVar(&r.flags.Organization, "organization", "", "Yunxiao organization identifier")
	cmd.PersistentFlags().StringVar(&r.flags.Project, "project", "", "Yunxiao project identifier")
	cmd.PersistentFlags().StringVarP(&r.flags.Repo, "repo", "R", "", "Yunxiao repository identifier")
	cmd.PersistentFlags().StringVar(&r.flags.JSONFields, "json", "", "Output JSON with comma-separated fields")
	cmd.PersistentFlags().StringVar(&r.flags.JQ, "jq", "", "Filter JSON output using a jq expression")
	cmd.PersistentFlags().StringVar(&r.flags.Template, "template", "", "Render output using a Go template")
	cmd.PersistentFlags().BoolVar(&r.flags.Plain, "plain", false, "Disable rich terminal rendering")
	cmd.PersistentFlags().BoolVar(&r.flags.Verbose, "verbose", false, "Print diagnostic request information")
	cmd.AddCommand(
		r.newAuthCommand(),
		r.newConfigCommand(),
		r.newAPICommand(),
		r.newRepoCommand(),
		r.newBranchCommand(),
		r.newCommitCommand(),
		r.newFileCommand(),
		r.newSSHKeyCommand(),
		r.newMRCommand(),
		r.newPipelineCommand(),
		r.newRunCommand(),
		r.newProjectCommand(),
		r.newWorkItemCommand(),
		r.newSearchCommand(),
		r.newExtensionCommand(),
		r.newAliasCommand(),
		r.newCompletionCommand(),
	)
	r.registerExtensionCommands(cmd)
	return cmd
}

func (r *Root) version() string {
	parts := []string{r.build.Version}
	if r.build.Commit != "" && r.build.Commit != "none" {
		parts = append(parts, r.build.Commit)
	}
	if r.build.Date != "" && r.build.Date != "unknown" {
		parts = append(parts, r.build.Date)
	}
	return strings.Join(parts, " ")
}

func (r *Root) renderOptions() output.Options {
	return output.Options{
		JSONFields: output.ParseJSONFields(r.flags.JSONFields),
		JQ:         r.flags.JQ,
		Template:   r.flags.Template,
		Plain:      r.flags.Plain,
	}
}

func jsonFieldRequested(fields []string, name string) bool {
	for _, field := range fields {
		if field == name {
			return true
		}
	}
	return false
}

func (r *Root) resolvedConfig() (config.Resolved, error) {
	return r.configStore.Resolve(config.Values{
		Endpoint:     r.flags.Endpoint,
		Organization: r.flags.Organization,
		Project:      r.flags.Project,
		Repo:         r.flags.Repo,
	})
}

func (r *Root) defaultRepo() (string, error) {
	resolved, err := r.resolvedConfig()
	if err != nil {
		return "", err
	}
	if resolved.Repo.Value == "" {
		return "", MissingContextError{
			Name: "repository",
			Fix:  "Run: yunxiao repo set-default <repo> or pass --repo <repo>.",
		}
	}
	return resolved.Repo.Value, nil
}

func (r *Root) defaultProject() (string, error) {
	resolved, err := r.resolvedConfig()
	if err != nil {
		return "", err
	}
	if resolved.Project.Value == "" {
		return "", MissingContextError{
			Name: "project",
			Fix:  "Run: yunxiao config set project <project> --scope repo or pass --project <project>.",
		}
	}
	return resolved.Project.Value, nil
}

func (r *Root) handleError(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *api.Error
	if errors.As(err, &apiErr) {
		(&apiErrorView{
			Summary:    apiErr.Summary,
			RequestID:  apiErr.RequestID,
			HTTPStatus: apiErr.HTTPStatus,
			Code:       apiErr.Code,
			Message:    apiErr.Message,
		}).Write(r.io.ErrOut)
		return ExitError{Code: 1, Err: err}
	}
	var extensionExit *extension.ExitError
	if errors.As(err, &extensionExit) {
		return ExitError{Code: extensionExit.Code, Err: err}
	}
	var missing MissingContextError
	if errors.As(err, &missing) {
		fmt.Fprintf(r.io.ErrOut, "X missing %s context\n  - %s\n", missing.Name, missing.Fix)
		return ExitError{Code: 1, Err: err}
	}
	fmt.Fprintf(r.io.ErrOut, "X %v\n", err)
	return ExitError{Code: 1, Err: err}
}

type MissingContextError struct {
	Name string
	Fix  string
}

func (e MissingContextError) Error() string {
	return "missing " + e.Name + " context"
}

type apiErrorView struct {
	Summary    string
	RequestID  string
	HTTPStatus int
	Code       string
	Message    string
}

func (e *apiErrorView) Error() string {
	return e.Message
}

func (e *apiErrorView) Write(out io.Writer) {
	fmt.Fprintf(out, "X %s\n", e.Summary)
	if e.RequestID != "" {
		fmt.Fprintf(out, "  - Request ID: %s\n", e.RequestID)
	}
	if e.HTTPStatus != 0 {
		fmt.Fprintf(out, "  - HTTP status: %d\n", e.HTTPStatus)
	}
	if e.Code != "" {
		fmt.Fprintf(out, "  - Code: %s\n", e.Code)
	}
	if e.Message != "" {
		fmt.Fprintf(out, "  - Message: %s\n", e.Message)
	}
}

func requireArgs(count int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < count {
			return fmt.Errorf("requires at least %d arg(s)", count)
		}
		return nil
	}
}
