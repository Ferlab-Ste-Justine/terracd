package cache

import(
	"os"
	"path"

	"github.com/Ferlab-Ste-Justine/terracd/fs"
	"github.com/Ferlab-Ste-Justine/terracd/s3"
)

func cacheProviders(workDir string, cacheDir string) error {
	lockFileSrc := path.Join(workDir, ".terraform.lock.hcl")
	lockFileDest := path.Join(cacheDir, ".terraform.lock.hcl")
	providerDirSrc := path.Join(workDir, ".terraform", "providers")
	providerDirDest := path.Join(cacheDir, ".terraform", "providers")

	lockFileSrcExists, lockFileSrcErr := fs.PathExists(lockFileSrc)
	if lockFileSrcErr != nil {
		return lockFileSrcErr
	}

	providerDirSrcExists, providerDirSrcErr := fs.PathExists(providerDirSrc)
	if providerDirSrcErr != nil {
		return providerDirSrcErr
	}

	providerDirDestExists, providerDirDestErr := fs.PathExists(providerDirDest)
	if providerDirDestErr != nil {
		return providerDirDestErr
	}

	if lockFileSrcExists && providerDirSrcExists {
		if providerDirDestExists {
			remErr := os.RemoveAll(providerDirDest)
			if remErr != nil {
				return remErr
			}
		}
		
		copyErr := fs.CopyPrivateFile(lockFileSrc, lockFileDest)
		if copyErr != nil {
			return copyErr
		}

		ensErr := fs.AssurePrivateDir(providerDirDest)
		if ensErr != nil {
			return ensErr
		}

		copyErr = fs.CopyDir(providerDirDest, providerDirSrc)
		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}

type ProviderCacheInfo struct {
	VersionsHash string `yaml:"versions_hash"`
}

func GetProviderCacheInfo(workDir string, conf ProviderCacheConfig) (ProviderCacheInfo, error) {
	hash, hashErr := fs.GetFileSha256(path.Join(workDir, conf.VersionsFile))
	if hashErr != nil {
		return ProviderCacheInfo{}, hashErr
	}

	return ProviderCacheInfo{hash}, nil
}

func (info *ProviderCacheInfo) ShouldUse(otherInfo *ProviderCacheInfo) bool {
	return info.VersionsHash == otherInfo.VersionsHash && info.VersionsHash != ""
}

type ProviderCacheConfig struct {
	VersionsFile string            `yaml:"versions_file"`
	S3           s3.S3ClientConfig `yaml:"s3"`
}

func (conf *ProviderCacheConfig) Initialize() error {
	if conf.IsDefined() && conf.S3.IsDefined() {
		return conf.S3.Auth.GetKeyAuth()
	}

	return nil
}

func (conf *ProviderCacheConfig) IsDefined() bool {
	return conf.VersionsFile != ""
}

func (conf *ProviderCacheConfig) Load(workDir string, cacheDir string, stateCache ProviderCacheInfo) (ProviderCacheInfo, fs.DirInfo, error) {	
	if !conf.IsDefined() {
		return ProviderCacheInfo{}, fs.DirInfo{}, nil
	}

	var cacheInfo ProviderCacheInfo
	var cacheInfoErr error	
	cacheInfo, cacheInfoErr = GetProviderCacheInfo(workDir, *conf)
	if cacheInfoErr != nil {
		return ProviderCacheInfo{}, fs.DirInfo{}, cacheInfoErr
	}

	if !cacheInfo.ShouldUse(&stateCache) {
		return cacheInfo, fs.GetEmptyDirInfo(cacheDir), nil
	}

	if conf.S3.IsDefined() {
		syncErr := s3.SyncToFs(conf.S3, cacheDir)
		if syncErr != nil {
			return cacheInfo, fs.GetEmptyDirInfo(cacheDir), syncErr
		}
	}

	mergeErr := fs.MergeDirs(workDir, []string{cacheDir})
	if mergeErr != nil {
		return cacheInfo, fs.GetEmptyDirInfo(cacheDir), mergeErr
	}

	dirInfo, dirInfoErr := fs.GetDirInfo(cacheDir)
	if dirInfoErr != nil {
		return cacheInfo, fs.GetEmptyDirInfo(cacheDir), dirInfoErr
	}

	return cacheInfo, dirInfo, nil
}

func (conf *ProviderCacheConfig) Save(workDir string, cacheDir string, prevCacheInfo fs.DirInfo) error {
	if !conf.IsDefined() {
		return nil
	}

	cacheErr := cacheProviders(workDir, cacheDir)
	if cacheErr != nil {
		return cacheErr
	}

	if conf.S3.IsDefined() {
		dirInfo, dirInfoErr := fs.GetDirInfo(cacheDir)
		if dirInfoErr != nil {
			return dirInfoErr
		}

		if prevCacheInfo.Differs(&dirInfo) {
			return s3.SyncFromFs(conf.S3, cacheDir)
		}
	}

	return nil
}