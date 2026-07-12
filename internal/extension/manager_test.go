package extension

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gouzil/yunxiao-cli/internal/terminal"
)

func TestNormalizeNameAndSource(t *testing.T) {
	short, full, err := NormalizeName("yunxiao-team-report")
	if err != nil {
		t.Fatal(err)
	}
	if short != "team-report" || full != "yunxiao-team-report" {
		t.Fatalf("name = %q %q", short, full)
	}
	if _, _, err := NormalizeName("Team_Report"); err == nil {
		t.Fatal("expected invalid name error")
	}
	short, full, err = NameFromSource("git@codeup.aliyun.com:org/yunxiao-team-report.git")
	if err != nil {
		t.Fatal(err)
	}
	if short != "team-report" || full != "yunxiao-team-report" {
		t.Fatalf("source name = %q %q", short, full)
	}
}

func TestDefaultDirs(t *testing.T) {
	dirs := DefaultDirs("/home/test", fakeEnv{EnvXDGData: "/xdg/data", EnvXDGState: "/xdg/state"})
	wantDataDir := filepath.Join("/xdg/data", "yunxiao", "extensions")
	wantStateDir := filepath.Join("/xdg/state", "yunxiao", "extensions")
	if runtime.GOOS == "windows" {
		wantDataDir = filepath.Join("/home/test", "AppData", "Local", "yunxiao", "extensions")
		wantStateDir = filepath.Join("/home/test", "AppData", "Local", "yunxiao", "state", "extensions")
	}
	if dirs.DataDir != wantDataDir {
		t.Fatalf("data dir = %q", dirs.DataDir)
	}
	if dirs.StateDir != wantStateDir {
		t.Fatalf("state dir = %q", dirs.StateDir)
	}
	dirs = DefaultDirs("/home/test", fakeEnv{EnvConfigDir: "/cfg"})
	if dirs.DataDir != filepath.Join("/cfg", "data", "extensions") {
		t.Fatalf("config data dir = %q", dirs.DataDir)
	}
}

