package state

import (
	"errors"

	"github.com/Ferlab-Ste-Justine/terracd/s3"
)

type StateStore interface {
	Initialize() error
	Read() (State, error)
	Write(State) error
	Cleanup() error
}

type StateStoreConfig struct {
	Fs   FsConfig
	Etcd EtcdConfig
	S3   s3.S3ClientConfig
}

func (conf *StateStoreConfig) IsDefined() bool {
	return conf.Fs.IsDefined() || conf.Etcd.IsDefined()
}

func (conf *StateStoreConfig) GetStore(fsStorePath string) (StateStore, error) {
	if conf.Fs.IsDefined() {
		return &FsStateStore{
			Config: FsConfig{
				Enabled: true,
				Path: fsStorePath,
			},
		}, nil
	} else if conf.Etcd.IsDefined() {
		return &EtcdStateStore{Config: conf.Etcd}, nil
	}

	return &FsStateStore{}, errors.New("Tried to create a store though no valid definition was found")
}