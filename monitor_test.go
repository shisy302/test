package main

import (
	"net"
	"testing"
	"time"
)

func TestCheckTCP_SuccessOnOpenPort(t *testing.T) {
	// 启动一个本地监听器模拟开放端口
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	result := checkTCP(addr, 2*time.Second)

	if !result.Healthy {
		t.Errorf("期望健康，得到: %s", result.Error)
	}
	if result.Latency <= 0 {
		t.Errorf("期望延迟 > 0，得到: %v", result.Latency)
	}
	if result.Address != addr {
		t.Errorf("期望地址 %s，得到 %s", addr, result.Address)
	}
	if result.Protocol != "tcp" {
		t.Errorf("期望协议 tcp，得到 %s", result.Protocol)
	}
}

func TestCheckTCP_FailOnClosedPort(t *testing.T) {
	result := checkTCP("127.0.0.1:19999", 500*time.Millisecond)

	if result.Healthy {
		t.Error("期望不健康，但返回了健康")
	}
	if result.Error == "" {
		t.Error("期望有错误信息")
	}
	if result.Protocol != "tcp" {
		t.Errorf("期望协议 tcp，得到 %s", result.Protocol)
	}
}

func TestCheckUDP_SuccessOnListeningPort(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	go func() {
		buf := make([]byte, 1024)
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		conn.WriteToUDP(buf[:n], remote)
	}()

	result := checkUDP(conn.LocalAddr().String(), 2*time.Second)

	if !result.Healthy {
		t.Errorf("期望健康，得到: %s", result.Error)
	}
	if result.Latency <= 0 {
		t.Errorf("期望延迟 > 0，得到: %v", result.Latency)
	}
	if result.Protocol != "udp" {
		t.Errorf("期望协议 udp，得到 %s", result.Protocol)
	}
}

func TestCheckUDP_FailOnNoListener(t *testing.T) {
	result := checkUDP("127.0.0.1:19999", 500*time.Millisecond)

	if result.Healthy {
		t.Error("期望不健康，但返回了健康")
	}
	if result.Protocol != "udp" {
		t.Errorf("期望协议 udp，得到 %s", result.Protocol)
	}
}

func TestCheckUDP_FailOnTimeout(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	result := checkUDP(conn.LocalAddr().String(), 500*time.Millisecond)

	if result.Healthy {
		t.Error("期望不健康，但返回了健康")
	}
	if result.Error != "no response within timeout" {
		t.Errorf("期望错误 'no response within timeout'，得到: %s", result.Error)
	}
}
