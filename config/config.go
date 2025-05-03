package config

import (
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/Ferlab-Ste-Justine/terracd/hook"
	"github.com/Ferlab-Ste-Justine/terracd/metrics"
	"github.com/Ferlab-Ste-Justine/terracd/cache"
	"github.com/Ferlab-Ste-Justine/terracd/recurrence"
	"github.com/Ferlab-Ste-Justine/terracd/source"
	"github.com/Ferlab-Ste-Justine/terracd/state"
)

type ConfigTimeouts struct {
	TerraformInit    time.Duration `yaml:"terraform_init"`
	TerraformPlan    time.Duration `yaml:"terraform_plan"`
	TerraformApply   time.Duration `yaml:"terraform_apply"`
	TerraformDestroy time.Duration `yaml:"terraform_destroy"`
	TerraformPull    time.Duration `yaml:"terraform_pull"`
	TerraformPush    time.Duration `yaml:"terraform_push"`
	Wait             time.Duration
}

type BackendMigration struct {
	CurrentBackend string `yaml:"current_backend"`
	NextBackend    string `yaml:"next_backend"`
}

type Config struct {
	TerraformPath    string                      `yaml:"terraform_path"`
	Sources          source.Sources
	Timeouts         ConfigTimeouts
	Recurrence       recurrence.Recurrence
	RandomJitter     time.Duration               `yaml:"random_jitter"`
	BackendMigration BackendMigration            `yaml:"backend_migration"`
	Command          string
	TerminationHooks hook.TerminationHooks       `yaml:"termination_hooks"`
	WorkingDirectory string                      `yaml:"working_directory"`
	StateStore       state.StateStoreConfig      `yaml:"state_store"`
	Cache            cache.CacheConfig                  
	Metrics          metrics.MetricsClientConfig
}

func getConfigFilePath() string {
	path := os.Getenv("TERRACD_CONFIG_FILE")
	if path == "" {
		return "config.yml"
	}
	return path
}

func GetConfig() (Config, error) {
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

	if c.Command != "apply" && c.Command != "plan" && c.Command != "destroy" && c.Command != "wait" && c.Command != "migrate_backend" {
		return c, errors.New("Valid command values can only be 'plan', 'apply', 'destroy', 'wait' or 'migrate_backend'")
	}

	for _, thook := range []hook.TerminationHook{c.TerminationHooks.Success, c.TerminationHooks.Failure, c.TerminationHooks.Always} {
		if (thook.HttpCall.Endpoint != "" && thook.HttpCall.Method == "") || (thook.HttpCall.Endpoint == "" && thook.HttpCall.Method != "") {
			return c, errors.New("If an http call is defined in a termination hook, both the method and endpoint must be defined")
		}
	}

	if c.WorkingDirectory == "" {
		wd, wdErr := os.Getwd()
		if wdErr != nil {
			return c, wdErr
		}

		c.WorkingDirectory = wd
	}

	c.WorkingDirectory, err = filepath.Abs(c.WorkingDirectory)
	if err != nil {
		return c, err
	}

	if c.Recurrence.IsDefined() && (!c.StateStore.IsDefined()) {
		return c, errors.New("If a reccurrence is defined, a state store must also be defined in order to enforce it")
	}

	if c.Cache.IsDefined() && (!c.StateStore.IsDefined()) {
		return c, errors.New("If cache is defined, a state store must also be defined in order to manage it")
	}

	for _, src := range c.Sources {
		if src.GetType() == source.TypeUndefined {
			return c, errors.New("One of the listed sources could not be properly interpreted")
		}
	}

	return c, c.Cache.Initialize()
}
