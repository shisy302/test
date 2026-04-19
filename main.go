package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type targetResult struct {
	URL        string
	Up         bool
	StatusCode int
	Latency    time.Duration
	Err        string
}

func parseFlags() ([]string, time.Duration, time.Duration) {
	var targets stringslice
	interval := flag.Duration("interval", 10*time.Second, "probe interval (e.g. 10s, 1m)")
	timeout := flag.Duration("timeout", 5*time.Second, "per-request timeout (e.g. 5s, 10s)")
	flag.Var(&targets, "target", "health check target URL (repeatable, required)")
	flag.Parse()
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "error: at least one -target is required")
		flag.Usage()
		os.Exit(1)
	}
	return []string(targets), *interval, *timeout
}

type stringslice []string

func (s *stringslice) String() string   { return strings.Join(*s, ", ") }
func (s *stringslice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func checkTarget(ctx context.Context, url string, timeout time.Duration) targetResult {
	start := time.Now()
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return targetResult{URL: url, Err: err.Error(), Latency: time.Since(start)}
	}
	resp, err := client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return targetResult{URL: url, Err: err.Error(), Latency: latency}
	}
	defer resp.Body.Close()
	up := resp.StatusCode >= 200 && resp.StatusCode < 300
	return targetResult{URL: url, Up: up, StatusCode: resp.StatusCode, Latency: latency}
}

func worker(ctx context.Context, url string, interval, timeout time.Duration, ch chan<- targetResult) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ch <- checkTarget(ctx, url, timeout)
		}
	}
}

func printer(ch <-chan targetResult) {
	const layout = "2006-01-02 15:04:05"
	for r := range ch {
		ts := time.Now().Format(layout)
		if r.Up {
			fmt.Printf("[%s] %s  UP    %d  %s\n", ts, r.URL, r.StatusCode, r.Latency.Round(time.Millisecond))
		} else {
			status := "-"
			if r.StatusCode != 0 {
				status = fmt.Sprintf("%d", r.StatusCode)
			}
			fmt.Printf("[%s] %s  DOWN  %s  %s\n", ts, r.URL, status, r.Err)
		}
	}
}

func main() {
	targets, interval, timeout := parseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ch := make(chan targetResult, len(targets))

	go printer(ch)

	var wg sync.WaitGroup
	for _, t := range targets {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			worker(ctx, url, interval, timeout, ch)
		}(t)
	}

	<-ctx.Done()
	fmt.Println("\nshutting down...")
	stop()

	wg.Wait()
	close(ch)
}
