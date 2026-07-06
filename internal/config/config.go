package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvEndpoint     = "YUNXIAO_ENDPOINT"
	EnvOrganization = "YUNXIAO_ORGANIZATION"
	EnvProject      = "YUNXIAO_PROJECT"
	EnvRepo         = "YUNXIAO_REPO"
	EnvToken        = "YUNXIAO_TOKEN"
)

type Scope string

const (
	ScopeGlobal Scope = "global"
	ScopeRepo   Scope = "repo"
)

type Source string

const (
	SourceDefault Source = "default"
	SourceGlobal  Source = "global"
	SourceRepo    Source = "repo"
	SourceEnv     Source = "environment"
	SourceFlag    Source = "flag"
)

type Values struct {
	Endpoint     string `json:"endpoint,omitempty"`
	Organization string `json:"organization,omitempty"`
	Project      string `json:"project,omitempty"`
	Repo         string `json:"repo,omitempty"`
	GitProtocol  string `json:"gitProtocol,omitempty"`
}

type File struct {
	Version int     `json:"version"`
	Values  Values  `json:"values"`
	Aliases []Alias `json:"aliases,omitempty"`
}

type Alias struct {
	Name      string `json:"name"`
	Expansion string `json:"expansion"`
}

type Entry struct {
	Key    Key    `json:"key"`
	Value  string `json:"value"`
	Source Source `json:"source"`
}

type Key string

const (
	KeyEndpoint     Key = "endpoint"
	KeyOrganization Key = "organization"
	KeyProject      Key = "project"
	KeyRepo         Key = "repo"
	KeyGitProtocol  Key = "git_protocol"
)

type Paths struct {
	Global string `json:"global"`
	Repo   string `json:"repo"`
}

type Store struct {
	paths Paths
	env   Env
}

type Env interface {
	LookupEnv(key string) (string, bool)
}

type OSEnv struct{}

func (OSEnv) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func DefaultPaths(home string, workingDir string) Paths {
	return Paths{
		Global: filepath.Join(home, ".config", "yunxiao", "config.json"),
		Repo:   filepath.Join(workingDir, ".yunxiao", "config.json"),
	}
}

func NewStore(paths Paths, env Env) *Store {
	if env == nil {
		env = OSEnv{}
	}
	return &Store{paths: paths, env: env}
}

func (s *Store) Paths() Paths {
	return s.paths
}

func (s *Store) Load(scope Scope) (File, error) {
	path, err := s.pathFor(scope)
	if err != nil {
		return File{}, err
	}
	return readFile(path)
}

func (s *Store) Save(scope Scope, file File) error {
	path, err := s.pathFor(scope)
	if err != nil {
		return err
	}
	if file.Version == 0 {
		file.Version = 1
	}
	return writeFile(path, file)
}

func (s *Store) Set(scope Scope, key Key, value string) error {
	file, err := s.Load(scope)
	if err != nil {
		return err
	}
	setValue(&file.Values, key, value)
	return s.Save(scope, file)
}

func (s *Store) Get(scope Scope, key Key) (string, error) {
	file, err := s.Load(scope)
	if err != nil {
		return "", err
	}
	return valueFor(file.Values, key), nil
}

func (s *Store) List(explicit Values) ([]Entry, error) {
	resolved, err := s.Resolve(explicit)
	if err != nil {
		return nil, err
	}
	entries := []Entry{
		{Key: KeyEndpoint, Value: resolved.Endpoint.Value, Source: resolved.Endpoint.Source},
		{Key: KeyOrganization, Value: resolved.Organization.Value, Source: resolved.Organization.Source},
		{Key: KeyProject, Value: resolved.Project.Value, Source: resolved.Project.Source},
		{Key: KeyRepo, Value: resolved.Repo.Value, Source: resolved.Repo.Source},
		{Key: KeyGitProtocol, Value: resolved.GitProtocol.Value, Source: resolved.GitProtocol.Source},
	}
	return entries, nil
}

type ResolvedValue struct {
	Value  string `json:"value"`
	Source Source `json:"source"`
}

