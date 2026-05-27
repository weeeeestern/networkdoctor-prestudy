package config

import (
	"flag"
	"time"
)

// Config contains runtime knobs exposed as CLI flags. Most threshold values
// are written into the BPF nd_config map so the kernel program can apply them.
type Config struct {
	Iface       string
	Listen      string
	MetricsPath string

	GCInterval time.Duration
	DNSTimeout time.Duration
	UDPIdle    time.Duration

	SlowTCPRTT    time.Duration
	ShortTCPConn  time.Duration
	SlowDNS       time.Duration
	LongUDPFlow   time.Duration
	HighRunqlat   time.Duration
	RetransWindow time.Duration
	EventCooldown time.Duration

	RetransBurst uint64
	CwndDropPct  uint64
}

// ParseFlags parses CLI flags using the original flag names and defaults.
func ParseFlags() Config {
	var cfg Config
	flag.StringVar(&cfg.Iface, "iface", "eth0", "interface to attach TC ingress/egress programs to")
	flag.StringVar(&cfg.Listen, "listen", ":9102", "HTTP listen address for Prometheus metrics")
	flag.StringVar(&cfg.MetricsPath, "metrics-path", "/metrics", "Prometheus metrics path")
	flag.DurationVar(&cfg.GCInterval, "gc-interval", 10*time.Second, "BPF map garbage collection interval")
	flag.DurationVar(&cfg.DNSTimeout, "dns-timeout", 5*time.Second, "delete DNS pending entries older than this")
	flag.DurationVar(&cfg.UDPIdle, "udp-idle-timeout", 60*time.Second, "delete UDP flow entries idle for longer than this")
	flag.DurationVar(&cfg.SlowTCPRTT, "slow-tcp-rtt", 100*time.Millisecond, "slow TCP RTT threshold written to nd_config")
	flag.DurationVar(&cfg.ShortTCPConn, "short-tcp-conn", 100*time.Millisecond, "short TCP connection threshold written to nd_config")
	flag.DurationVar(&cfg.SlowDNS, "slow-dns", 50*time.Millisecond, "slow DNS threshold written to nd_config")
	flag.DurationVar(&cfg.LongUDPFlow, "long-udp-flow", 30*time.Second, "long UDP flow threshold written to nd_config")
	flag.DurationVar(&cfg.HighRunqlat, "high-runqlat", 5*time.Millisecond, "high run queue latency threshold written to nd_config")
	flag.DurationVar(&cfg.RetransWindow, "retrans-window", time.Second, "TCP retransmit burst window written to nd_config")
	flag.DurationVar(&cfg.EventCooldown, "event-cooldown", 5*time.Second, "anomaly event cooldown written to nd_config")
	flag.Uint64Var(&cfg.RetransBurst, "retrans-burst", 3, "TCP retransmit burst count threshold written to nd_config")
	flag.Uint64Var(&cfg.CwndDropPct, "cwnd-drop-percent", 50, "TCP cwnd drop percentage threshold written to nd_config")
	flag.Parse()
	return cfg
}
