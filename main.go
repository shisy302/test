package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 3*time.Second, "连接超时时间")
	interval := flag.Duration("interval", 0, "持续监视间隔，0 表示仅检测一次")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: tcp-monitor [选项] <host:port> [host:port...]\n\n选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n  tcp-monitor localhost:22\n  tcp-monitor -interval 5s localhost:22 8.8.8.8:53\n")
	}
	flag.Parse()

	targets := flag.Args()
	if len(targets) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	check := func() {
		fmt.Printf("[%s]\n", time.Now().Format("2006-01-02 15:04:05"))
		allHealthy := true
		for _, target := range targets {
			r := checkTCP(target, *timeout)
			status := "✓ 健康"
			detail := fmt.Sprintf("延迟 %v", r.Latency.Round(time.Millisecond))
			if !r.Healthy {
				status = "✗ 异常"
				detail = r.Error
				allHealthy = false
			}
			fmt.Printf("  %-25s %s  (%s)\n", r.Address, status, detail)
		}
		fmt.Println(strings.Repeat("-", 50))
		if !allHealthy && *interval == 0 {
			os.Exit(1)
		}
	}

	if *interval == 0 {
		check()
		return
	}
	for {
		check()
		time.Sleep(*interval)
	}
}
