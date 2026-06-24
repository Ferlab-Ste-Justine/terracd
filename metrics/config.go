package metrics

import (
	"errors"

	"github.com/Ferlab-Ste-Justine/terracd/auth"
)

type MetricsCollectorConfig struct {
	Url  string
	Auth auth.Auth
	Type string
}

func (conf *MetricsCollectorConfig) GetType() (string, error) {
	if conf.Type == "" {
		return "prom_pushgateway", nil
	}

	if conf.Type != "prom_pushgateway" && conf.Type != "prom_remote_write" {
		return "", errors.New("Metrics collector configuration support only the following types: prom_pushgateway, prom_remote_write")
	}

	return conf.Type, nil
}


type MetricsClientConfig struct {
	JobName          string                   `yaml:"job_name"`
	IncludeProviders bool                     `yaml:"include_providers"`
	Collector        MetricsCollectorConfig   `yaml:"collector"`
}

func (conf *MetricsClientConfig) IsDefined() bool {
	return conf.JobName != ""
}

func (conf *MetricsClientConfig) GetClient() (MetricsClient, error) {
	cType, cTypeErr := conf.Collector.GetType()
	if cTypeErr != nil {
		return nil, cTypeErr
	}

	if cType == "prom_pushgateway" {
		return &PromPushGateway{Config: conf}, nil
	} else {
		return &PromRemoteWrite{Config: conf}, nil
	}
}