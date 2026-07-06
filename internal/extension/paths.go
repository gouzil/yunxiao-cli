package extension

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/gouzi/yunxiao-cli/internal/config"
)

const (
	EnvConfigDir = "YUNXIAO_CONFIG_DIR"
	EnvXDGData   = "XDG_DATA_HOME"
	EnvXDGState  = "XDG_STATE_HOME"
)

var shortNamePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type Dirs struct {
	DataDir  string
	StateDir string
}

func DefaultDirs(home string, env config.Env) Dirs {
	if env == nil {
		env = config.OSEnv{}
	}
	if raw, ok := env.LookupEnv(EnvConfigDir); ok && strings.TrimSpace(raw) != "" {
		base := strings.TrimSpace(raw)
		return Dirs{
			DataDir:  filepath.Join(base, "data", "extensions"),
			StateDir: filepath.Join(base, "state", "extensions"),
		}
	}
	if runtime.GOOS == "windows" {
		localAppData := ""
		if raw, ok := env.LookupEnv("LocalAppData"); ok {
			localAppData = strings.TrimSpace(raw)
		}
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return Dirs{
			DataDir:  filepath.Join(localAppData, "yunxiao", "extensions"),
			StateDir: filepath.Join(localAppData, "yunxiao", "state", "extensions"),
		}
	}
	dataBase := filepath.Join(home, ".local", "share")
	if raw, ok := env.LookupEnv(EnvXDGData); ok && strings.TrimSpace(raw) != "" {
		dataBase = strings.TrimSpace(raw)
	}
	stateBase := filepath.Join(home, ".local", "state")
	if raw, ok := env.LookupEnv(EnvXDGState); ok && strings.TrimSpace(raw) != "" {
		stateBase = strings.TrimSpace(raw)
	}
	return Dirs{
		DataDir:  filepath.Join(dataBase, "yunxiao", "extensions"),
		StateDir: filepath.Join(stateBase, "yunxiao", "extensions"),
	}
}

func NormalizeName(value string) (short string, full string, err error) {
	name := strings.TrimSpace(value)
	if name == "" {
		return "", "", fmt.Errorf("extension name is required")
	}
	if strings.HasPrefix(name, Prefix) {
		name = strings.TrimPrefix(name, Prefix)
	}
	if !shortNamePattern.MatchString(name) {
		return "", "", fmt.Errorf("invalid extension name %q: use lowercase letters, digits, and single hyphens", value)
	}
	return name, Prefix + name, nil
}

func NameFromSource(source string) (short string, full string, err error) {
	clean := strings.TrimSpace(source)
	if clean == "" {
		return "", "", fmt.Errorf("extension source is required")
	}
	base := clean
	if stat, statErr := os.Stat(clean); statErr == nil && stat.IsDir() {
		base = filepath.Base(filepath.Clean(clean))
	} else {
		base = remoteBase(clean)
	}
	base = strings.TrimSuffix(base, ".git")
	if !strings.HasPrefix(base, Prefix) {
		return "", "", fmt.Errorf("extension source %q must be named %s<name>", source, Prefix)
	}
	return NormalizeName(base)
}

func remoteBase(source string) string {
	trimmed := strings.TrimRight(source, "/")
	if strings.Contains(trimmed, "://") {
		return path.Base(trimmed)
	}
	if index := strings.LastIndexAny(trimmed, "/:"); index >= 0 {
		return trimmed[index+1:]
	}
	return filepath.Base(trimmed)
}

func IsRemoteSource(source string) bool {
	value := strings.TrimSpace(source)
	if value == "" {
		return false
	}
	if stat, err := os.Stat(value); err == nil {
		return !stat.IsDir()
	}
	return strings.Contains(value, "://") || strings.Contains(value, "@") || strings.HasSuffix(value, ".git")
}
