//go:build linux

package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"networkdoctor-agent/internal/bpf"
	"networkdoctor-agent/internal/config"
	"networkdoctor-agent/internal/metrics"
	"networkdoctor-agent/internal/worker"
)

// Run wires the agent together: load BPF programs/maps, attach hooks, start
// background readers, and expose BPF map contents as Prometheus metrics.
func Run(cfg config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rt, err := bpf.Start(bpf.Options{
		Iface: cfg.Iface,

		SlowTCPRTTUs:         uint64(cfg.SlowTCPRTT / time.Microsecond),
		ShortTCPConnUs:       uint64(cfg.ShortTCPConn / time.Microsecond),
		SlowDNSUs:            uint64(cfg.SlowDNS / time.Microsecond),
		LongUDPFlowUs:        uint64(cfg.LongUDPFlow / time.Microsecond),
		HighRunqlatUs:        uint64(cfg.HighRunqlat / time.Microsecond),
		RetransBurstWindowNs: uint64(cfg.RetransWindow / time.Nanosecond),
		EventCooldownNs:      uint64(cfg.EventCooldown / time.Nanosecond),

		RetransBurstCount: cfg.RetransBurst,
		CwndDropPercent:   cfg.CwndDropPct,
	})
	if err != nil {
		return err
	}
	defer rt.Close()

	go worker.RunEventWorker(ctx, rt)
	go worker.RunGCWorker(ctx, rt, cfg.GCInterval, cfg.DNSTimeout, cfg.UDPIdle)

	return metrics.Serve(ctx, cfg.Listen, cfg.MetricsPath, rt)
}
