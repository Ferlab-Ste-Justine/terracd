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
	JobName     string                   `yaml:"job_name"`
	PushGateway MetricsPushGatewayConfig `yaml:"pushgateway"`
}

func (conf *MetricsClientConfig) IsDefined() bool {
	return conf.JobName != ""
}

type MetricsClient struct {
	Config    MetricsClientConfig
	pusher    *push.Pusher
	timestamp prometheus.Gauge
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

	cli.timestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "terracd_timestamp_seconds",
		Help: "Timestamp of completion for terracd command in seconds since epoch.",
	})
	
	cli.pusher = cli.pusher.Collector(cli.timestamp)

	return nil
}

func (cli *MetricsClient) Push(cmd string, result string, now time.Time) error {
	cli.timestamp.Set(float64(now.Unix()))
	return cli.pusher.Grouping("command", cmd).Grouping("result", result).Push()
}

func PushMetrics(conf MetricsClientConfig, cmd string, result string, now time.Time) error {
	if !conf.IsDefined() {
		return nil
	}

	cli := MetricsClient{Config: conf}

	err := cli.Initialize()
	if err != nil {
		return err
	}

	return cli.Push(cmd, result, now)
}