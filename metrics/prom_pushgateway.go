package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type PromPushGateway struct {
	Config    *MetricsClientConfig
	pusher    *push.Pusher
}

func (cli *PromPushGateway) Initialize() error {
	cli.pusher = push.New(cli.Config.Collector.Url, cli.Config.JobName)

	passErr := cli.Config.Collector.Auth.ResolvePassword()
	if passErr != nil {
		return passErr
	}

	tls, tlsErr := cli.Config.Collector.Auth.GetTlsConfigs()
	if tlsErr != nil {
		return tlsErr
	}

	cli.pusher = cli.pusher.Client(&http.Client{Transport: &http.Transport{TLSClientConfig: tls}})

	if cli.Config.Collector.Auth.HasPassword() {
		cli.pusher = cli.pusher.BasicAuth(
			cli.Config.Collector.Auth.Username,
			cli.Config.Collector.Auth.Password,
		)
	}

	return nil
}

func (cli *PromPushGateway) Push(cmd string, result string, providers []Provider, now time.Time) error {
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