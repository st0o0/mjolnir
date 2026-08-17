package health

import (
	"context"
	"log"
	"time"

	"github.com/st0o0/mjolnir/internal/nut"
)

type Collector struct {
	upsNames []string
	interval time.Duration
	stop     chan struct{}
}

func NewCollector(upsNames []string, interval time.Duration) *Collector {
	return &Collector{
		upsNames: upsNames,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (c *Collector) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, name := range c.upsNames {
				vars, err := nut.Query(name)
				if err != nil {
					log.Printf("[mjolnir] failed to query UPS %s: %v", name, err)
					continue
				}
				UpdateMetrics(name, vars)
			}
		}
	}
}
