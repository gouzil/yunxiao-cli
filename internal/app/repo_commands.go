package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gouzil/yunxiao-cli/internal/config"
	"github.com/gouzil/yunxiao-cli/internal/output"
	"github.com/gouzil/yunxiao-cli/internal/yunxiao"
	"github.com/spf13/cobra"
)

func (r *Root) newRepoCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "repo", Short: "Manage Yunxiao repositories"}
	cmd.AddCommand(
		r.newRepoListCommand(),
		r.newRepoViewCommand(),
		r.newRepoCreateCommand(),
		r.newRepoUpdateCommand(),
		r.newRepoActionCommand("archive"),
		r.newRepoActionCommand("unarchive"),
		r.newRepoActionCommand("delete"),
		r.newRepoSetDefaultCommand(),
	)
	return cmd
}

func (r *Root) newRepoListCommand() *cobra.Command {
	var options yunxiao.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			result, err := r.services.Repo.ListRepositories(cmd.Context(), yunxiao.ListRepositoriesRequest{Organization: resolved.Organization.Value, Options: options})
			if err != nil {
				return err
			}
			options := r.renderOptions()
			value := any(result)
			if len(options.JSONFields) > 0 && !jsonFieldRequested(options.JSONFields, "repositories") && !jsonFieldRequested(options.JSONFields, "meta") {
				value = result.Repositories
			}
			return r.renderer.Render(value, output.RepositoryTable(result.Repositories), options)
		},
	}
	addListFlags(cmd, &options)
	return cmd
}

func (r *Root) newRepoViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view [repo]",
		Short: "View a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRef, err := argOrDefaultRepo(r, args)
			if err != nil {
				return err
			}
			result, err := r.resolveRepository(cmd.Context(), repoRef)
			if err != nil {
				return err
			}
			return r.renderer.RenderDetail(result, output.RepositoryDetail(result.Repository), r.renderOptions())
		},
	}
	return cmd
}

func (r *Root) resolveRepository(ctx context.Context, repoRef string) (yunxiao.RepositoryResult, error) {
	resolved, err := r.resolvedConfig()
	if err != nil {
		return yunxiao.RepositoryResult{}, err
	}
	result, err := r.services.Repo.ListRepositories(ctx, yunxiao.ListRepositoriesRequest{Organization: resolved.Organization.Value})
	if err != nil {
		return yunxiao.RepositoryResult{}, err
	}
	matches := []yunxiao.Repository{}
	for _, repo := range result.Repositories {
		if repo.ID == repoRef || repo.Path == repoRef || repo.Name == repoRef {
			matches = append(matches, repo)
		}
	}
	if len(matches) == 0 {
		return yunxiao.RepositoryResult{}, fmt.Errorf("repository %q not found", repoRef)
	}
	if len(matches) > 1 {
		return yunxiao.RepositoryResult{}, fmt.Errorf("repository %q is ambiguous; use the repository ID or full path", repoRef)
	}
	repo := matches[0]
	organization, _, hasPathOrganization := strings.Cut(repo.Path, "/")
	if !hasPathOrganization {
		organization = resolved.Organization.Value
	}
	if repo.ID == "" || organization == "" {
		return yunxiao.RepositoryResult{Repository: repo, Meta: result.Meta}, nil
	}
	return r.services.Repo.GetRepository(ctx, yunxiao.GetRepositoryRequest{Organization: organization, RepositoryID: repo.ID})
}

func (r *Root) newRepoCreateCommand() *cobra.Command {
	var request yunxiao.CreateRepositoryRequest
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a repository",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			request.Name = args[0]
			request.Organization = resolved.Organization.Value
			result, err := r.services.Repo.CreateRepository(cmd.Context(), request)
			if err != nil {
				return err
			}
			repo := result.Repository
			fmt.Fprintf(r.io.Out, "✓ Created repository %s\n- ID: %s\n- Default branch: %s\n- Web URL: %s\n", repo.Path, repo.ID, yunxiao.Unknown(repo.DefaultBranch), yunxiao.Unknown(repo.WebURL))
			return nil
		},
	}
	cmd.Flags().StringVar(&request.Namespace, "namespace", "", "Repository namespace")
	cmd.Flags().StringVar(&request.Visibility, "visibility", "", "Repository visibility")
	cmd.Flags().StringVar(&request.DefaultBranch, "default-branch", "master", "Default branch name")
	cmd.Flags().StringVar(&request.Description, "description", "", "Repository description")
	return cmd
}

