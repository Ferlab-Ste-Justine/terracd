package source

type SourceType int64

const (
	TypeUndefined SourceType = iota
	TypeGitRepo 
	TypeDirectory
	TypeBackendHttp
)

type RepoAuth struct {
	SshKeyPath     string `yaml:"ssh_key_path"`
	KnownHostsPath string `yaml:"known_hosts_path"`
}

type Repo struct {
	Url                string
	Ref                string
	Path               string
	Auth               RepoAuth
	GpgPublicKeysPaths []string `yaml:"gpg_public_keys_paths"`
}

type Source struct {
	Dir         string
	Repo        Repo
	BackendHttp BackendHttp     `yaml:"backend_http"`
}

func (src *Source) GetType() SourceType {
	if src.Dir != "" {
		return TypeDirectory
	}
	if src.Repo.Url != "" {
		return TypeGitRepo
	}
	if src.BackendHttp.Filename != "" {
		return TypeBackendHttp
	}

	return TypeUndefined
}

type Sources []Source