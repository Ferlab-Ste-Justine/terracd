package metrics

import (
	"github.com/Ferlab-Ste-Justine/terracd/auth"
)

type PrometheusPushgatewayConfig struct {
	Url  string
	Auth auth.Auth
}

func (conf *PrometheusPushgatewayConfig) IsDefined() bool {
	return conf.Url != ""
}

type PrometheusRemoteWriteConfig struct {
	Url  string
	Auth auth.Auth
}

func (conf *PrometheusRemoteWriteConfig) IsDefined() bool {
	return conf.Url != ""
}

type MetricsCollectorConfig struct {
	PrometheusPushgateway PrometheusPushgatewayConfig `yaml:"prometheus_pushgateway"`
	PrometheusRemoteWrite PrometheusRemoteWriteConfig `yaml:"prometheus_remote_write"`
}

func (conf *MetricsCollectorConfig) IsDefined() bool {
	return conf.PrometheusPushgateway.IsDefined() || conf.PrometheusRemoteWrite.IsDefined()
}

type MetricsClientConfig struct {
	JobName          string                   `yaml:"job_name"`
	IncludeProviders bool                     `yaml:"include_providers"`
	Collector        MetricsCollectorConfig   `yaml:"collector"`
}

type MetricsClientBaseConfig struct {
	JobName          string                   `yaml:"job_name"`
	IncludeProviders bool                     `yaml:"include_providers"`
}

func (conf *MetricsClientConfig) GetBaseConfig() MetricsClientBaseConfig {
	return MetricsClientBaseConfig{
		JobName: conf.JobName,
		IncludeProviders: conf.IncludeProviders,
	}
}

func (conf *MetricsClientConfig) IsDefined() bool {
	return conf.JobName != "" && conf.Collector.IsDefined()
}

func (conf *MetricsClientConfig) GetClient() (MetricsClient, error) {
	if conf.Collector.PrometheusPushgateway.IsDefined() {
		return &PromPushGateway{BaseConfig: conf.GetBaseConfig(), CollectorConfig: conf.Collector.PrometheusPushgateway}, nil
	} else {
		return &PromRemoteWrite{BaseConfig: conf.GetBaseConfig(), CollectorConfig: conf.Collector.PrometheusRemoteWrite}, nil
	}
}