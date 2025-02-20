package s3

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"
	yaml "gopkg.in/yaml.v2"
)

type S3KeyAuth struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

type S3AuthConfig struct {
	CaCert    string `yaml:"ca_cert"`
	KeyAuth   string `yaml:"key_auth"`
	AccessKey string `yaml:"-"`
	SecretKey string `yaml:"-"`
}

func (conf *S3AuthConfig) GetKeyAuth() error {
	var a S3KeyAuth

	b, err := ioutil.ReadFile(conf.KeyAuth)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading the s3 key auth file at path '%s': %s", conf.KeyAuth, err.Error()))
	}

	err = yaml.Unmarshal(b, &a)
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing the s3 key auth file: %s", err.Error()))
	}

	conf.AccessKey = a.AccessKey
	conf.SecretKey = a.SecretKey

	return nil
}

type S3ClientConfig struct {
	ObjectsPrefix     string        `yaml:"objects_prefix"`
	Endpoint          string
	Bucket            string
	Path              string
	Region            string
	Auth              S3AuthConfig
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
	RequestTimeout    time.Duration `yaml:"request_timeout"`
}

func (conf *S3ClientConfig) IsDefined() bool {
	return conf.Endpoint != ""
}