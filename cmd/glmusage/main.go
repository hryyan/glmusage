package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hryyan/glmusage/internal/api"
	"github.com/hryyan/glmusage/internal/display"
	"github.com/hryyan/glmusage/internal/timeutil"
	"github.com/hryyan/glmusage/internal/upgrade"
)

// 构建时通过 -ldflags "-X main.version=v0.1.0" 注入
var version = "dev"

var (
	upgradeCmd   = flag.Bool("upgrade", false, "升级到最新版本")
	versionCmd   = flag.Bool("version", false, "显示版本号")
	watchFlag    = flag.Bool("watch", false, "持续监控模式")
	intervalFlag = flag.Int("interval", 60, "watch 模式刷新间隔（秒）")
	noColorFlag  = flag.Bool("no-color", false, "禁用彩色输出")
)

func main() {
	flag.Parse()

	switch {
	case *versionCmd:
		fmt.Printf("glmusage %s\n", version)
		return
	case *upgradeCmd:
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := upgrade.DoUpgrade(ctx, version); err != nil {
			fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m\n", err)
			os.Exit(1)
		}
		return
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
		runWatch(client)
	} else {
		runOnce(client)
	}
}

func runOnce(client *api.Client) {
	result, err := fetchUsage(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m\n", err)
		os.Exit(1)
	}
	display.Render(result, display.Config{NoColor: *noColorFlag})
}

func runWatch(client *api.Client) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(*intervalFlag) * time.Second)
	defer ticker.Stop()

	refresh(client)

	for {
		select {
		case <-ticker.C:
			refresh(client)
		case <-sigCh:
			fmt.Println("\n已退出监控模式。")
			return
		}
	}
}

func refresh(client *api.Client) {
	fmt.Print("\033[2J\033[H")

	result, err := fetchUsage(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m\n", err)
		return
	}

	display.Render(result, display.Config{NoColor: *noColorFlag})
	fmt.Printf("\033[2m每 %d 秒刷新，按 Ctrl+C 退出\033[0m\n", *intervalFlag)
}

func fetchUsage(client *api.Client) (*api.UsageResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	start, end := timeutil.Last24Hours()
	return client.FetchAll(ctx, timeutil.FormatForAPI(start), timeutil.FormatForAPI(end))
}
