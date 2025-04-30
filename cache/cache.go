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

type CacheInfo struct {
	VersionsHash string `yaml:"versions_hash"`
}

func GetCacheInfo(workDir string, conf CacheConfig) (CacheInfo, error) {
	hash, hashErr := fs.GetFileSha256(path.Join(workDir, conf.VersionsFile))
	if hashErr != nil {
		return CacheInfo{}, hashErr
	}

	return CacheInfo{hash}, nil
}

func (info *CacheInfo) ShouldUse(otherInfo *CacheInfo) bool {
	return info.VersionsHash == otherInfo.VersionsHash && info.VersionsHash != ""
}

type CacheConfig struct {
	VersionsFile string            `yaml:"versions_file"`
	S3           s3.S3ClientConfig `yaml:"s3"`
}

func (conf *CacheConfig) IsDefined() bool {
	return conf.VersionsFile != ""
}

func (conf *CacheConfig) Load(workDir string, cacheDir string, stateCache CacheInfo) (CacheInfo, fs.DirInfo, error) {	
	if !conf.IsDefined() {
		return CacheInfo{}, fs.DirInfo{}, nil
	}

	var cacheInfo CacheInfo
	var cacheInfoErr error	
	cacheInfo, cacheInfoErr = GetCacheInfo(workDir, *conf)
	if cacheInfoErr != nil {
		return CacheInfo{}, fs.DirInfo{}, cacheInfoErr
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

func (conf *CacheConfig) Save(workDir string, cacheDir string, prevCacheInfo fs.DirInfo) error {
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