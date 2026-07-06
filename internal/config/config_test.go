package config

import (
	"path/filepath"
	"testing"
)

type fakeEnv map[string]string

func (f fakeEnv) LookupEnv(key string) (string, bool) {
	value, ok := f[key]
	return value, ok
}

func TestResolvePriority(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(Paths{
		Global: filepath.Join(dir, "global.json"),
		Repo:   filepath.Join(dir, "repo.json"),
	}, fakeEnv{
		EnvEndpoint:     "env-endpoint",
		EnvOrganization: "env-org",
		EnvRepo:         "env-repo",
	})
	if err := store.Save(ScopeGlobal, File{Values: Values{Endpoint: "global-endpoint", Organization: "global-org", Repo: "global-repo", GitProtocol: "ssh"}}); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(ScopeRepo, File{Values: Values{Endpoint: "repo-endpoint", Organization: "repo-org", Repo: "repo-repo"}}); err != nil {
		t.Fatal(err)
	}

	resolved, err := store.Resolve(Values{Endpoint: "flag-endpoint"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Endpoint.Value != "flag-endpoint" || resolved.Endpoint.Source != SourceFlag {
		t.Fatalf("endpoint = %#v", resolved.Endpoint)
	}
	if resolved.Organization.Value != "env-org" || resolved.Organization.Source != SourceEnv {
		t.Fatalf("organization = %#v", resolved.Organization)
	}
	if resolved.Repo.Value != "env-repo" || resolved.Repo.Source != SourceEnv {
		t.Fatalf("repo = %#v", resolved.Repo)
	}
	if resolved.GitProtocol.Value != "ssh" || resolved.GitProtocol.Source != SourceGlobal {
		t.Fatalf("git protocol = %#v", resolved.GitProtocol)
	}
}
