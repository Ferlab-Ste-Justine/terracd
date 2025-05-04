package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/Ferlab-Ste-Justine/terracd/auth"
)

type MetricsPushGatewayConfig struct {
	Url  string
	Auth auth.Auth
}

type MetricsClientConfig struct {
	JobName          string                   `yaml:"job_name"`
	IncludeProviders bool                     `yaml:"include_providers"`
	PushGateway      MetricsPushGatewayConfig `yaml:"pushgateway"`
}

func (conf *MetricsClientConfig) IsDefined() bool {
	return conf.JobName != ""
}

type MetricsClient struct {
	Config    MetricsClientConfig
	pusher    *push.Pusher
}

func (cli *MetricsClient) Initialize() error {
	cli.pusher = push.New(cli.Config.PushGateway.Url, cli.Config.JobName)

	passErr := cli.Config.PushGateway.Auth.ResolvePassword()
	if passErr != nil {
		return passErr
	}

	tls, tlsErr := cli.Config.PushGateway.Auth.GetTlsConfigs()
	if tlsErr != nil {
		return tlsErr
	}

	cli.pusher = cli.pusher.Client(&http.Client{Transport: &http.Transport{TLSClientConfig: tls}})

	if cli.Config.PushGateway.Auth.HasPassword() {
		cli.pusher = cli.pusher.BasicAuth(
			cli.Config.PushGateway.Auth.Username,
			cli.Config.PushGateway.Auth.Password,
		)
	}

	return nil
}

func (cli *MetricsClient) Push(cmd string, result string, providers []Provider, now time.Time) error {
	currentTime := now.Unix()

	cmdTimestamp := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "terracd_command_timestamp_seconds",
		Help: "Timestamp of completion for terracd command in seconds since epoch.",
		ConstLabels: prometheus.Labels{"command": cmd, "result": result},
	})
	cmdTimestamp.Set(float64(currentTime))
	cli.pusher = cli.pusher.Collector(cmdTimestamp)

	for _, provider := range providers {
		providerUseTimestamp := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "terracd_provider_use_timestamp_seconds",
			Help: "Timestamp when terracd used a specific terraform provider in seconds since epoch.",
			ConstLabels: prometheus.Labels{"registry": provider.Registry, "organisation": provider.Organization, "provider": provider.Name, "version": provider.Version},
		})
		providerUseTimestamp.Set(float64(currentTime))
		cli.pusher = cli.pusher.Collector(providerUseTimestamp)
	
	} 
	
	return cli.pusher.Push()
}

func PushMetrics(conf MetricsClientConfig, cmd string, result string, providers []Provider, now time.Time) error {
	if !conf.IsDefined() {
		return nil
	}

	cli := MetricsClient{Config: conf}

	err := cli.Initialize()
	if err != nil {
		return err
	}

	return cli.Push(cmd, result, providers, now)
}