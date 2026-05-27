//go:build linux

package bpf

import (
	"errors"
	"log"
	"time"

	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"

	"networkdoctor-agent/internal/model"
)

// GC removes stale entries from maps where the kernel side cannot safely know
// that a flow/query has timed out.
func (r *Runtime) GC(dnsTimeout, udpIdle time.Duration) model.GCStats {
	now, err := monotonicNowNS()
	if err != nil {
		log.Printf("map gc: read monotonic time: %v", err)
		return model.GCStats{}
	}

	return model.GCStats{
		DNSPendingDeleted: gcDNSPending(r.objs.NdDnsPending, now, uint64(dnsTimeout/time.Nanosecond)),
		UDPFlowsDeleted:   gcUDPFlows(r.objs.NdUdpFlows, now, uint64(udpIdle/time.Nanosecond)),
	}
}

// monotonicNowNS returns CLOCK_MONOTONIC in nanoseconds so userspace timeout
// checks use the same time base as bpf_ktime_get_ns().
func monotonicNowNS() (uint64, error) {
	var ts unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_MONOTONIC, &ts); err != nil {
		return 0, err
	}
	return uint64(ts.Sec)*uint64(time.Second) + uint64(ts.Nsec), nil
}

// gcDNSPending deletes DNS query entries that never received a matching
// response before the configured timeout.
func gcDNSPending(m *ebpf.Map, nowNS, timeoutNS uint64) int {
	var key bpfNdDnsKey
	var val bpfNdDnsPendingValue
	deleted := 0

	iter := m.Iterate()
	for iter.Next(&key, &val) {
		if val.StartNs == 0 || nowNS <= val.StartNs || nowNS-val.StartNs <= timeoutNS {
			continue
		}
		if err := m.Delete(&key); err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
			log.Printf("map gc: delete nd_dns_pending: %v", err)
			continue
		}
		deleted++
	}
	if err := iter.Err(); err != nil {
		log.Printf("map gc: iterate nd_dns_pending: %v", err)
	}
	return deleted
}

// gcUDPFlows deletes UDP flow state after it has been idle long enough. This
// keeps the LRU map from retaining old flows until pressure evicts them.
func gcUDPFlows(m *ebpf.Map, nowNS, idleNS uint64) int {
	var key bpfNdFlow4Key
	var val bpfNdUdpFlowValue
	deleted := 0

	iter := m.Iterate()
	for iter.Next(&key, &val) {
		if val.LastSeenNs == 0 || nowNS <= val.LastSeenNs || nowNS-val.LastSeenNs <= idleNS {
			continue
		}
		if err := m.Delete(&key); err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
			log.Printf("map gc: delete nd_udp_flows: %v", err)
			continue
		}
		deleted++
	}
	if err := iter.Err(); err != nil {
		log.Printf("map gc: iterate nd_udp_flows: %v", err)
	}
	return deleted
}
