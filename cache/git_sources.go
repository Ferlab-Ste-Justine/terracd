package cache

import(
	"github.com/Ferlab-Ste-Justine/terracd/s3"
)

type GitSourcesCacheConfig struct { 
	S3 s3.S3ClientConfig
}

func (conf *GitSourcesCacheConfig) IsDefined() bool {
	return conf.S3.IsDefined()
}

func (conf *GitSourcesCacheConfig) Initialize() error {
	if conf.IsDefined() {
		return conf.S3.Auth.GetKeyAuth()
	}

	return nil
}


func (conf *GitSourcesCacheConfig) Load(gitReposDir string) error {	
	if !conf.IsDefined() {
		return nil
	}

	if conf.S3.IsDefined() {
		syncErr := s3.SyncToFs(conf.S3, gitReposDir)
		if syncErr != nil {
			return syncErr
		}
	}

	return nil
}

func (conf *GitSourcesCacheConfig) Save(gitReposDir string) error {
	if !conf.IsDefined() {
		return nil
	}

	if conf.S3.IsDefined() {
		return s3.SyncFromFs(conf.S3, gitReposDir)
	}

	return nil
}