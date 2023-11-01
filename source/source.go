package source

import (
	"path"
)

type SourceType int64

const (
	TypeUndefined SourceType = iota
	TypeGitRepo 
	TypeDirectory
	TypeBackendHttp
)

func (srcType SourceType) ToString() string {
	switch srcType {
	case TypeGitRepo:
		return "GitRepo"
	case TypeDirectory:
		return "Directory"
	case TypeBackendHttp:
		return "BackendHttp"
	default:
		return "undefined"
	}
}

type Source struct {
	Dir         string
	GitRepo     GitRepo     `yaml:"repo"`
	BackendHttp BackendHttp `yaml:"backend_http"`
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

func (srcs *Sources) GetFsPaths(reposDir string) []string {
	paths := []string{}
	for _, src:= range *srcs {
		srcType := src.GetType()
		if srcType == TypeGitRepo {
			dir := src.GitRepo.GetDir()
			dir = path.Join(reposDir, dir)
			if src.GitRepo.Path != "" {
				dir = path.Join(dir, src.GitRepo.Path)
			}
			paths = append(paths, dir)
		} else if srcType == TypeDirectory {
			paths = append(paths, src.Dir)
		}
	}

	return paths
}