package metrics

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

type PromRemoteWrite struct {
	Config    *MetricsClientConfig
	client    *http.Client
}

func (cli *PromRemoteWrite) Initialize() error {
	passErr := cli.Config.Collector.Auth.ResolvePassword()
	if passErr != nil {
		return passErr
	}

	tls, tlsErr := cli.Config.Collector.Auth.GetTlsConfigs()
	if tlsErr != nil {
		return tlsErr
	}

	cli.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tls,
		},
	}

	return nil
}

func (cli *PromRemoteWrite) Push(cmd string, result string, providers []Provider, now time.Time) error {
	currentTime := now.Unix()
	
	promTS := []prompb.TimeSeries{}

	promTS = append(promTS, prompb.TimeSeries{
		Labels: []prompb.Label{
			prompb.Label{Name: "command", Value: cmd},
			prompb.Label{Name: "result", Value: result},
		}, 
		Samples: []prompb.Sample{prompb.Sample{Timestamp: now.UnixMilli(), Value: float64(currentTime)}},
	})

	for _, provider := range providers {
		promTS = append(promTS, prompb.TimeSeries{
			Labels: []prompb.Label{
				prompb.Label{Name: "registry", Value: provider.Registry},
				prompb.Label{Name: "organisation", Value: provider.Organization},
				prompb.Label{Name: "provider", Value: provider.Name},
				prompb.Label{Name: "version", Value: provider.Version},
			}, 
			Samples: []prompb.Sample{prompb.Sample{Timestamp: now.UnixMilli(), Value: float64(currentTime)}},
		})
	} 

	body, marErr := proto.Marshal(&prompb.WriteRequest{
		Timeseries: promTS,
	})
	if marErr != nil {
		return marErr
	}

	compressedBody := snappy.Encode(nil, body)

	req, reqErr := http.NewRequest("POST", cli.Config.Collector.Url, bytes.NewBuffer(compressedBody))
	if reqErr != nil {
		return reqErr
	}

	req.Header.Add("X-Prometheus-Remote-Write-Version", "0.1.0")
	req.Header.Add("Content-Encoding", "snappy")
	req.Header.Set("Content-Type", "application/x-protobuf")

	if cli.Config.Collector.Auth.HasPassword() {
		req.SetBasicAuth(cli.Config.Collector.Auth.Username, cli.Config.Collector.Auth.Password)
	}

	resp, respErr := cli.client.Do(req)
	if respErr != nil {
		return respErr
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 300 {
		respBody, respBodyErr := io.ReadAll(resp.Body)
		if respBodyErr != nil {
			return respBodyErr
		}

		return fmt.Errorf("Prometheus remote write returned status %d and message: %s", resp.StatusCode, respBody)
	}

	return nil
}