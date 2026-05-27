//go:build linux

package worker

import (
	"context"
	"log"
	"time"

	"networkdoctor-agent/internal/bpf"
)

// RunGCWorker periodically removes stale entries from BPF maps where userspace
// owns timeout/idle expiration.
func RunGCWorker(ctx context.Context, rt *bpf.Runtime, interval, dnsTimeout, udpIdle time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := rt.GC(dnsTimeout, udpIdle)
			if stats.DNSPendingDeleted > 0 || stats.UDPFlowsDeleted > 0 {
				log.Printf("map gc deleted dns_pending=%d udp_flows=%d", stats.DNSPendingDeleted, stats.UDPFlowsDeleted)
			}
		}
	}
}