func TestInstallLocalListRemoveAndDispatch(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "yunxiao-hello")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(source, testExecutableName("yunxiao-hello"))
	if err := os.WriteFile(executable, []byte("#!/usr/bin/env bash\necho hi\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	exec := &fakeExecRunner{}
	manager := NewManager(Options{
		Dirs:  Dirs{DataDir: filepath.Join(dir, "data"), StateDir: filepath.Join(dir, "state")},
		Env:   fakeEnv{"CI": "1"},
		Exec:  exec,
		GOOS:  runtime.GOOS,
		Clock: staticClock{},
	})
	ext, err := manager.InstallLocal(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	if ext.Name != "hello" || ext.Kind != KindLocal {
		t.Fatalf("extension = %#v", ext)
	}
	extensions, err := manager.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(extensions) != 1 || extensions[0].Name != "hello" {
		t.Fatalf("extensions = %#v", extensions)
	}
	out := &bytes.Buffer{}
	handled, err := manager.Dispatch(context.Background(), DispatchRequest{
		Name: "hello",
		Args: []string{"--name", "world"},
		IO:   terminal.IOStreams{In: strings.NewReader(""), Out: out, ErrOut: &bytes.Buffer{}},
		Context: ExecutionContext{
			Endpoint:     "endpoint",
			Organization: "org",
			Project:      "project",
			Repo:         "repo",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected dispatch to handle extension")
	}
	wantExecutable := filepath.Join(manager.dirs.DataDir, "yunxiao-hello", testExecutableName("yunxiao-hello"))
	if runtime.GOOS == "windows" {
		wantExecutable = executable
	}
	if exec.executable != wantExecutable {
		t.Fatalf("executable = %q", exec.executable)
	}
	if !reflect.DeepEqual(exec.args, []string{"--name", "world"}) {
		t.Fatalf("args = %#v", exec.args)
	}
	wantEnv := []string{
		"YUNXIAO_EXTENSION=1",
		"YUNXIAO_EXTENSION_NAME=hello",
		"YUNXIAO_EXTENSION_DIR=" + source,
		"YUNXIAO_ENDPOINT=endpoint",
		"YUNXIAO_ORGANIZATION=org",
		"YUNXIAO_PROJECT=project",
		"YUNXIAO_REPO=repo",
	}
	if !reflect.DeepEqual(exec.env, wantEnv) {
		t.Fatalf("env = %#v, want %#v", exec.env, wantEnv)
	}
	if err := manager.Remove("hello"); err != nil {
		t.Fatal(err)
	}
	extensions, err = manager.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(extensions) != 0 {
		t.Fatalf("extensions after remove = %#v", extensions)
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatalf("local source was removed: %v", err)
	}
}

func TestInstallLocalWindowsAcceptsScriptEntry(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "yunxiao-script")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(source, "yunxiao-script")
	if err := os.WriteFile(entry, []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	exec := &fakeExecRunner{}
	manager := NewManager(Options{
		Dirs: Dirs{DataDir: filepath.Join(dir, "data"), StateDir: filepath.Join(dir, "state")},
		Env:  fakeEnv{"CI": "1"},
		Exec: exec,
		GOOS: "windows",
	})
	if _, err := manager.InstallLocal(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Dispatch(context.Background(), DispatchRequest{Name: "script", IO: terminal.IOStreams{}}); err != nil {
		t.Fatal(err)
	}
	if exec.executable != entry {
		t.Fatalf("executable = %q, want %q", exec.executable, entry)
	}
}

func TestInstallLocalWindowsPrefersBinaryEntry(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "yunxiao-binary")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"yunxiao-binary", "yunxiao-binary.exe"} {
		if err := os.WriteFile(filepath.Join(source, name), nil, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	exec := &fakeExecRunner{}
	manager := NewManager(Options{
		Dirs: Dirs{DataDir: filepath.Join(dir, "data"), StateDir: filepath.Join(dir, "state")},
		Env:  fakeEnv{"CI": "1"},
		Exec: exec,
		GOOS: "windows",
	})
	if _, err := manager.InstallLocal(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Dispatch(context.Background(), DispatchRequest{Name: "binary", IO: terminal.IOStreams{}}); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(source, "yunxiao-binary.exe")
	if exec.executable != want {
		t.Fatalf("executable = %q, want %q", exec.executable, want)
	}
}

func TestInstallRemoteUsesGitAndWritesManifest(t *testing.T) {
	dir := t.TempDir()
	git := &fakeGitRunner{t: t}
	manager := NewManager(Options{
		Dirs: Dirs{DataDir: filepath.Join(dir, "data"), StateDir: filepath.Join(dir, "state")},
		Git:  git,
		Env:  fakeEnv{"CI": "1"},
		GOOS: runtime.GOOS,
	})
	ext, err := manager.Install(context.Background(), InstallSource{Source: "https://example.test/yunxiao-remote.git", Ref: "v1"})
	if err != nil {
		t.Fatal(err)
	}
	if !ext.Pinned || ext.SourceRef != "v1" || ext.CurrentCommit != "abc123" {
		t.Fatalf("extension = %#v", ext)
	}
	if !git.seen("clone https://example.test/yunxiao-remote.git") {
		t.Fatalf("git commands = %#v", git.commands)
	}
	if !git.seen("checkout v1") {
		t.Fatalf("git commands = %#v", git.commands)
	}
}

func TestCreateTemplates(t *testing.T) {
	dir := t.TempDir()
	manager := NewManager(Options{Dirs: Dirs{DataDir: filepath.Join(dir, "data"), StateDir: filepath.Join(dir, "state")}, Env: fakeEnv{"CI": "1"}})
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := manager.Create(context.Background(), "script-tool", TemplateScript); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "yunxiao-script-tool", "yunxiao-script-tool")); err != nil {
		t.Fatal(err)
	}
	if err := manager.Create(context.Background(), "go-tool", TemplateGo); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "yunxiao-go-tool", "main.go")); err != nil {
		t.Fatal(err)
	}
}

type fakeEnv map[string]string

func (f fakeEnv) LookupEnv(key string) (string, bool) {
	value, ok := f[key]
	return value, ok
}

type staticClock struct{}

func (staticClock) Now() time.Time {
	return time.Unix(1000, 0).UTC()
}

type fakeExecRunner struct {
	executable string
	args       []string
	env        []string
}

func (f *fakeExecRunner) Run(ctx context.Context, executable string, args []string, env []string, streams terminal.IOStreams) error {
	f.executable = executable
	f.args = append([]string(nil), args...)
	f.env = append([]string(nil), env...)
	return nil
}

type fakeGitRunner struct {
	t        *testing.T
	commands []string
}

func (f *fakeGitRunner) Run(ctx context.Context, dir string, args ...string) (string, error) {
	f.commands = append(f.commands, strings.Join(args, " "))
	if len(args) >= 3 && args[0] == "clone" {
		target := args[2]
		if err := os.MkdirAll(target, 0o755); err != nil {
			f.t.Fatal(err)
		}
		full := filepath.Base(target)
		if err := os.WriteFile(filepath.Join(target, testExecutableName(full)), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
			f.t.Fatal(err)
		}
		return "", nil
	}
	if strings.Join(args, " ") == "rev-parse HEAD" {
		return "abc123\n", nil
	}
	return "", nil
}

func (f *fakeGitRunner) seen(prefix string) bool {
	for _, command := range f.commands {
		if strings.HasPrefix(command, prefix) {
			return true
		}
	}
	return false
}

func testExecutableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
