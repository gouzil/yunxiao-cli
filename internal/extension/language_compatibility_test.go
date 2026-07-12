package extension

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gouzil/yunxiao-cli/internal/terminal"
)

const languageCompatibilityEnv = "YUNXIAO_TEST_EXTENSION_LANGUAGE_COMPATIBILITY"

func TestExtensionLanguageCompatibility(t *testing.T) {
	if os.Getenv(languageCompatibilityEnv) != "1" {
		t.Skip("set " + languageCompatibilityEnv + "=1 to run")
	}
	if runtime.GOOS == "windows" {
		t.Fatal("the language compatibility matrix requires Unix executable entries")
	}
	for _, name := range []string{"bash", "python3", "node", "go"} {
		if _, err := exec.LookPath(name); err != nil {
			t.Fatalf("required runtime %q not found: %v", name, err)
		}
	}

	root := t.TempDir()
	sources := filepath.Join(root, "sources")
	manager := NewManager(Options{
		Dirs:  Dirs{DataDir: filepath.Join(root, "data"), StateDir: filepath.Join(root, "state")},
		Env:   fakeEnv{"CI": "1"},
		Exec:  RealExecRunner{GOOS: runtime.GOOS},
		GOOS:  runtime.GOOS,
		Clock: staticClock{},
	})

	fixtures := []struct {
		name   string
		source string
		entry  string
		build  bool
	}{
		{name: "shell", source: "yunxiao-shell", entry: "yunxiao-shell"},
		{name: "python", source: "yunxiao-python", entry: "yunxiao-python"},
		{name: "node", source: "yunxiao-node", entry: "yunxiao-node"},
		{name: "go", source: "yunxiao-go", entry: "main.go", build: true},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			sourceDir := filepath.Join(sources, fixture.source)
			entry := filepath.Join(sourceDir, fixture.entry)
			copyFixture(t, filepath.Join("..", "..", "test", "extensions", fixture.source, fixture.entry), entry, fixture.build)
			if fixture.build {
				binary := filepath.Join(sourceDir, fixture.source)
				command := exec.CommandContext(context.Background(), "go", "build", "-o", binary, entry)
				if output, err := command.CombinedOutput(); err != nil {
					t.Fatalf("build Go fixture: %v\n%s", err, output)
				}
			}

			if _, err := manager.InstallLocal(context.Background(), sourceDir); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			handled, err := manager.Dispatch(context.Background(), DispatchRequest{
				Name: fixture.name,
				Args: []string{"alpha", "two words"},
				IO:   terminal.IOStreams{In: strings.NewReader(""), Out: &stdout, ErrOut: &stderr},
				Context: ExecutionContext{
					Endpoint:     "https://example.test",
					Organization: "org-1",
					Project:      "project-1",
					Repo:         "repo-1",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if !handled {
				t.Fatal("extension was not dispatched")
			}
			wantStdout := fmt.Sprintf("args=alpha|two words\nYUNXIAO_EXTENSION=1\nYUNXIAO_EXTENSION_NAME=%s\nYUNXIAO_EXTENSION_DIR=%s\nYUNXIAO_ENDPOINT=https://example.test\nYUNXIAO_ORGANIZATION=org-1\nYUNXIAO_PROJECT=project-1\nYUNXIAO_REPO=repo-1\n", fixture.name, sourceDir)
			if stdout.String() != wantStdout {
				t.Fatalf("stdout:\n%s\nwant:\n%s", stdout.String(), wantStdout)
			}
			if got, want := stderr.String(), "stderr="+fixture.name+"\n"; got != want {
				t.Fatalf("stderr = %q, want %q", got, want)
			}
		})
	}
}

func copyFixture(t *testing.T, source, target string, goSource bool) {
	t.Helper()
	body, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	mode := os.FileMode(0o755)
	if goSource {
		mode = 0o644
	}
	if err := os.WriteFile(target, body, mode); err != nil {
		t.Fatal(err)
	}
}
