package extension

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gouzil/yunxiao-cli/internal/config"
	"github.com/gouzil/yunxiao-cli/internal/terminal"
)

const localPathFile = ".local-path"

type Options struct {
	Dirs      Dirs
	Home      string
	Env       config.Env
	Git       GitRunner
	Exec      ExecRunner
	Clock     Clock
	GOOS      string
	FindShell func() (string, error)
}

type LocalManager struct {
	dirs      Dirs
	env       config.Env
	git       GitRunner
	exec      ExecRunner
	clock     Clock
	goos      string
	findShell func() (string, error)
}

func NewManager(options Options) *LocalManager {
	env := options.Env
	if env == nil {
		env = config.OSEnv{}
	}
	dirs := options.Dirs
	if dirs.DataDir == "" || dirs.StateDir == "" {
		dirs = DefaultDirs(options.Home, env)
	}
	goos := options.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	git := options.Git
	if git == nil {
		git = RealGitRunner{}
	}
	findShell := options.FindShell
	if findShell == nil {
		findShell = defaultFindShell
	}
	execRunner := options.Exec
	if execRunner == nil {
		execRunner = RealExecRunner{GOOS: goos, FindShell: findShell}
	}
	clock := options.Clock
	if clock == nil {
		clock = realClock{}
	}
	return &LocalManager{
		dirs:      dirs,
		env:       env,
		git:       git,
		exec:      execRunner,
		clock:     clock,
		goos:      goos,
		findShell: findShell,
	}
}

func (m *LocalManager) List() ([]Extension, error) {
	entries, err := os.ReadDir(m.dirs.DataDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	extensions := make([]Extension, 0, len(entries))
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), Prefix) {
			continue
		}
		ext, err := m.readInstalled(entry.Name())
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, ext)
	}
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Name < extensions[j].Name
	})
	return extensions, nil
}

func (m *LocalManager) Install(ctx context.Context, source InstallSource) (Extension, error) {
	if !IsRemoteSource(source.Source) {
		return m.InstallLocal(ctx, source.Source)
	}
	return m.installGit(ctx, source)
}

func (m *LocalManager) InstallLocal(ctx context.Context, dir string) (Extension, error) {
	abs, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return Extension{}, err
	}
	stat, err := os.Stat(abs)
	if err != nil {
		return Extension{}, err
	}
	if !stat.IsDir() {
		return Extension{}, fmt.Errorf("local extension source must be a directory: %s", dir)
	}
	short, full, err := NameFromSource(abs)
	if err != nil {
		return Extension{}, err
	}
	sourceExecutable := filepath.Join(abs, m.executableName(full))
	if err := requireExecutable(sourceExecutable, m.goos); err != nil {
		return Extension{}, err
	}
	targetDir := filepath.Join(m.dirs.DataDir, full)
	if err := ensureNotInstalled(targetDir, full); err != nil {
		return Extension{}, err
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return Extension{}, err
	}
	executablePath := sourceExecutable
	if m.goos == "windows" {
		if err := os.WriteFile(filepath.Join(targetDir, localPathFile), []byte(abs+"\n"), 0o600); err != nil {
			return Extension{}, err
		}
	} else {
		executablePath = filepath.Join(targetDir, full)
		if err := os.Symlink(sourceExecutable, executablePath); err != nil {
			return Extension{}, err
		}
	}
	ext := Extension{
		Name:           short,
		FullName:       full,
		Kind:           KindLocal,
		ExecutablePath: executablePath,
		LocalPath:      abs,
	}
	if err := m.writeManifest(targetDir, ext); err != nil {
		return Extension{}, err
	}
	return ext, nil
}

