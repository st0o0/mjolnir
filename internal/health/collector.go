package health

import (
	"context"
	"log"
	"time"

	"github.com/st0o0/mjolnir/internal/nut"
)

type Collector struct {
	querier  nut.Querier
	writer   *MetricWriter
	upsNames []string
	interval time.Duration
}

func NewCollector(querier nut.Querier, writer *MetricWriter, upsNames []string, interval time.Duration) *Collector {
	return &Collector{
		querier:  querier,
		writer:   writer,
		upsNames: upsNames,
		interval: interval,
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
				vars, err := c.querier.ListVars(name)
				if err != nil {
					log.Printf("[mjolnir] failed to query UPS %s: %v", name, err)
					c.writer.SetError(name)
					continue
				}
				c.writer.Write(name, vars)
			}
		}
	}
}
