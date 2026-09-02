package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/grokbot-2/agentcontent/internal/ports"
)

type Adapter struct {
	key     string
	baseURL string
	model   string
	client  *http.Client
}

func New(key, baseURL, model string) *Adapter {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &Adapter{
		key:     strings.TrimSpace(key),
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 45 * time.Second},
	}
}

func (a *Adapter) Stub() bool { return a.key == "" }

func (a *Adapter) Complete(ctx context.Context, system, user string) (string, error) {
	if a.Stub() {
		return stubComplete(system, user), nil
	}
	body, _ := json.Marshal(map[string]any{
		"model": a.model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.7,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+a.key)
	req.Header.Set("Content-Type", "application/json")
	res, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("ai http %d: %s", res.StatusCode, raw)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("ai: empty completion")
	}
	return parsed.Choices[0].Message.Content, nil
}

func stubComplete(system, user string) string {
	u := strings.TrimSpace(user)
	if len(u) > 180 {
		u = u[:180] + "…"
	}
	return fmt.Sprintf(`# Stub AI (không có OPENAI_API_KEY)

Bạn hỏi: %s

Gợi ý (seed / domain):
- So sánh viewsPerDay với average viewsPerDay của video cùng kênh — mặc định 2.5x. Diagram *100 chỉ để seed nhìn “to”, không phải ngưỡng sản phẩm.
- Gắn ý tưởng vào 1 pillar của Brand Blueprint.
- Mode kịch bản: Nhanh (hook 15s), Auto (b-roll list), Sâu (thesis + case).

_system hint_: %s`, u, trim(system, 160))
}

func trim(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

var _ ports.AIProviderPort = (*Adapter)(nil)
