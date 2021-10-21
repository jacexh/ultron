package ultron

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/wosai/ultron/v2/pkg/statistics"
	"golang.org/x/sync/errgroup"
)

type (
	StatsProvider interface {
		ID() string
		Start()
		Stop()
		Submit(ctx context.Context, batch uint32) (*statistics.StatisticianGroup, error)
	}

	StatsAggregator interface {
		Aggregate(bool) statistics.SummaryReport
		Add(...StatsProvider)
		Remove(string)
	}

	Slave interface {
		GetStatsProvider() StatsProvider
	}

	statsAggregator struct {
		counter    uint32
		providers  map[string]StatsProvider
		aggTimeout time.Duration
		buffer     map[uint32]map[string]*statistics.StatisticianGroup
		mu         sync.RWMutex
	}
)

func newStatsAggregator() *statsAggregator {
	return &statsAggregator{
		providers:  make(map[string]StatsProvider),
		aggTimeout: 3 * time.Second,
	}
}

func (agg *statsAggregator) Aggregate(fullHistory bool) (statistics.SummaryReport, error) {
	agg.mu.Lock()
	var providers []StatsProvider
	batch := agg.counter
	agg.counter++

	agg.buffer[batch] = make(map[string]*statistics.StatisticianGroup)
	for _, provider := range agg.providers {
		providers = append(providers, provider)
		agg.buffer[batch][provider.ID()] = nil
	}
	agg.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), agg.aggTimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	for _, provider := range providers {
		provider := provider // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			sg, err := provider.Submit(ctx, batch)
			if err != nil {
				return err
			}
			agg.mu.Lock()
			defer agg.mu.Unlock()
			agg.buffer[batch][provider.ID()] = sg
			return nil
		})
	}
	g.Wait()

	// 检查是否完成
	agg.mu.Lock()
	defer agg.mu.Unlock()

	returns := agg.buffer[batch]
	delete(agg.buffer, batch)
	sg := statistics.NewStatisticianGroup()
	for id, ret := range returns {
		if ret == nil {
			log.Printf("after %d seconds, provider %s did not submit StatisticianGroup yet, drop this batch: %d\n", agg.aggTimeout/time.Second, id, batch) // TODO: 之后补充处理细节
			return statistics.SummaryReport{}, fmt.Errorf("[%s] lost some SubStatisticianGroup in batch %d", id, batch)
		} else {
			sg.Merge(ret)
		}
	}
	return sg.Report(fullHistory), nil
}

func (agg *statsAggregator) Remove(id string) {
	agg.mu.Lock()
	defer agg.mu.Unlock()

	delete(agg.providers, id)
}

func (agg *statsAggregator) Add(providers ...StatsProvider) {
	agg.mu.Lock()
	defer agg.mu.Unlock()

	for _, provider := range providers {
		agg.providers[provider.ID()] = provider
	}
}