func (m *LocalManager) Upgrade(ctx context.Context, name string, force bool) error {
	extensions, err := m.List()
	if err != nil {
		return err
	}
	if len(extensions) == 0 {
		return fmt.Errorf("no installed extensions found")
	}
	selected := extensions
	if strings.TrimSpace(name) != "" {
		short, _, err := NormalizeName(name)
		if err != nil {
			return err
		}
		selected = nil
		for _, ext := range extensions {
			if ext.Name == short {
				selected = []Extension{ext}
				break
			}
		}
		if len(selected) == 0 {
			return fmt.Errorf("no extension matched %q", name)
		}
	}
	var failed []string
	for _, ext := range selected {
		if err := m.upgradeOne(ctx, ext, force); err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", ext.Name, err))
		}
	}
	if len(failed) > 0 {
		return errors.New(strings.Join(failed, "; "))
	}
	return nil
}

func (m *LocalManager) Remove(name string) error {
	_, full, err := NormalizeName(name)
	if err != nil {
		return err
	}
	targetDir := filepath.Join(m.dirs.DataDir, full)
	if _, err := os.Lstat(targetDir); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("no extension found: %s", strings.TrimPrefix(full, Prefix))
	} else if err != nil {
		return err
	}
	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}
	return os.RemoveAll(m.UpdateDir(full))
}

func (m *LocalManager) Dispatch(ctx context.Context, request DispatchRequest) (bool, error) {
	short, full, err := NormalizeName(request.Name)
	if err != nil {
		return false, err
	}
	ext, err := m.readInstalled(full)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	notice := m.startUpdateCheck(ctx, ext)
	err = m.exec.Run(ctx, ext.ExecutablePath, request.Args, extensionEnv(short, ext, request.Context), request.IO)
	m.printUpdateNotice(notice, request.IO.ErrOut)
	if err != nil {
		return true, err
	}
	return true, nil
}

func (m *LocalManager) Create(ctx context.Context, name string, template TemplateType) error {
	_, full, err := NormalizeName(name)
	if err != nil {
		return err
	}
	if template == "" {
		template = TemplateScript
	}
	if template != TemplateScript && template != TemplateGo {
		return fmt.Errorf("unsupported extension template %q", template)
	}
	if _, err := os.Stat(full); err == nil {
		return fmt.Errorf("directory already exists: %s", full)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(full, 0o755); err != nil {
		return err
	}
	if template == TemplateGo {
		return createGoTemplate(full)
	}
	return createScriptTemplate(full)
}

func (m *LocalManager) UpdateDir(name string) string {
	_, full, err := NormalizeName(name)
	if err != nil {
		full = name
	}
	return filepath.Join(m.dirs.StateDir, full)
}

func (m *LocalManager) installGit(ctx context.Context, source InstallSource) (Extension, error) {
	short, full, err := NameFromSource(source.Source)
	if err != nil {
		return Extension{}, err
	}
	targetDir := filepath.Join(m.dirs.DataDir, full)
	if err := ensureNotInstalled(targetDir, full); err != nil {
		return Extension{}, err
	}
	tmpDir, err := os.MkdirTemp("", "yunxiao-extension-*")
	if err != nil {
		return Extension{}, err
	}
	defer os.RemoveAll(tmpDir)
	cloneDir := filepath.Join(tmpDir, full)
	if _, err := m.git.Run(ctx, "", "clone", source.Source, cloneDir); err != nil {
		return Extension{}, err
	}
	if strings.TrimSpace(source.Ref) != "" {
		if _, err := m.git.Run(ctx, cloneDir, "checkout", source.Ref); err != nil {
			return Extension{}, err
		}
	}
	executablePath := filepath.Join(cloneDir, m.executableName(full))
	if err := requireExecutable(executablePath, m.goos); err != nil {
		return Extension{}, err
	}
	commit, _ := m.git.Run(ctx, cloneDir, "rev-parse", "HEAD")
	if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
		return Extension{}, err
	}
	if err := os.Rename(cloneDir, targetDir); err != nil {
		return Extension{}, err
	}
	ext := Extension{
		Name:           short,
		FullName:       full,
		Kind:           KindGit,
		ExecutablePath: filepath.Join(targetDir, m.executableName(full)),
		SourceURL:      source.Source,
		SourceRef:      strings.TrimSpace(source.Ref),
		CurrentCommit:  strings.TrimSpace(commit),
		Pinned:         strings.TrimSpace(source.Ref) != "",
	}
	if err := m.writeManifest(targetDir, ext); err != nil {
		return Extension{}, err
	}
	_ = os.RemoveAll(m.UpdateDir(full))
	return ext, nil
}

