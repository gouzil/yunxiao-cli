package app

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/gouzil/yunxiao-cli/internal/config"
	"github.com/gouzil/yunxiao-cli/internal/extension"
	"github.com/gouzil/yunxiao-cli/internal/output"
	"github.com/gouzil/yunxiao-cli/internal/terminal"
	"github.com/spf13/cobra"
)

func (r *Root) newExtensionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension",
		Aliases: []string{"extensions", "ext"},
		Short:   "Manage Yunxiao CLI extensions",
	}
	cmd.AddCommand(
		r.newExtensionInstallCommand(),
		r.newExtensionListCommand(),
		r.newExtensionExecCommand(),
		r.newExtensionUpgradeCommand(),
		r.newExtensionRemoveCommand(),
		r.newExtensionCreateCommand(),
	)
	return cmd
}

func (r *Root) newExtensionInstallCommand() *cobra.Command {
	var ref string
	var yes bool
	cmd := &cobra.Command{
		Use:   "install <source>",
		Short: "Install a Yunxiao CLI extension from a Git URL or local directory",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if r.extensions == nil {
				return fmt.Errorf("extension manager is not configured")
			}
			if err := r.confirmExtensionInstall(yes); err != nil {
				return err
			}
			ext, err := r.extensions.Install(cmd.Context(), extension.InstallSource{Source: args[0], Ref: ref, Yes: yes})
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Installed extension %s\n", ext.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&ref, "ref", "", "Git tag, branch, or commit to install")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Confirm extension source trust non-interactively")
	return cmd
}

func (r *Root) newExtensionListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed extensions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if r.extensions == nil {
				return r.renderer.Render(nil, output.Table{Empty: "No extensions installed."}, r.renderOptions())
			}
			extensions, err := r.extensions.List()
			if err != nil {
				return err
			}
			r.annotateExtensionConflicts(extensions)
			rows := make([]output.Row, 0, len(extensions))
			for _, ext := range extensions {
				rows = append(rows, output.Row{
					ext.Name,
					string(ext.Kind),
					extensionSource(ext),
					extensionVersion(ext),
					fmt.Sprint(ext.Pinned),
					fmt.Sprint(ext.Kind == extension.KindLocal),
					ext.Conflict,
				})
			}
			return r.renderer.Render(extensions, output.Table{
				Headers: []string{"NAME", "KIND", "SOURCE", "VERSION", "PINNED", "LOCAL", "CONFLICT"},
				Rows:    rows,
				Empty:   "No extensions installed.",
			}, r.renderOptions())
		},
	}
}

func (r *Root) newExtensionExecCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "exec <name> [args]",
		Short:              "Run an installed extension explicitly",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("requires extension name")
			}
			return r.dispatchExtension(cmd.Context(), args[0], args[1:])
		},
	}
}

func (r *Root) newExtensionUpgradeCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "upgrade [name]",
		Short: "Upgrade installed Git-managed extensions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if r.extensions == nil {
				return fmt.Errorf("extension manager is not configured")
			}
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			if err := r.extensions.Upgrade(cmd.Context(), name, force); err != nil {
				return err
			}
			fmt.Fprintln(r.io.Out, "✓ Successfully checked extension upgrades")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Force upgrade pinned extensions to remote HEAD")
	return cmd
}

func (r *Root) newExtensionRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove an installed extension",
		Args:    requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if r.extensions == nil {
				return fmt.Errorf("extension manager is not configured")
			}
			if err := r.extensions.Remove(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Removed extension %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func (r *Root) newExtensionCreateCommand() *cobra.Command {
	var templateText string
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new extension scaffold",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if r.extensions == nil {
				return fmt.Errorf("extension manager is not configured")
			}
			template := extension.TemplateType(templateText)
			if err := r.extensions.Create(cmd.Context(), args[0], template); err != nil {
				return err
			}
			_, full, _ := extension.NormalizeName(args[0])
			fmt.Fprintf(r.io.Out, "✓ Created extension %s\n", full)
			return nil
		},
	}
	cmd.Flags().StringVar(&templateText, "template", string(extension.TemplateScript), "Extension template: script or go")
	return cmd
}

func (r *Root) registerExtensionCommands(root *cobra.Command) {
	if r.extensions == nil {
		return
	}
	extensions, err := r.extensions.List()
	if err != nil {
		return
	}
	for _, ext := range extensions {
		if r.extensionConflict(root, ext.Name) != "" {
			continue
		}
		extName := ext.Name
		root.AddCommand(&cobra.Command{
			Use:                extName,
			Short:              fmt.Sprintf("Extension %s", extName),
			DisableFlagParsing: true,
			Annotations:        map[string]string{"yunxiao-extension": "true"},
			RunE: func(cmd *cobra.Command, args []string) error {
				return r.dispatchExtension(cmd.Context(), extName, args)
			},
		})
	}
}

func (r *Root) dispatchExtension(ctx context.Context, name string, args []string) error {
	if r.extensions == nil {
		return fmt.Errorf("extension manager is not configured")
	}
	resolved, err := r.resolvedConfig()
	if err != nil {
		return err
	}
	handled, err := r.extensions.Dispatch(ctx, extension.DispatchRequest{
		Name: name,
		Args: args,
		IO:   r.io,
		Context: extension.ExecutionContext{
			Endpoint:     resolved.Endpoint.Value,
			Organization: resolved.Organization.Value,
			Project:      resolved.Project.Value,
			Repo:         resolved.Repo.Value,
		},
	})
	if err != nil {
		return err
	}
	if !handled {
		return fmt.Errorf("no extension found: %s", name)
	}
	return nil
}

func (r *Root) annotateExtensionConflicts(extensions []extension.Extension) {
	if r.rootCmd == nil {
		return
	}
	for index := range extensions {
		extensions[index].Conflict = r.extensionConflict(r.rootCmd, extensions[index].Name)
	}
}

func (r *Root) extensionConflict(root *cobra.Command, name string) string {
	if rootCommandNameExists(root, name) {
		return "core command"
	}
	file, err := r.configStore.Load(config.ScopeGlobal)
	if err == nil {
		for _, alias := range file.Aliases {
			if alias.Name == name {
				return "alias"
			}
		}
	}
	return ""
}

func rootCommandNameExists(root *cobra.Command, name string) bool {
	for _, command := range root.Commands() {
		if command.Annotations["yunxiao-extension"] == "true" {
			continue
		}
		if command.Name() == name {
			return true
		}
		for _, alias := range command.Aliases {
			if alias == name {
				return true
			}
		}
	}
	return false
}

func (r *Root) confirmExtensionInstall(yes bool) error {
	if yes {
		return nil
	}
	caps := terminal.DetectCapabilities()
	if !caps.StdinTTY {
		return fmt.Errorf("extension installation runs unverified local code; pass --yes to confirm in non-interactive mode")
	}
	fmt.Fprint(r.io.ErrOut, "Extensions are not verified by Yunxiao CLI. Install anyway? [y/N] ")
	reader := bufio.NewReader(r.io.In)
	answer, err := reader.ReadString('\n')
	if err != nil && strings.TrimSpace(answer) == "" {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("extension installation cancelled")
	}
}

func extensionSource(ext extension.Extension) string {
	if ext.Kind == extension.KindLocal {
		return ext.LocalPath
	}
	return ext.SourceURL
}

func extensionVersion(ext extension.Extension) string {
	if ext.SourceRef != "" {
		return ext.SourceRef
	}
	if ext.CurrentCommit != "" {
		if len(ext.CurrentCommit) > 12 {
			return ext.CurrentCommit[:12]
		}
		return ext.CurrentCommit
	}
	return ""
}
