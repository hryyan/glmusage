package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hryyan/glmusage/internal/api"
	"github.com/hryyan/glmusage/internal/display"
	"github.com/hryyan/glmusage/internal/timeutil"
	"github.com/hryyan/glmusage/internal/upgrade"
)

var version = "dev"

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "用法: glmusage [命令] [选项]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "命令:")
		fmt.Fprintln(os.Stderr, "  (无)       查询用量")
		fmt.Fprintln(os.Stderr, "  upgrade    升级到最新版本")
		fmt.Fprintln(os.Stderr, "  version    显示版本号")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "选项:")
		fmt.Fprintln(os.Stderr, "  -watch         持续监控模式")
		fmt.Fprintln(os.Stderr, "  -interval 60   监控刷新间隔（秒）")
		fmt.Fprintln(os.Stderr, "  -no-color      禁用彩色输出")
	}

	watchFlag := flag.Bool("watch", false, "持续监控模式")
	intervalFlag := flag.Int("interval", 60, "watch 模式刷新间隔（秒）")
	noColorFlag := flag.Bool("no-color", false, "禁用彩色输出")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "upgrade":
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			if err := upgrade.DoUpgrade(ctx, version); err != nil {
				fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m\n", err)
				os.Exit(1)
			}
			return
		case "version", "-v", "--version":
			fmt.Printf("glmusage %s\n", version)
			return
		default:
			fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[0])
			flag.Usage()
			os.Exit(1)
		}
	}

	token := os.Getenv("GLM_AUTH_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "\033[31m错误: 请设置环境变量 GLM_AUTH_TOKEN\033[0m")
		fmt.Fprintln(os.Stderr, "  export GLM_AUTH_TOKEN=\"your-token-here\"")
		os.Exit(1)
	}

	client := api.NewClient(api.Config{
		AuthToken: token,
	})

	if *watchFlag {
		runWatch(client, *intervalFlag, *noColorFlag)
	} else {
		runOnce(client, *noColorFlag)
	}
}

func runOnce(client *api.Client, noColor bool) {
	result, err := fetchUsage(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m\n", err)
		os.Exit(1)
	}
	display.Render(result, display.Config{NoColor: noColor})
}

func runWatch(client *api.Client, interval int, noColor bool) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	refresh(client, interval, noColor)

	for {
		select {
		case <-ticker.C:
			refresh(client, interval, noColor)
		case <-sigCh:
			fmt.Println("\n已退出监控模式。")
			return
		}
	}
}

func refresh(client *api.Client, interval int, noColor bool) {
	fmt.Print("\033[2J\033[H")

	result, err := fetchUsage(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m\n", err)
		return
	}

	if result.Token5h.Type != "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			_ = os.WriteFile(filepath.Join(home, ".glm_cost"), []byte(fmt.Sprintf("%.0f%%\n", result.Token5h.Percentage)), 0644)
		}
	}

	display.Render(result, display.Config{NoColor: noColor})
	fmt.Printf("\033[2m每 %d 秒刷新，按 Ctrl+C 退出\033[0m\n", interval)
}

func fetchUsage(client *api.Client) (*api.UsageResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	start, end := timeutil.Last24Hours()
	return client.FetchAll(ctx, timeutil.FormatForAPI(start), timeutil.FormatForAPI(end))
}
