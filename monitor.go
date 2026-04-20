package main

import (
	"net"
	"time"
)

type CheckResult struct {
	Address  string
	Protocol string
	Healthy  bool
	Latency  time.Duration
	Error    string
}

func checkTCP(address string, timeout time.Duration) CheckResult {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Address:  address,
			Protocol: "tcp",
			Healthy:  false,
			Latency:  latency,
			Error:    err.Error(),
		}
	}
	conn.Close()
	return CheckResult{
		Address:  address,
		Protocol: "tcp",
		Healthy:  true,
		Latency:  latency,
	}
}

func checkUDP(address string, timeout time.Duration) CheckResult {
	start := time.Now()
	conn, err := net.DialTimeout("udp", address, timeout)
	if err != nil {
		return CheckResult{
			Address:  address,
			Protocol: "udp",
			Healthy:  false,
			Latency:  time.Since(start),
			Error:    err.Error(),
		}
	}
	defer conn.Close()

	_, err = conn.Write([]byte{0x00})
	if err != nil {
		return CheckResult{
			Address:  address,
			Protocol: "udp",
			Healthy:  false,
			Latency:  time.Since(start),
			Error:    err.Error(),
		}
	}

	conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Address:  address,
			Protocol: "udp",
			Healthy:  false,
			Latency:  latency,
			Error:    "no response within timeout",
		}
	}
	return CheckResult{
		Address:  address,
		Protocol: "udp",
		Healthy:  true,
		Latency:  latency,
	}
}
