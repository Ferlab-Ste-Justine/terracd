package main

import (
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

type ConfigSourceRepoAuth struct {
	SshKeyPath     string `yaml:"ssh_key_path"`
	KnownHostsPath string `yaml:"known_hosts_path"`
}

type ConfigSourceRepo struct {
	Url                string
	Ref                string
	Path               string
	Auth               ConfigSourceRepoAuth
	GpgPublicKeysPaths []string `yaml:"gpg_public_keys_paths"`
}

type ConfigSource struct {
	Dir  string
	Repo ConfigSourceRepo
}

type ConfigTimeouts struct {
	TerraformInit  time.Duration `yaml:"terraform_init"`
	TerraformPlan  time.Duration `yaml:"terraform_plan"`
	TerraformApply time.Duration `yaml:"terraform_apply"`
	TerraformPull  time.Duration `yaml:"terraform_pull"`
	TerraformPush  time.Duration `yaml:"terraform_push"`
	Wait           time.Duration
}

type BackendMigration struct {
	CurrentBackend string `yaml:"current_backend"`
	NextBackend    string `yaml:"next_backend"`
}

type Config struct {
	TerraformPath    string `yaml:"terraform_path"`
	Sources          []ConfigSource
	Timeouts         ConfigTimeouts
	BackendMigration BackendMigration `yaml:"backend_migration"`
	Command          string
	TerminationHooks TerminationHooks `yaml:"termination_hooks"`
}

func getConfigFilePath() string {
	path := os.Getenv("TERRACD_CONFIG_FILE")
	if path == "" {
		return "config.yml"
	}
	return path
}

func getConfig() (Config, error) {
	var c Config

	b, err := ioutil.ReadFile(getConfigFilePath())
	if err != nil {
		return c, errors.New(fmt.Sprintf("Error reading the configuration file: %s", err.Error()))
	}
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return c, errors.New(fmt.Sprintf("Error parsing the configuration file: %s", err.Error()))
	}

	if c.Command == "" {
		c.Command = "apply"
	}

	if c.Command != "apply" && c.Command != "plan" && c.Command != "wait" && c.Command != "migrate_backend" {
		return c, errors.New("Valid command values can only be 'plan', 'apply', 'wait' or 'migrate_backend'")
	}

	for _, hook := range []TerminationHook{c.TerminationHooks.Success, c.TerminationHooks.Failure, c.TerminationHooks.Always} {
		if (hook.HttpCall.Endpoint != "" && hook.HttpCall.Method == "") || (hook.HttpCall.Endpoint == "" && hook.HttpCall.Method != "") {
			return c, errors.New("If an http call is defined in a termination hook, both the method and endpoint must be defined")
		}
	}

	return c, nil
}
