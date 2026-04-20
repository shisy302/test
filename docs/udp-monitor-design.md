# UDP 服务监控设计

## 概述

为现有 TCP 端口健康检测工具添加 UDP 服务监控能力，采用双协议并行检测方案：对每个目标同时执行 TCP 拨号检测和 UDP 探测，输出两行结果。

## 需求

- 每个 `host:port` 目标同时检测 TCP 和 UDP 两种协议
- UDP 探测方式：发送空探测包 + 等待任意响应
- 收到任意响应判定为健康，超时无响应判定为异常
- 输出格式清晰区分协议类型

## 数据结构变更

`CheckResult` 新增 `Protocol` 字段：

```go
type CheckResult struct {
    Address  string
    Protocol string // "tcp" 或 "udp"
    Healthy  bool
    Latency  time.Duration
    Error    string
}
```

## UDP 探测逻辑

新增 `checkUDP` 函数（monitor.go）：

1. 使用 `net.DialTimeout("udp", address, timeout)` 建立连接
2. 发送 1 字节探测包（`0x00`）
3. 设置读取超时，等待响应
4. 收到任意响应 → 健康
5. 超时或 ICMP Port Unreachable → 异常，错误信息 "no response within timeout"

## 主流程变更（main.go）

- `check` 闭包内对每个目标先调 `checkTCP`，再调 `checkUDP`
- 地址显示格式：`host:port/TCP` 和 `host:port/UDP`
- `allHealthy`：TCP 和 UDP 任一异常即为异常

## 输出格式

```
[2026-04-20 10:00:00]
  8.8.8.8:53/TCP  ✓ 健康  (延迟 12ms)
  8.8.8.8:53/UDP  ✓ 健康  (延迟 15ms)
  localhost:22/TCP  ✓ 健康  (延迟 1ms)
  localhost:22/UDP  ✗ 异常  (no response within timeout)
--------------------------------------------------
```

## 测试覆盖

- `TestCheckUDP_SuccessOnListeningPort`：启动 UDP 监听器，回复后验证健康
- `TestCheckUDP_FailOnNoListener`：对无监听端口检测，验证异常
- `TestCheckUDP_FailOnTimeout`：启动 UDP 监听器但不回复，验证超时异常
- 更新现有 TCP 测试以适配 `Protocol` 字段

## 边界情况

- `-timeout` 同时用于 TCP 和 UDP
- Linux 上 ICMP Port Unreachable 会被 `ReadFrom` 捕获为错误，正确处理为异常
- UDP 无连接特性导致超时是主要异常路径

## 文件变更清单

| 文件 | 变更 |
|------|------|
| monitor.go | `CheckResult` 加 `Protocol` 字段；`checkTCP` 设置 Protocol；新增 `checkUDP` |
| main.go | 输出格式加协议标识；每个目标检测 TCP+UDP |
| monitor_test.go | 更新现有测试；新增 3 个 UDP 测试 |
