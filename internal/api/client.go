package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://open.bigmodel.cn"
const defaultTimeout = 10 * time.Second

// Config 是 API 客户端配置。
type Config struct {
	BaseURL   string
	AuthToken string
	Timeout   time.Duration
}

// ModelUsage 表示模型用量统计数据。
type ModelUsage struct {
	Times       []string `json:"x_time"`
	CallCounts  []int    `json:"modelCallCount"`
	TotalCalls  int      `json:"totalModelCallCount"`
	TotalTokens int      `json:"totalTokensUsage"`
}

// QuotaLimit 表示单项配额限制。
type QuotaLimit struct {
	Type          string  `json:"type"`
	Unit          int     `json:"unit"`
	Number        int     `json:"number"`
	Usage         float64 `json:"usage"`         // TIME_LIMIT: 总额度
	CurrentValue  float64 `json:"currentValue"`  // TIME_LIMIT: 已用量
	Remaining     float64 `json:"remaining"`     // TIME_LIMIT: 剩余
	Percentage    float64 `json:"percentage"`    // 使用百分比
	NextResetTime int64   `json:"nextResetTime"` // 下次重置时间戳(ms)
}

// UsageResult 是聚合后的完整查询结果。
type UsageResult struct {
	Platform  string
	Level     string
	Timestamp time.Time
	Usage     ModelUsage
	MCP       QuotaLimit // TIME_LIMIT
	Token5h   QuotaLimit // TOKENS_LIMIT
}

// Client 是 GLM API 客户端。
type Client struct {
	config Config
	http   *http.Client
}

// NewClient 创建 API 客户端。
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	return &Client{
		config: cfg,
		http:   &http.Client{Timeout: cfg.Timeout},
	}
}

// FetchAll 并行获取所有数据并返回聚合结果。
func (c *Client) FetchAll(ctx context.Context, startTime, endTime string) (*UsageResult, error) {
	type result struct {
		usage  *modelUsageResp
		limits *quotaResp
		err    error
	}

	ch := make(chan result, 2)

	go func() {
		u, err := c.fetchModelUsage(ctx, startTime, endTime)
		ch <- result{usage: u, err: err}
	}()
	go func() {
		q, err := c.fetchQuotaLimits(ctx)
		ch <- result{limits: q, err: err}
	}()

	var usageResp *modelUsageResp
	var quotaRespData *quotaResp
	for i := 0; i < 2; i++ {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		if r.usage != nil {
			usageResp = r.usage
		}
		if r.limits != nil {
			quotaRespData = r.limits
		}
	}

	res := &UsageResult{
		Platform:  "智谱AI",
		Timestamp: time.Now(),
	}

	if usageResp != nil {
		res.Usage = usageResp.Data
	}

	if quotaRespData != nil {
		res.Level = quotaRespData.Data.Level
		for _, l := range quotaRespData.Data.Limits {
			switch l.Type {
			case "TIME_LIMIT":
				res.MCP = l
			case "TOKENS_LIMIT":
				res.Token5h = l
			}
		}
	}

	return res, nil
}

func (c *Client) fetchModelUsage(ctx context.Context, startTime, endTime string) (*modelUsageResp, error) {
	u, _ := url.Parse(c.config.BaseURL)
	u.Path = "/api/monitor/usage/model-usage"
	q := u.Query()
	q.Set("startTime", startTime)
	q.Set("endTime", endTime)
	u.RawQuery = q.Encode()

	var resp modelUsageResp
	if err := c.get(ctx, u.String(), &resp); err != nil {
		return nil, fmt.Errorf("获取模型用量失败: %w", err)
	}
	return &resp, nil
}

func (c *Client) fetchQuotaLimits(ctx context.Context) (*quotaResp, error) {
	u, _ := url.Parse(c.config.BaseURL)
	u.Path = "/api/monitor/usage/quota/limit"

	var resp quotaResp
	if err := c.get(ctx, u.String(), &resp); err != nil {
		return nil, fmt.Errorf("获取配额限制失败: %w", err)
	}
	return &resp, nil
}

func (c *Client) get(ctx context.Context, rawURL string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.config.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("JSON 解析失败: %w (body: %s)", err, truncate(string(body), 200))
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// --- JSON 响应类型 ---

type modelUsageResp struct {
	Code int        `json:"code"`
	Data ModelUsage `json:"data"`
	Msg  string     `json:"msg"`
}

type quotaResp struct {
	Code int       `json:"code"`
	Data quotaData `json:"data"`
	Msg  string    `json:"msg"`
}

type quotaData struct {
	Limits []QuotaLimit `json:"limits"`
	Level  string       `json:"level"`
}
