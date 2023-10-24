package metrics

import (
	"net/http"
	"time"

	//"github.com/prometheus/client_golang/prometheus/push"

	"github.com/Ferlab-Ste-Justine/terracd/auth"
)

type MetricsPushGatewayConfig struct {
	Url  string
	Auth auth.Auth
}

type MetricsClientConfig struct {
	JobName     string                   `yaml:"job_name"`
	PushGateway MetricsPushGatewayConfig `yaml:"push_gateway"`
}

type MetricsClient struct {
	Config MetricsClientConfig
	client *http.Client
}

func (cli *MetricsClient) Connect() error {
	passErr := cli.Config.PushGateway.Auth.ResolvePassword()
	if passErr != nil {
		return passErr
	}

	tls, tlsErr := cli.Config.PushGateway.Auth.GetTlsConfigs()
	if tlsErr != nil {
		return tlsErr
	}

	cli.client = &http.Client{Transport: &http.Transport{TLSClientConfig: tls}}

	return nil
}

func (cli *MetricsClient) Push(cmd string, now time.Time) error {
	return nil
}

func(cli *MetricsClient) Close() error {
	return nil
}