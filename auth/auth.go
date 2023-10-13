package auth

import (
	"errors"
	"fmt"
	"io/ioutil"
	yaml "gopkg.in/yaml.v2"
)

type PasswordAuth struct {
	Username string
	Password string
}

type Auth struct {
	CaCert       string `yaml:"ca_cert"`
	ClientCert   string `yaml:"client_cert"`
	ClientKey    string `yaml:"client_key"`
	PasswordAuth string `yaml:"password_auth"`
	Username     string `yaml:"-"`
	Password     string `yaml:"-"`
}

func (auth *Auth) ResolvePassword() error {
	var a PasswordAuth

	if auth.ClientCert != "" {
		return nil
	}

	b, err := ioutil.ReadFile(auth.PasswordAuth)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading the password auth file: %s", err.Error()))
	}

	err = yaml.Unmarshal(b, &a)
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing the password auth file: %s", err.Error()))
	}

	auth.Username = a.Username
	auth.Password = a.Password

	return nil
}