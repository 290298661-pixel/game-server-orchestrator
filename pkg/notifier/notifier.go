package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Level string

const (
	LevelInfo     Level = "INFO"
	LevelWarning  Level = "WARNING"
	LevelCritical Level = "CRITICAL"
)

type ScalingEvent struct {
	Timestamp       time.Time
	Fleet           string
	Namespace       string
	Decision        string
	CurrentReplicas int32
	DesiredReplicas int32
	Reason          string
	Nodes           []string
}

type Notifier struct {
	feishuURL  string
	dingtalkURL string
	httpClient *http.Client
}

func NewNotifier(feishuURL, dingtalkURL string) *Notifier {
	return &Notifier{
		feishuURL:   feishuURL,
		dingtalkURL: dingtalkURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// NotifyScalingEvent pushes a scaling event to configured IM channels.
func (n *Notifier) NotifyScalingEvent(ctx context.Context, event ScalingEvent) {
	title := "🎮 Game Fleet Scaling Event"
	level := LevelInfo
	emoji := "📊"

	switch event.Decision {
	case "SCALE_UP":
		emoji = "📈"
		level = LevelInfo
	case "SCALE_DOWN":
		emoji = "📉"
		level = LevelInfo
	case "FROZEN":
		emoji = "🚫"
		level = LevelWarning
	}

	msg := RenderScalingMessage(title, event, emoji, level)

	if n.feishuURL != "" {
		if err := n.sendFeishu(ctx, msg); err != nil {
			log.Printf("[ERROR] [notifier] feishu send failed: %v", err)
		}
	}

	if n.dingtalkURL != "" {
		if err := n.sendDingtalk(ctx, msg); err != nil {
			log.Printf("[ERROR] [notifier] dingtalk send failed: %v", err)
		}
	}
}

func (n *Notifier) sendFeishu(ctx context.Context, msg *ScalingMessage) error {
	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]string{"content": msg.Title, "tag": "plain_text"},
			},
			"elements": []map[string]interface{}{
				{"tag": "div", "text": map[string]string{"content": msg.Body, "tag": "lark_md"}},
			},
		},
	}
	return n.post(ctx, n.feishuURL, payload)
}

func (n *Notifier) sendDingtalk(ctx context.Context, msg *ScalingMessage) error {
	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": msg.Title,
			"text":  msg.Body,
		},
	}
	return n.post(ctx, n.dingtalkURL, payload)
}

func (n *Notifier) post(ctx context.Context, url string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("post to webhook: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