func (m *LocalManager) upgradeOne(ctx context.Context, ext Extension, force bool) error {
	if ext.Kind == KindLocal {
		return fmt.Errorf("local extensions can not be upgraded")
	}
	if ext.Pinned && !force {
		return fmt.Errorf("pinned extensions can not be upgraded")
	}
	dir := filepath.Join(m.dirs.DataDir, ext.FullName)
	if force {
		if _, err := m.git.Run(ctx, dir, "fetch", "origin", "HEAD"); err != nil {
			return err
		}
		if _, err := m.git.Run(ctx, dir, "reset", "--hard", "FETCH_HEAD"); err != nil {
			return err
		}
		ext.Pinned = false
		ext.SourceRef = ""
	} else if _, err := m.git.Run(ctx, dir, "pull", "--ff-only"); err != nil {
		return err
	}
	commit, _ := m.git.Run(ctx, dir, "rev-parse", "HEAD")
	ext.CurrentCommit = strings.TrimSpace(commit)
	ext.ExecutablePath = filepath.Join(dir, m.executableName(ext.FullName))
	if err := requireExecutable(ext.ExecutablePath, m.goos); err != nil {
		return err
	}
	if err := m.writeManifest(dir, ext); err != nil {
		return err
	}
	return os.RemoveAll(m.UpdateDir(ext.FullName))
}

func (m *LocalManager) readInstalled(full string) (Extension, error) {
	dir := filepath.Join(m.dirs.DataDir, full)
	ext, err := readManifest(filepath.Join(dir, ManifestName))
	if err != nil {
		return Extension{}, err
	}
	if ext.Kind == KindLocal && m.goos == "windows" && ext.LocalPath != "" {
		ext.ExecutablePath = filepath.Join(ext.LocalPath, m.executableName(ext.FullName))
	}
	return ext, nil
}

func (m *LocalManager) writeManifest(dir string, ext Extension) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data := manifest{
		Version:        1,
		Name:           ext.Name,
		FullName:       ext.FullName,
		Kind:           ext.Kind,
		ExecutablePath: ext.ExecutablePath,
		SourceURL:      ext.SourceURL,
		SourceRef:      ext.SourceRef,
		CurrentCommit:  ext.CurrentCommit,
		Pinned:         ext.Pinned,
		LocalPath:      ext.LocalPath,
	}
	body, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(filepath.Join(dir, ManifestName), body, 0o600)
}

func readManifest(path string) (Extension, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return Extension{}, err
	}
	var data manifest
	if err := json.Unmarshal(body, &data); err != nil {
		return Extension{}, fmt.Errorf("parse extension manifest %s: %w", path, err)
	}
	if data.Version != 1 {
		return Extension{}, fmt.Errorf("unsupported extension manifest version %d in %s", data.Version, path)
	}
	if data.Name == "" || data.FullName == "" || data.ExecutablePath == "" {
		return Extension{}, fmt.Errorf("invalid extension manifest %s: missing required fields", path)
	}
	if data.Kind != KindGit && data.Kind != KindLocal {
		return Extension{}, fmt.Errorf("invalid extension manifest %s: unsupported kind %q", path, data.Kind)
	}
	return Extension{
		Name:           data.Name,
		FullName:       data.FullName,
		Kind:           data.Kind,
		ExecutablePath: data.ExecutablePath,
		SourceURL:      data.SourceURL,
		SourceRef:      data.SourceRef,
		CurrentCommit:  data.CurrentCommit,
		Pinned:         data.Pinned,
		LocalPath:      data.LocalPath,
	}, nil
}

