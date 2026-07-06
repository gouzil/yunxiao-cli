package browser

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

type Opener interface {
	Open(url string) error
}

type SystemOpener struct{}

func (SystemOpener) Open(rawURL string) error {
	var command string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		command = "open"
		args = []string{rawURL}
	case "windows":
		command = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", rawURL}
	default:
		command = "xdg-open"
		args = []string{rawURL}
	}
	return exec.Command(command, args...).Start()
}

type Environment struct {
	GOOS    string
	Display string
	Wayland string
	Term    string
}

func CurrentEnvironment() Environment {
	return Environment{
		GOOS:    runtime.GOOS,
		Display: os.Getenv("DISPLAY"),
		Wayland: os.Getenv("WAYLAND_DISPLAY"),
		Term:    os.Getenv("TERM"),
	}
}

func CanOpenGraphically(env Environment) bool {
	if env.GOOS == "darwin" || env.GOOS == "windows" {
		return true
	}
	return env.Display != "" || env.Wayland != ""
}

func OpenOrPrint(opener Opener, rawURL string, env Environment, out io.Writer) error {
	if !CanOpenGraphically(env) {
		_, err := fmt.Fprintln(out, rawURL)
		return err
	}
	if err := opener.Open(rawURL); err != nil {
		_, writeErr := fmt.Fprintln(out, rawURL)
		if writeErr != nil {
			return writeErr
		}
		return err
	}
	_, err := fmt.Fprintf(out, "Opening %s in your browser.\n", rawURL)
	return err
}