func (r *Root) newRepoUpdateCommand() *cobra.Command {
	var request yunxiao.UpdateRepositoryRequest
	cmd := &cobra.Command{
		Use:   "update <repo>",
		Short: "Update a repository",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repositoryID, organization, _, err := r.repositoryContext(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			request.Organization = organization
			request.RepositoryID = repositoryID
			result, err := r.services.Repo.UpdateRepository(cmd.Context(), request)
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Updated repository %s\n", firstRepoName(result.Repository, args[0]))
			return nil
		},
	}
	cmd.Flags().StringVar(&request.Name, "name", "", "Repository name")
	cmd.Flags().StringVar(&request.Description, "description", "", "Repository description")
	cmd.Flags().StringVar(&request.Visibility, "visibility", "", "Repository visibility")
	return cmd
}

func (r *Root) newRepoActionCommand(action string) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   action + " <repo>",
		Short: action + " a repository",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repositoryID, organization, _, err := r.repositoryContext(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			request := yunxiao.RepositoryActionRequest{Organization: organization, RepositoryID: repositoryID, Yes: yes}
			var result yunxiao.RepositoryActionResult
			switch action {
			case "archive":
				result, err = r.services.Repo.ArchiveRepository(cmd.Context(), request)
			case "unarchive":
				result, err = r.services.Repo.UnarchiveRepository(cmd.Context(), request)
			case "delete":
				result, err = r.services.Repo.DeleteRepository(cmd.Context(), request)
			}
			if err != nil {
				return err
			}
			if !result.Supported {
				fmt.Fprintln(r.io.Out, "X repo unarchive is not supported by the current Yunxiao API mapping")
				fmt.Fprintln(r.io.Out, "  - Status: pending API confirmation")
				return nil
			}
			fmt.Fprintf(r.io.Out, "✓ %s repository %s\n", pastTense(action), firstRepoName(result.Repository, args[0]))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation")
	return cmd
}

func (r *Root) newRepoSetDefaultCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-default <repo>",
		Short: "Bind the current local repository to a Yunxiao repository",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := r.configStore.Set(config.ScopeRepo, config.KeyRepo, args[0]); err != nil {
				return err
			}
			path := r.configStore.Paths().Repo
			fmt.Fprintf(r.io.Out, "✓ Set default repository to %s\n- Scope: repository\n- Config: %s\n", args[0], path)
			return nil
		},
	}
}

func (r *Root) newBranchCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "branch", Short: "View repository branches"}
	cmd.AddCommand(r.newBranchListCommand(), r.newBranchViewCommand())
	return cmd
}

func (r *Root) newBranchListCommand() *cobra.Command {
	var options yunxiao.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			result, err := r.services.Branch.ListBranches(cmd.Context(), yunxiao.ListBranchesRequest{Organization: organization, RepositoryID: repo, Options: options})
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.BranchTable(result.Branches), r.renderOptions())
		},
	}
	addListFlags(cmd, &options)
	return cmd
}

func (r *Root) newBranchViewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "view <branch>",
		Short: "View a branch",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			result, err := r.services.Branch.GetBranch(cmd.Context(), yunxiao.GetBranchRequest{Organization: organization, RepositoryID: repo, Branch: args[0]})
			if err != nil {
				return err
			}
			branch := result.Branch
			detail := output.Detail{Title: branch.Name, Fields: []output.DetailField{
				{Name: "Repository", Value: repo},
				{Name: "Default", Value: fmt.Sprint(branch.Default)},
				{Name: "Protected", Value: yunxiao.YesNoUnknown(branch.Protected)},
				{Name: "Commit", Value: yunxiao.Unknown(branch.CommitSHA)},
				{Name: "Author", Value: yunxiao.Unknown(branch.Author)},
				{Name: "Updated", Value: yunxiao.FormatTime(branch.UpdatedAt)},
			}}
			return r.renderer.RenderDetail(result, detail, r.renderOptions())
		},
	}
}

