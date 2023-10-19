package source

type SourceType int64

const (
	TypeUndefined SourceType = iota
	TypeGitRepo 
	TypeDirectory
	TypeBackendHttp
)

type Source struct {
	Dir         string
	GitRepo        GitRepo
	BackendHttp BackendHttp     `yaml:"backend_http"`
}

func (src *Source) GetType() SourceType {
	if src.Dir != "" {
		return TypeDirectory
	}
	if src.GitRepo.Url != "" {
		return TypeGitRepo
	}
	if src.BackendHttp.Filename != "" {
		return TypeBackendHttp
	}

	return TypeUndefined
}

type Sources []Source