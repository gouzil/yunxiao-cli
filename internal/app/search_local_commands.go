package app

import (
	"fmt"
	"strings"

	"github.com/gouzi/yunxiao-cli/internal/browser"
	"github.com/gouzi/yunxiao-cli/internal/config"
	"github.com/gouzi/yunxiao-cli/internal/output"
	"github.com/gouzi/yunxiao-cli/internal/yunxiao"
	"github.com/spf13/cobra"
)

func (r *Root) newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "search", Short: "Search Yunxiao resources"}
	cmd.AddCommand(
		r.newSearchRepoCommand(),
		r.newSearchCodeCommand(),
		r.newSearchCommitCommand(),
		r.newSearchMRCommand(),
		r.newSearchWorkItemCommand(),
	)
	return cmd
}

func (r *Root) newSearchRepoCommand() *cobra.Command {
	var request yunxiao.SearchRepositoriesRequest
	cmd := &cobra.Command{
		Use:   "repo <query>",
		Short: "Search repositories",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Query = args[0]
			request.Organization = resolved.Organization.Value
			result, err := r.services.Search.SearchRepositories(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.RepositoryTable(result.Repositories), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	return cmd
}

func (r *Root) newSearchCodeCommand() *cobra.Command {
	var request yunxiao.SearchCodeRequest
	cmd := &cobra.Command{
		Use:   "code <query>",
		Short: "Search code",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRef, err := repoFlagOrDefault(r, request.RepositoryID)
			if err != nil {
				return err
			}
			repo, organization, defaultBranch, err := r.repositoryContext(cmd.Context(), repoRef)
			if err != nil {
				return err
			}
			request.RepositoryID = repo
			request.Organization = organization
			if request.Ref == "" {
				request.Ref = defaultBranch
			}
			request.Query = args[0]
			result, err := r.services.Search.SearchCode(cmd.Context(), request)
			if err != nil {
				return err
			}
			rows := make([]output.Row, 0, len(result.Matches))
			for _, match := range result.Matches {
				rows = append(rows, output.Row{match.Repository, match.Ref, match.Path, fmt.Sprint(match.Line), match.Match})
			}
			return r.renderer.Render(result, output.Table{Headers: []string{"REPOSITORY", "REF", "PATH", "LINE", "MATCH"}, Rows: rows, Empty: "No code results found."}, r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	cmd.Flags().StringVar(&request.RepositoryID, "repo", "", "Repository ID")
	cmd.Flags().StringVar(&request.Ref, "ref", "", "Git ref")
	return cmd
}

func (r *Root) newSearchCommitCommand() *cobra.Command {
	var request yunxiao.SearchCommitsRequest
	cmd := &cobra.Command{
		Use:   "commit <query>",
		Short: "Search commits",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRef, err := repoFlagOrDefault(r, request.RepositoryID)
			if err != nil {
				return err
			}
			repo, organization, defaultBranch, err := r.repositoryContext(cmd.Context(), repoRef)
			if err != nil {
				return err
			}
			request.RepositoryID = repo
			request.Organization = organization
			if request.Ref == "" {
				request.Ref = defaultBranch
			}
			request.Query = args[0]
			result, err := r.services.Search.SearchCommits(cmd.Context(), request)
			if err != nil {
				return err
			}
			rows := make([]output.Row, 0, len(result.Matches))
			for _, match := range result.Matches {
				rows = append(rows, output.Row{match.Repository, match.Commit.SHA, match.Commit.AuthorName, yunxiao.FormatTime(match.Commit.Date), match.Commit.Title})
			}
			return r.renderer.Render(result, output.Table{Headers: []string{"REPOSITORY", "SHA", "AUTHOR", "DATE", "TITLE"}, Rows: rows, Empty: "No commit results found."}, r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	cmd.Flags().StringVar(&request.RepositoryID, "repo", "", "Repository ID")
	cmd.Flags().StringVar(&request.Ref, "ref", "", "Git ref")
	return cmd
}

func (r *Root) newSearchMRCommand() *cobra.Command {
	var request yunxiao.SearchMergeRequestsRequest
	cmd := &cobra.Command{
		Use:   "mr <query>",
		Short: "Search merge requests",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRef, err := repoFlagOrDefault(r, request.RepositoryID)
			if err != nil {
				return err
			}
			repo, organization, _, err := r.repositoryContext(cmd.Context(), repoRef)
			if err != nil {
				return err
			}
			request.RepositoryID = repo
			request.Organization = organization
			request.Query = args[0]
			result, err := r.services.Search.SearchMergeRequests(cmd.Context(), request)
			if err != nil {
				return err
			}
			rows := make([]output.Row, 0, len(result.Matches))
			for _, match := range result.Matches {
				mr := match.MergeRequest
				rows = append(rows, output.Row{match.Repository, mrID(mr), mr.State, mr.Title, yunxiao.FormatTime(mr.UpdatedAt)})
			}
			return r.renderer.Render(result, output.Table{Headers: []string{"REPOSITORY", "ID", "STATE", "TITLE", "UPDATED"}, Rows: rows, Empty: "No merge request results found."}, r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	cmd.Flags().StringVar(&request.RepositoryID, "repo", "", "Repository ID")
	cmd.Flags().StringVar(&request.State, "state", "", "Merge request state")
	return cmd
}

func (r *Root) newSearchWorkItemCommand() *cobra.Command {
	var request yunxiao.SearchWorkItemsRequest
	cmd := &cobra.Command{
		Use:   "workitem <query>",
		Short: "Search work items",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			request.Query = args[0]
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Organization = resolved.Organization.Value
			if request.ProjectID == "" {
				request.ProjectID = resolved.Project.Value
			}
			result, err := r.services.Search.SearchWorkItems(cmd.Context(), request)
			if err != nil {
				return err
			}
			rows := make([]output.Row, 0, len(result.Matches))
			for _, match := range result.Matches {
				item := match.WorkItem
				rows = append(rows, output.Row{item.ID, item.Type, item.State, item.Title, item.ProjectName})
			}
			return r.renderer.Render(result, output.Table{Headers: []string{"ID", "TYPE", "STATE", "TITLE", "PROJECT"}, Rows: rows, Empty: "No work item results found."}, r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	cmd.Flags().StringVar(&request.ProjectID, "project-id", "", "Project ID")
	cmd.Flags().StringVar(&request.State, "state", "", "Work item state")
	return cmd
}

func (r *Root) newAliasCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "alias", Short: "Manage command aliases"}
	cmd.AddCommand(r.newAliasListCommand(), r.newAliasSetCommand(), r.newAliasDeleteCommand())
	return cmd
}

func (r *Root) newAliasListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := r.configStore.Load(config.ScopeGlobal)
			if err != nil {
				return err
			}
			rows := make([]output.Row, 0, len(file.Aliases))
			for _, alias := range file.Aliases {
				rows = append(rows, output.Row{alias.Name, alias.Expansion})
			}
			return r.renderer.Render(file.Aliases, output.Table{Headers: []string{"ALIAS", "EXPANSION"}, Rows: rows, Empty: "No aliases found."}, r.renderOptions())
		},
	}
}

func (r *Root) newAliasSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <alias> <expansion>",
		Short: "Set an alias",
		Args:  requireArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := r.configStore.Load(config.ScopeGlobal)
			if err != nil {
				return err
			}
			expansion := strings.Join(args[1:], " ")
			found := false
			for index := range file.Aliases {
				if file.Aliases[index].Name == args[0] {
					file.Aliases[index].Expansion = expansion
					found = true
				}
			}
			if !found {
				file.Aliases = append(file.Aliases, config.Alias{Name: args[0], Expansion: expansion})
			}
			if err := r.configStore.Save(config.ScopeGlobal, file); err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Added alias %s: %s\n", args[0], expansion)
			return nil
		},
	}
}

func (r *Root) newAliasDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <alias>",
		Short: "Delete an alias",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := r.configStore.Load(config.ScopeGlobal)
			if err != nil {
				return err
			}
			next := make([]config.Alias, 0, len(file.Aliases))
			for _, alias := range file.Aliases {
				if alias.Name != args[0] {
					next = append(next, alias)
				}
			}
			file.Aliases = next
			if err := r.configStore.Save(config.ScopeGlobal, file); err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Deleted alias %s\n", args[0])
			return nil
		},
	}
}

func (r *Root) newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return r.rootCmd.GenBashCompletion(r.io.Out)
			case "zsh":
				return r.rootCmd.GenZshCompletion(r.io.Out)
			case "fish":
				return r.rootCmd.GenFishCompletion(r.io.Out, true)
			case "powershell":
				return r.rootCmd.GenPowerShellCompletion(r.io.Out)
			default:
				return fmt.Errorf("unsupported shell %q", args[0])
			}
		},
	}
}

func repoFlagOrDefault(root *Root, repo string) (string, error) {
	if repo != "" {
		return repo, nil
	}
	return root.defaultRepo()
}

func openURL(root *Root, rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("resource does not have a Web URL")
	}
	return browser.OpenOrPrint(browser.SystemOpener{}, rawURL, browser.CurrentEnvironment(), root.io.Out)
}
