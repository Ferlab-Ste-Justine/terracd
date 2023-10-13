package state

import (
	"errors"
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
}

func (conf *StateStoreConfig) IsDefined() bool {
	return conf.Fs.IsDefined() || conf.Etcd.IsDefined()
}

func (conf *StateStoreConfig) GetStore() (StateStore, error) {
	if conf.Fs.IsDefined() {
		return &FsStateStore{Config: conf.Fs}, nil
	} else if conf.Etcd.IsDefined() {
		return &EtcdStateStore{Config: conf.Etcd}, nil
	}

	return &FsStateStore{}, errors.New("Tried to create a store though no valid definition was found")
}