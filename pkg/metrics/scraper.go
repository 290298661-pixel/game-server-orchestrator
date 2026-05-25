package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GameServerMetrics struct {
	Players      int32   `json:"players"`
	CPUPercent   float64 `json:"cpuPercent"`
	MemoryMB     float64 `json:"memoryMB"`
	AllocRate    float64 `json:"allocRate"`
}

type Scraper struct {
	client  *http.Client
	metricsPath string
}

func NewScraper(timeout time.Duration) *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: timeout,
		},
		metricsPath: "/api/v1/metrics",
	}
}

// Scrape fetches metrics from a single game server.
func (s *Scraper) Scrape(ctx context.Context, endpoint string) (*GameServerMetrics, error) {
	url := fmt.Sprintf("http://%s%s", endpoint, s.metricsPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build scrape request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("scrape %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scrape %s: status %d", endpoint, resp.StatusCode)
	}

	var m GameServerMetrics
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("decode metrics from %s: %w", endpoint, err)
	}

	return &m, nil
}

// ScrapeFleet collects metrics from all endpoints and aggregates them.
func (s *Scraper) ScrapeFleet(ctx context.Context, endpoints []string) (*FleetMetricsAggregate, error) {
	agg := &FleetMetricsAggregate{}

	for _, ep := range endpoints {
		m, err := s.Scrape(ctx, ep)
		if err != nil {
			agg.Errors = append(agg.Errors, fmt.Sprintf("%s: %v", ep, err))
			continue
		}
		agg.TotalPlayers += m.Players
		agg.TotalCPU += m.CPUPercent
		agg.TotalMemoryMB += m.MemoryMB
		agg.ServerCount++
	}

	if agg.ServerCount > 0 {
		agg.AvgCPU = agg.TotalCPU / float64(agg.ServerCount)
		agg.AvgMemoryMB = agg.TotalMemoryMB / float64(agg.ServerCount)
	}

	return agg, nil
}

type FleetMetricsAggregate struct {
	TotalPlayers  int32
	TotalCPU      float64
	TotalMemoryMB float64
	AvgCPU        float64
	AvgMemoryMB   float64
	ServerCount   int32
	Errors        []string
}
