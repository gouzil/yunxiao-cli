package terminal

import (
	"io"
	"os"
	"time"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/glamour"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

func SystemIO() IOStreams {
	return IOStreams{
		In:     os.Stdin,
		Out:    colorable.NewColorableStdout(),
		ErrOut: colorable.NewColorableStderr(),
	}
}

type Capabilities struct {
	StdinTTY  bool `json:"stdinTty"`
	StdoutTTY bool `json:"stdoutTty"`
	StderrTTY bool `json:"stderrTty"`
	Color     bool `json:"color"`
}

func DetectCapabilities() Capabilities {
	stdoutTTY := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	return Capabilities{
		StdinTTY:  term.IsTerminal(int(os.Stdin.Fd())),
		StdoutTTY: stdoutTTY,
		StderrTTY: isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd()),
		Color:     stdoutTTY,
	}
}

type Spinner struct {
	value *spinner.Spinner
}

func NewSpinner(message string) Spinner {
	value := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	value.Suffix = " " + message
	return Spinner{value: value}
}

func (s Spinner) Start() {
	if s.value != nil {
		s.value.Start()
	}
}

func (s Spinner) Stop() {
	if s.value != nil {
		s.value.Stop()
	}
}

type MarkdownRenderer struct {
	plain bool
}

func NewMarkdownRenderer(plain bool) MarkdownRenderer {
	return MarkdownRenderer{plain: plain}
}

func (r MarkdownRenderer) Render(markdown string) (string, error) {
	if r.plain {
		return markdown, nil
	}
	return glamour.Render(markdown, "auto")
}

type Prompt interface {
	Input(title string, value *string) error
	Confirm(title string, value *bool) error
	Form(title string, fields []PromptField) error
}

type PromptField struct {
	Label string  `json:"label"`
	Value *string `json:"-"`
}

type HuhPrompt struct{}

func (HuhPrompt) Input(title string, value *string) error {
	return huh.NewInput().Title(title).Value(value).Run()
}

func (HuhPrompt) Confirm(title string, value *bool) error {
	return huh.NewConfirm().Title(title).Value(value).Run()
}

func (prompt HuhPrompt) Form(title string, fields []PromptField) error {
	for _, field := range fields {
		if err := prompt.Input(field.Label, field.Value); err != nil {
			return err
		}
	}
	return nil
}

var (
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type StateModel interface {
	Init() string
	Update(message StateMessage) (StateModel, StateCommand)
	View() string
}

type StateMessage struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type StateCommand struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type StateProgramFactory interface {
	NewProgram(model StateModel) StateProgram
}

type StateProgram interface {
	Run() (StateModel, error)
}
