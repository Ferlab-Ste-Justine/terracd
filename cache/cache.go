package cache

import(
	"path"

	"github.com/Ferlab-Ste-Justine/terracd/fs"
)


type CacheInfo struct {
	VersionsHash string `yaml:"versions_hash"`
}

func GetCacheInfo(workDir string, conf CacheConfig) (*CacheInfo, error) {
	hash, hashErr := fs.GetFileSha256(path.Join(workDir, conf.VersionsFile))
	if hashErr != nil {
		return nil, hashErr
	}
	return &CacheInfo{hash}, nil
}

func (info *CacheInfo) ShouldUse(otherInfo *CacheInfo) bool {
	return info.VersionsHash == otherInfo.VersionsHash && info.VersionsHash != ""
}

type CacheConfig struct {
	VersionsFile string `yaml:"versions_file"`
}

func (conf *CacheConfig) IsDefined() bool {
	return conf.VersionsFile != ""
}