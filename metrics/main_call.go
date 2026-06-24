package metrics

import (
	"time"
)

func PushMetrics(conf MetricsClientConfig, cmd string, result string, providers []Provider, now time.Time) error {
	if !conf.IsDefined() {
		return nil
	}

	cli, cliErr := conf.GetClient()
	if cliErr != nil {
		return cliErr
	}

	err := cli.Initialize()
	if err != nil {
		return err
	}

	return cli.Push(cmd, result, providers, now)
}