func (m *LocalManager) executableName(full string) string {
	if m.goos == "windows" {
		return full + ".exe"
	}
	return full
}

func ensureNotInstalled(targetDir string, full string) error {
	if _, err := os.Lstat(targetDir); err == nil {
		return fmt.Errorf("extension already installed: %s", strings.TrimPrefix(full, Prefix))
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func requireExecutable(path string, goos string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("extension executable not found: %s", path)
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("extension executable is a directory: %s", path)
	}
	if goos != "windows" && info.Mode()&0o111 == 0 {
		return fmt.Errorf("extension executable is not executable: %s", path)
	}
	return nil
}

func extensionEnv(name string, ext Extension, ctx ExecutionContext) []string {
	env := []string{
		"YUNXIAO_EXTENSION=1",
		"YUNXIAO_EXTENSION_NAME=" + name,
		"YUNXIAO_EXTENSION_DIR=" + extensionDir(ext),
	}
	if ctx.Endpoint != "" {
		env = append(env, "YUNXIAO_ENDPOINT="+ctx.Endpoint)
	}
	if ctx.Organization != "" {
		env = append(env, "YUNXIAO_ORGANIZATION="+ctx.Organization)
	}
	if ctx.Project != "" {
		env = append(env, "YUNXIAO_PROJECT="+ctx.Project)
	}
	if ctx.Repo != "" {
		env = append(env, "YUNXIAO_REPO="+ctx.Repo)
	}
	return env
}

func extensionDir(ext Extension) string {
	if ext.Kind == KindLocal && ext.LocalPath != "" {
		return ext.LocalPath
	}
	return filepath.Dir(ext.ExecutablePath)
}

func (m *LocalManager) startUpdateCheck(ctx context.Context, ext Extension) <-chan string {
	if ext.Kind != KindGit || !m.shouldCheckForUpdate() {
		return nil
	}
	statePath := filepath.Join(m.UpdateDir(ext.FullName), StateName)
	state, _ := readState(statePath)
	if !state.CheckedAt.IsZero() && m.clock.Now().Sub(state.CheckedAt) < 24*time.Hour {
		return nil
	}
	ch := make(chan string, 1)
	go func() {
		defer close(ch)
		dir := filepath.Join(m.dirs.DataDir, ext.FullName)
		local, _ := m.git.Run(ctx, dir, "rev-parse", "HEAD")
		remote, err := m.git.Run(ctx, dir, "ls-remote", "origin", "HEAD")
		_ = writeState(statePath, updateState{CheckedAt: m.clock.Now(), LatestHead: strings.TrimSpace(remote)})
		if err != nil {
			return
		}
		remoteHead := strings.Fields(remote)
		localHead := strings.TrimSpace(local)
		if len(remoteHead) > 0 && localHead != "" && remoteHead[0] != localHead {
			ch <- fmt.Sprintf("A new version of %s is available. Run: yunxiao extension upgrade %s\n", ext.Name, ext.Name)
		}
	}()
	return ch
}

func (m *LocalManager) shouldCheckForUpdate() bool {
	if value, ok := m.env.LookupEnv("YUNXIAO_NO_EXTENSION_UPDATE_NOTIFIER"); ok && strings.TrimSpace(value) != "" {
		return false
	}
	if value, ok := m.env.LookupEnv("CI"); ok && strings.TrimSpace(value) != "" {
		return false
	}
	caps := terminal.DetectCapabilities()
	return caps.StdoutTTY && caps.StderrTTY
}

func (m *LocalManager) printUpdateNotice(ch <-chan string, errOut ioWriter) {
	if ch == nil || errOut == nil {
		return
	}
	select {
	case message := <-ch:
		if message != "" {
			fmt.Fprint(errOut, message)
		}
	default:
	}
}

type ioWriter interface {
	Write([]byte) (int, error)
}
