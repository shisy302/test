package main

import (
	"net"
	"time"
)

type CheckResult struct {
	Address string
	Healthy bool
	Latency time.Duration
	Error   string
}

func checkTCP(address string, timeout time.Duration) CheckResult {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Address: address,
			Healthy: false,
			Latency: latency,
			Error:   err.Error(),
		}
	}
	conn.Close()
	return CheckResult{
		Address: address,
		Healthy: true,
		Latency: latency,
	}
}
