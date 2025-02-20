package state

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
	yaml "gopkg.in/yaml.v2"

	"github.com/Ferlab-Ste-Justine/terracd/auth"
)

type EtcdConfig struct {
	Prefix            string
	Endpoints         []string
	ConnectionTimeout time.Duration	`yaml:"connection_timeout"`
	RequestTimeout    time.Duration `yaml:"request_timeout"`
	RetryInterval     time.Duration `yaml:"retry_interval"`
	Retries           uint64
	Auth              auth.Auth
}

func (conf *EtcdConfig) IsDefined() bool {
	return len(conf.Endpoints) > 0
}

type EtcdStateStore struct {
	Config EtcdConfig
	client *client.EtcdClient
}

func (store *EtcdStateStore) Initialize() error {
	passErr := store.Config.Auth.ResolvePassword()
	if passErr != nil {
		return passErr
	}

	cli, cliErr := client.Connect(context.Background(), client.EtcdClientOptions{
		ClientCertPath:    store.Config.Auth.ClientCert,
		ClientKeyPath:     store.Config.Auth.ClientKey,
		CaCertPath:        store.Config.Auth.CaCert,
		Username:          store.Config.Auth.Username,
		Password:          store.Config.Auth.Password,
		EtcdEndpoints:     store.Config.Endpoints,
		ConnectionTimeout: store.Config.ConnectionTimeout,
		RequestTimeout:    store.Config.RequestTimeout,
		RetryInterval:     store.Config.RetryInterval,
		Retries:           store.Config.Retries,
	})

	store.client = cli
	return cliErr
}

func (store *EtcdStateStore) Read() (State, error) {
	var st State

	keyInfo, err := store.client.GetKey(fmt.Sprintf("%s%s", store.Config.Prefix, "state.yml"), client.GetKeyOptions{})
	if err != nil {
		return st, errors.New(fmt.Sprintf("Error retrieving state info: %s", err.Error()))
	}

	if !keyInfo.Found() {
		return st, nil
	}

	err = yaml.Unmarshal([]byte(keyInfo.Value), &st)
	if err != nil {
		return st, errors.New(fmt.Sprintf("Error deserializing the state info: %s", err.Error()))
	}

	return st, nil
}

func (store *EtcdStateStore) Write(state State) error {
	output, err := yaml.Marshal(&state)
	if err != nil {
		return errors.New(fmt.Sprintf("Error serializing the state info: %s", err.Error()))
	}

	_, err = store.client.PutKey(fmt.Sprintf("%s%s", store.Config.Prefix, "recurrence.yml"), string(output))
	if err != nil {
		return errors.New(fmt.Sprintf("Error writing the state file: %s", err.Error()))
	}

	return nil
}

func (store *EtcdStateStore) Cleanup() error {
	store.client.Close()
	return nil
}