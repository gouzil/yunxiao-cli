package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gouzi/yunxiao-cli/internal/api"
	"github.com/gouzi/yunxiao-cli/internal/auth"
	"github.com/gouzi/yunxiao-cli/internal/config"
	"github.com/gouzi/yunxiao-cli/internal/extension"
	"github.com/gouzi/yunxiao-cli/internal/keyring"
	"github.com/gouzi/yunxiao-cli/internal/output"
	"github.com/gouzi/yunxiao-cli/internal/terminal"
	"github.com/gouzi/yunxiao-cli/internal/yunxiao"
)

type BuildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

type Application struct {
	root *Root
}

func New(build BuildInfo) (*Application, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	ioStreams := terminal.SystemIO()
	store := config.NewStore(config.DefaultPaths(home, workingDir), config.OSEnv{})
	resolved, err := store.Resolve(config.Values{})
	if err != nil {
		return nil, err
	}
	credentialFile := filepath.Join(home, ".config", "yunxiao", "credentials.json")
	credentialStore := auth.NewCascadingStore(keyring.Backend{}, auth.NewFileCredentialStore(credentialFile))
	credential, _ := credentialStore.Get(resolved.Endpoint.Value)
	client := api.NewClient(api.ClientOptions{Endpoint: resolved.Endpoint.Value, Token: credential.Token, DebugOut: ioStreams.ErrOut})
	services := yunxiao.NewClientServices(client)
	manager := auth.NewManager(credentialStore, yunxiao.NewTokenVerifier(services))
	extensionManager := extension.NewManager(extension.Options{Home: home, Env: config.OSEnv{}})
	root := NewRoot(RootOptions{
		Build:       build,
		IO:          ioStreams,
		ConfigStore: store,
		AuthManager: manager,
		Extensions:  extensionManager,
		Services: yunxiao.ServiceSet{
			Auth:     services,
			Repo:     services,
			Branch:   services,
			Commit:   services,
			File:     services,
			SSHKey:   services,
			MR:       services,
			Pipeline: services,
			Run:      services,
			Project:  services,
			WorkItem: services,
			Search:   services,
			RawAPI:   services,
		},
		Renderer: output.NewRenderer(ioStreams.Out),
	})
	return &Application{root: root}, nil
}

func (a *Application) Execute(ctx context.Context, args []string) error {
	return a.root.Execute(ctx, args)
}

type ExitError struct {
	Code int
	Err  error
}

func (e ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("exit %d", e.Code)
	}
	return e.Err.Error()
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}
	return 1
}