func (r *Root) newCommitCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "commit", Short: "View repository commits"}
	cmd.AddCommand(r.newCommitListCommand(), r.newCommitViewCommand())
	return cmd
}

func (r *Root) newCommitListCommand() *cobra.Command {
	var request yunxiao.ListCommitsRequest
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commits",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, defaultBranch, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			request.RepositoryID = repo
			request.Organization = organization
			if request.Branch == "" {
				request.Branch = defaultBranch
			}
			result, err := r.services.Commit.ListCommits(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.CommitTable(result.Commits), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	cmd.Flags().StringVar(&request.Branch, "branch", "", "Branch filter")
	cmd.Flags().StringVar(&request.Path, "path", "", "Path filter")
	cmd.Flags().StringVar(&request.Since, "since", "", "Start commit or time")
	cmd.Flags().StringVar(&request.Until, "until", "", "End commit or time")
	return cmd
}

func (r *Root) newCommitViewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "view <sha>",
		Short: "View a commit",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			result, err := r.services.Commit.GetCommit(cmd.Context(), yunxiao.GetCommitRequest{Organization: organization, RepositoryID: repo, SHA: args[0]})
			if err != nil {
				return err
			}
			commit := result.Commit
			text := fmt.Sprintf("commit %s\nAuthor: %s <%s>\nDate:   %s\n\n    %s\n\n%s\n", commit.SHA, commit.AuthorName, commit.AuthorMail, yunxiao.FormatTime(commit.Date), commit.Title, commit.Message)
			return r.renderer.RenderText(result, text, r.renderOptions())
		},
	}
}

func (r *Root) newFileCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "file", Short: "View repository files"}
	cmd.AddCommand(r.newFileViewCommand(), r.newFileTreeCommand())
	return cmd
}

func (r *Root) newFileViewCommand() *cobra.Command {
	var request yunxiao.GetFileRequest
	cmd := &cobra.Command{
		Use:   "view <path>",
		Short: "View file content",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			request.RepositoryID = repo
			request.Organization = organization
			request.Path = args[0]
			result, err := r.services.File.GetFile(cmd.Context(), request)
			if err != nil {
				return err
			}
			if request.Metadata {
				detail := output.Detail{Title: result.File.Path, Fields: []output.DetailField{
					{Name: "Ref", Value: yunxiao.Unknown(result.File.Ref)},
					{Name: "Blob ID", Value: yunxiao.Unknown(result.File.BlobID)},
					{Name: "Size", Value: fmt.Sprint(result.File.Size)},
					{Name: "Encoding", Value: yunxiao.Unknown(result.File.Encoding)},
				}}
				return r.renderer.RenderDetail(result, detail, r.renderOptions())
			}
			content, err := fileContent(result.File)
			if err != nil {
				return err
			}
			return r.renderer.RenderText(result, content, r.renderOptions())
		},
	}
	cmd.Flags().StringVar(&request.Ref, "ref", "", "Git ref")
	cmd.Flags().BoolVar(&request.Metadata, "metadata", false, "Show file metadata")
	return cmd
}

func (r *Root) newFileTreeCommand() *cobra.Command {
	var request yunxiao.ListFilesRequest
	cmd := &cobra.Command{
		Use:   "tree [path]",
		Short: "View repository tree",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, organization, _, err := r.defaultRepository(cmd.Context())
			if err != nil {
				return err
			}
			request.RepositoryID = repo
			request.Organization = organization
			if len(args) > 0 {
				request.Path = args[0]
			}
			result, err := r.services.File.ListFiles(cmd.Context(), request)
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.FileTreeTable(result.Files), r.renderOptions())
		},
	}
	addListFlags(cmd, &request.Options)
	cmd.Flags().StringVar(&request.Ref, "ref", "", "Git ref")
	return cmd
}

