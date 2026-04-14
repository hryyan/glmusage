package display

import (
	"fmt"
	"math"
	"strings"

	"github.com/hryyan/glmusage/internal/api"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
)

type Config struct {
	NoColor bool
}

// Render 紧凑单行输出。
func Render(result *api.UsageResult, cfg Config) {
	c := colorFunc(cfg.NoColor)
	sep := c(Dim, " · ")

	var parts []string

	// 标题
	level := result.Level
	if level == "" {
		level = "?"
	}
	parts = append(parts, c(Bold, "GLM")+c(Cyan, " ["+level+"]"))

	// MCP
	if result.MCP.Type != "" {
		bar := progressBar(result.MCP.Percentage, 10, c)
		parts = append(parts, fmt.Sprintf("MCP %s (%.0f/%.0f)", bar, result.MCP.CurrentValue, result.MCP.Usage))
	}

	// Token 5h
	if result.Token5h.Type != "" {
		bar := progressBar(result.Token5h.Percentage, 10, c)
		parts = append(parts, fmt.Sprintf("Token5h %s", bar))
	}

	// 今日统计
	if result.Usage.TotalCalls > 0 || result.Usage.TotalTokens > 0 {
		parts = append(parts, fmt.Sprintf("今日 %s次 %s tok",
			formatNumber(result.Usage.TotalCalls),
			formatTokens(result.Usage.TotalTokens)))
	}

	fmt.Println(strings.Join(parts, sep))
}

func progressBar(pct float64, width int, c func(string, string) string) string {
	if pct > 100 {
		pct = 100
	}
	filled := int(math.Round(pct / 100 * float64(width)))
	if filled > width {
		filled = width
	}

	color := Green
	if pct > 85 {
		color = Red
	} else if pct > 60 {
		color = Yellow
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return c(color, bar) + fmt.Sprintf("%.0f%%", pct)
}

func formatTokens(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.2fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(ch))
	}
	return string(result)
}

func colorFunc(noColor bool) func(string, string) string {
	if noColor {
		return func(_, s string) string { return s }
	}
	return func(color, s string) string { return color + s + Reset }
}
