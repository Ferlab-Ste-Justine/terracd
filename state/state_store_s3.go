package state

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	minio "github.com/minio/minio-go/v7"
	yaml "gopkg.in/yaml.v2"

	"github.com/Ferlab-Ste-Justine/terracd/s3"
)

type S3StateStore struct {
	Config s3.S3ClientConfig
}

func (store *S3StateStore) Initialize() error {
	return store.Config.Auth.GetKeyAuth()
}

func (store *S3StateStore) Read() (State, error) {
	var st State

	conn, connErr := s3.Connect(store.Config)
	if connErr != nil {
		return st, connErr
	}

	exists, existsErr := s3.KeyExists(store.Config.Bucket, path.Join(store.Config.Path, "state.yml"), conn)
	if existsErr != nil {
		return st, existsErr
	}

	if !exists {
		return st, nil
	}

	objRead, readErr := conn.GetObject(context.Background(), store.Config.Bucket, path.Join(store.Config.Path, "state.yml"), minio.GetObjectOptions{})
	if readErr != nil {
		return st, readErr
	}

    data, transfErr := io.ReadAll(objRead)
    if transfErr != nil {
        return st, transfErr
    }

	unmarErr := yaml.Unmarshal(data, &st)
	if unmarErr != nil {
		return st, errors.New(fmt.Sprintf("Error deserializing the state info: %s", unmarErr.Error()))
	}

	return st, nil
}

func (store *S3StateStore) Write(state State) error {
	output, err := yaml.Marshal(&state)
	if err != nil {
		return errors.New(fmt.Sprintf("Error serializing the state info: %s", err.Error()))
	}

	conn, connErr := s3.Connect(store.Config)
	if connErr != nil {
		return connErr
	}

	_, putErr := conn.PutObject(
		context.Background(),
		store.Config.Bucket,
		path.Join(store.Config.Path, "state.yml"),
		bytes.NewReader(output),
		int64(len(output)),
		minio.PutObjectOptions{},
	)

	return putErr
}

func (store *S3StateStore) Cleanup() error {
	return nil
}