func (r *Root) newSSHKeyCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "ssh-key", Short: "Manage Yunxiao SSH keys"}
	cmd.AddCommand(r.newSSHKeyListCommand(), r.newSSHKeyAddCommand(), r.newSSHKeyDeleteCommand())
	return cmd
}

func (r *Root) newSSHKeyListCommand() *cobra.Command {
	var options yunxiao.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SSH keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			result, err := r.services.SSHKey.ListSSHKeys(cmd.Context(), yunxiao.ListSSHKeysRequest{Organization: resolved.Organization.Value, Options: options})
			if err != nil {
				return err
			}
			return r.renderer.Render(result, output.SSHKeyTable(result.Keys), r.renderOptions())
		},
	}
	addListFlags(cmd, &options)
	return cmd
}

func (r *Root) newSSHKeyAddCommand() *cobra.Command {
	var title string
	cmd := &cobra.Command{
		Use:   "add <public-key-file>",
		Short: "Add an SSH key",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			body, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			if title == "" {
				title = stringsTrimSuffix(filepath.Base(args[0]), ".pub")
			}
			result, err := r.services.SSHKey.CreateSSHKey(cmd.Context(), yunxiao.CreateSSHKeyRequest{Organization: resolved.Organization.Value, Title: title, PublicKey: string(body)})
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Added SSH key %s\n- ID: %s\n- Fingerprint: %s\n", result.Key.Title, result.Key.ID, result.Key.Fingerprint)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "SSH key title")
	return cmd
}

func (r *Root) newSSHKeyDeleteCommand() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <key-id>",
		Short: "Delete an SSH key",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			_, err = r.services.SSHKey.DeleteSSHKey(cmd.Context(), yunxiao.DeleteSSHKeyRequest{Organization: resolved.Organization.Value, ID: args[0], Yes: yes})
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Deleted SSH key %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation")
	return cmd
}

func addListFlags(cmd *cobra.Command, options *yunxiao.ListOptions) {
	cmd.Flags().IntVar(&options.Page, "page", 1, "Page number")
	cmd.Flags().IntVar(&options.PerPage, "per-page", 30, "Items per page")
	cmd.Flags().StringVarP(&options.Query, "query", "q", "", "Search query")
}

func argOrDefaultRepo(root *Root, args []string) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}
	return root.defaultRepo()
}

func (r *Root) defaultRepository(ctx context.Context) (string, string, string, error) {
	repoRef, err := r.defaultRepo()
	if err != nil {
		return "", "", "", err
	}
	return r.repositoryContext(ctx, repoRef)
}

func (r *Root) repositoryContext(ctx context.Context, repoRef string) (string, string, string, error) {
	result, err := r.resolveRepository(ctx, repoRef)
	if err != nil {
		return "", "", "", err
	}
	repo := result.Repository
	repositoryID := repo.ID
	if repositoryID == "" {
		repositoryID = repoRef
	}
	organization, _, ok := strings.Cut(repo.Path, "/")
	if !ok || organization == "" {
		resolved, err := r.resolvedConfig()
		if err != nil {
			return "", "", "", err
		}
		organization = resolved.Organization.Value
	}
	return repositoryID, organization, repo.DefaultBranch, nil
}

func fileContent(file yunxiao.FileEntry) (string, error) {
	if strings.EqualFold(file.Encoding, "base64") {
		decoded, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return "", fmt.Errorf("decode file content: %w", err)
		}
		return string(decoded), nil
	}
	return file.Content, nil
}

func firstRepoName(repo yunxiao.Repository, fallback string) string {
	if repo.Path != "" {
		return repo.Path
	}
	if repo.Name != "" {
		return repo.Name
	}
	if repo.ID != "" {
		return repo.ID
	}
	return fallback
}

func pastTense(action string) string {
	switch action {
	case "archive":
		return "Archived"
	case "delete":
		return "Deleted"
	default:
		return action
	}
}

func stringsTrimSuffix(value string, suffix string) string {
	if len(value) >= len(suffix) && value[len(value)-len(suffix):] == suffix {
		return value[:len(value)-len(suffix)]
	}
	return value
}
