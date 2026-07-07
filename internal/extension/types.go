package extension

import (
	"context"
	"io"
	"time"

	"github.com/gouzil/yunxiao-cli/internal/terminal"
)

const (
	Prefix       = "yunxiao-"
	ManifestName = "extension.json"
	StateName    = "state.yml"
)

type Kind string

const (
	KindGit   Kind = "git"
	KindLocal Kind = "local"
)

type TemplateType string

const (
	TemplateScript TemplateType = "script"
	TemplateGo     TemplateType = "go"
)

type Extension struct {
	Name           string `json:"name"`
	FullName       string `json:"fullName"`
	Kind           Kind   `json:"kind"`
	ExecutablePath string `json:"executablePath"`
	SourceURL      string `json:"sourceURL,omitempty"`
	SourceRef      string `json:"sourceRef,omitempty"`
	CurrentCommit  string `json:"currentCommit,omitempty"`
	Pinned         bool   `json:"pinned"`
	LocalPath      string `json:"localPath,omitempty"`
	Conflict       string `json:"conflict,omitempty"`
}

type manifest struct {
	Version        int    `json:"version"`
	Name           string `json:"name"`
	FullName       string `json:"fullName"`
	Kind           Kind   `json:"kind"`
	ExecutablePath string `json:"executablePath"`
	SourceURL      string `json:"sourceURL,omitempty"`
	SourceRef      string `json:"sourceRef,omitempty"`
	CurrentCommit  string `json:"currentCommit,omitempty"`
	Pinned         bool   `json:"pinned"`
	LocalPath      string `json:"localPath,omitempty"`
}

type InstallSource struct {
	Source string
	Ref    string
	Yes    bool
}

type ExecutionContext struct {
	Endpoint     string
	Organization string
	Project      string
	Repo         string
}

type DispatchRequest struct {
	Name    string
	Args    []string
	IO      terminal.IOStreams
	Context ExecutionContext
}

type Manager interface {
	List() ([]Extension, error)
	Install(ctx context.Context, source InstallSource) (Extension, error)
	InstallLocal(ctx context.Context, dir string) (Extension, error)
	Upgrade(ctx context.Context, name string, force bool) error
	Remove(name string) error
	Dispatch(ctx context.Context, request DispatchRequest) (bool, error)
	Create(ctx context.Context, name string, template TemplateType) error
	UpdateDir(name string) string
}

type GitRunner interface {
	Run(ctx context.Context, dir string, args ...string) (string, error)
}

type ExecRunner interface {
	Run(ctx context.Context, executable string, args []string, env []string, streams terminal.IOStreams) error
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "extension exited with non-zero status"
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

type writeCloser interface {
	io.Writer
	io.Closer
}
