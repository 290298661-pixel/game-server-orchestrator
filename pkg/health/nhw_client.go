package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HealthProvider interface {
	GetNodeHealth(ctx context.Context) (map[string]NodeHealth, error)
}

type NodeHealth struct {
	Status string `json:"status"`
	Score  int    `json:"score"`
	Reason string `json:"reason,omitempty"`
}

// NHWClient queries the Node Health Watcher API.
type NHWClient struct {
	endpoint   string
	httpClient *http.Client
}

func NewNHWClient(endpoint string, timeout time.Duration) *NHWClient {
	return &NHWClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type nhwResponse struct {
	Nodes map[string]NodeHealth `json:"nodes"`
}

func (c *NHWClient) GetNodeHealth(ctx context.Context) (map[string]NodeHealth, error) {
	url := fmt.Sprintf("%s/api/v1/node-health", c.endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build NHW request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query NHW: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NHW returned status %d", resp.StatusCode)
	}

	var result nhwResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode NHW response: %w", err)
	}

	return result.Nodes, nil
}

var _ HealthProvider = (*NHWClient)(nil)