type Resolved struct {
	Endpoint     ResolvedValue `json:"endpoint"`
	Organization ResolvedValue `json:"organization"`
	Project      ResolvedValue `json:"project"`
	Repo         ResolvedValue `json:"repo"`
	GitProtocol  ResolvedValue `json:"gitProtocol"`
}

func (s *Store) Resolve(explicit Values) (Resolved, error) {
	global, err := s.Load(ScopeGlobal)
	if err != nil {
		return Resolved{}, err
	}
	repo, err := s.Load(ScopeRepo)
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{
		Endpoint:     s.resolve(KeyEndpoint, explicit.Endpoint, repo.Values.Endpoint, global.Values.Endpoint, "openapi-rdc.aliyuncs.com"),
		Organization: s.resolve(KeyOrganization, explicit.Organization, repo.Values.Organization, global.Values.Organization, ""),
		Project:      s.resolve(KeyProject, explicit.Project, repo.Values.Project, global.Values.Project, ""),
		Repo:         s.resolve(KeyRepo, explicit.Repo, repo.Values.Repo, global.Values.Repo, ""),
		GitProtocol:  s.resolve(KeyGitProtocol, explicit.GitProtocol, repo.Values.GitProtocol, global.Values.GitProtocol, "https"),
	}, nil
}

func (s *Store) resolve(key Key, explicit string, repo string, global string, fallback string) ResolvedValue {
	if explicit != "" {
		return ResolvedValue{Value: explicit, Source: SourceFlag}
	}
	if envValue := strings.TrimSpace(s.envValue(key)); envValue != "" {
		return ResolvedValue{Value: envValue, Source: SourceEnv}
	}
	if repo != "" {
		return ResolvedValue{Value: repo, Source: SourceRepo}
	}
	if global != "" {
		return ResolvedValue{Value: global, Source: SourceGlobal}
	}
	return ResolvedValue{Value: fallback, Source: SourceDefault}
}

func (s *Store) envValue(key Key) string {
	names := map[Key]string{
		KeyEndpoint:     EnvEndpoint,
		KeyOrganization: EnvOrganization,
		KeyProject:      EnvProject,
		KeyRepo:         EnvRepo,
	}
	name := names[key]
	if name == "" {
		return ""
	}
	value, ok := s.env.LookupEnv(name)
	if !ok {
		return ""
	}
	return value
}

func (s *Store) pathFor(scope Scope) (string, error) {
	switch scope {
	case ScopeGlobal:
		return s.paths.Global, nil
	case ScopeRepo:
		return s.paths.Repo, nil
	default:
		return "", fmt.Errorf("unsupported config scope %q", scope)
	}
}

func readFile(path string) (File, error) {
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return File{Version: 1}, nil
	}
	if err != nil {
		return File{}, err
	}
	file := File{}
	if err := json.Unmarshal(body, &file); err != nil {
		return File{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	if file.Version == 0 {
		file.Version = 1
	}
	return file, nil
}

func writeFile(path string, file File) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	body, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(path, body, 0o600)
}

func ParseKey(value string) (Key, error) {
	key := Key(strings.TrimSpace(value))
	switch key {
	case KeyEndpoint, KeyOrganization, KeyProject, KeyRepo, KeyGitProtocol:
		return key, nil
	default:
		return "", fmt.Errorf("unsupported config key %q", value)
	}
}

func ParseScope(value string) (Scope, error) {
	scope := Scope(strings.TrimSpace(value))
	switch scope {
	case "", ScopeGlobal:
		return ScopeGlobal, nil
	case ScopeRepo:
		return ScopeRepo, nil
	default:
		return "", fmt.Errorf("unsupported config scope %q", value)
	}
}

func valueFor(values Values, key Key) string {
	switch key {
	case KeyEndpoint:
		return values.Endpoint
	case KeyOrganization:
		return values.Organization
	case KeyProject:
		return values.Project
	case KeyRepo:
		return values.Repo
	case KeyGitProtocol:
		return values.GitProtocol
	default:
		return ""
	}
}

func setValue(values *Values, key Key, value string) {
	switch key {
	case KeyEndpoint:
		values.Endpoint = value
	case KeyOrganization:
		values.Organization = value
	case KeyProject:
		values.Project = value
	case KeyRepo:
		values.Repo = value
	case KeyGitProtocol:
		values.GitProtocol = value
	}
}
