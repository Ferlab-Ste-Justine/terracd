package metrics

import (
	"time"
)

type MetricsClient interface {
    Initialize() error
	Push(cmd string, result string, providers []Provider, now time.Time) error
}