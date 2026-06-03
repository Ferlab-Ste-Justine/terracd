package cache

import(
	"github.com/Ferlab-Ste-Justine/terracd/fs"
	"github.com/Ferlab-Ste-Justine/terracd/s3"
)

type GitSourcesCacheConfig struct {
	Fs bool      
	S3 s3.S3ClientConfig
	Clear bool
}

func (conf *GitSourcesCacheConfig) IsDefined() bool {
	return conf.Fs || conf.S3.IsDefined()
}

func (conf *GitSourcesCacheConfig) Initialize() error {
	if conf.IsDefined() && conf.S3.IsDefined() {
		return conf.S3.Auth.GetKeyAuth()
	}

	return nil
}


func (conf *GitSourcesCacheConfig) Load(gitReposDir string, cacheDir string) error {	
	if !conf.IsDefined() || conf.Clear {
		return nil
	}

	ensErr := fs.AssurePrivateDir(cacheDir)
	if ensErr != nil {
		return ensErr
	}

	if conf.S3.IsDefined() {
		syncErr := s3.SyncToFs(conf.S3, cacheDir)
		if syncErr != nil {
			return syncErr
		}
	}

	copyErr := fs.CopyDir(gitReposDir, cacheDir)
	if copyErr != nil {
		return copyErr
	}

	return nil
}

func (conf *GitSourcesCacheConfig) Save(gitReposDir string, cacheDir string) error {
	if !conf.IsDefined() {
		return nil
	}

	copyErr := fs.SyncDir(cacheDir, gitReposDir)
	if copyErr != nil {
		return copyErr
	}

	if conf.S3.IsDefined() {
		return s3.SyncFromFs(conf.S3, cacheDir)
	}

	return nil
}