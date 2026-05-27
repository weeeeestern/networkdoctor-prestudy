//go:build linux

package bpf

import (
	"log"
	"runtime"

	"github.com/cilium/ebpf"

	"networkdoctor-agent/internal/model"
)

const (
	tcpEstablished = 1
	tcpTimeWait    = 6
)

// Snapshot reads the latest BPF map values and converts generated BPF structs
// into package-neutral model types.
func (r *Runtime) Snapshot() model.Snapshot {
	dns := sumDNSCounters(r.objs.NdDnsCounters)
	for rcode := uint32(0); rcode < 16; rcode++ {
		dns.RcodeTotal[rcode] = sumU64PerCPU(r.objs.NdDnsRcodeTotal, rcode)
	}

	return model.Snapshot{
		TCP:      sumTCPCounters(r.objs.NdTcpCounters),
		DNS:      dns,
		UDP:      sumUDPCounters(r.objs.NdUdpCounters),
		TCPState: sumTCPStateMetrics(r.objs.NdTcpStateDelta),
		Hist: model.Histograms{
			TCPSRTTUs:     sumHistogram(r.objs.NdTcpSrttUsHist),
			TCPCwnd:       sumHistogram(r.objs.NdTcpCwndHist),
			DNSLatencyUs:  sumHistogram(r.objs.NdDnsLatencyUsHist),
			RunqlatUs:     sumHistogram(r.objs.NdRunqlatUsHist),
			UDPDurationUs: sumHistogram(r.objs.NdUdpDurationUsHist),
		},
	}
}

func sumTCPStateMetrics(m *ebpf.Map) model.TCPStateMetrics {
	return model.TCPStateMetrics{
		Established: sumTCPState(m, tcpEstablished),
		TimeWait:    sumTCPState(m, tcpTimeWait),
	}
}

// sumTCPState reads the per-CPU state delta for a single TCP state and returns
// the sum across CPUs.
func sumTCPState(m *ebpf.Map, state uint32) int64 {
	values, err := lookupPerCPU[int64](m, state)
	if err != nil {
		log.Printf("collect nd_tcp_state_delta[%d]: %v", state, err)
		return 0
	}

	var total int64
	for _, v := range values {
		total += v
	}
	return total
}

// sumTCPCounters aggregates the per-CPU TCP counter struct into one total.
func sumTCPCounters(m *ebpf.Map) model.TCPMetrics {
	values, err := lookupPerCPU[bpfNdTcpCounters](m, 0)
	if err != nil {
		log.Printf("collect nd_tcp_counters: %v", err)
		return model.TCPMetrics{}
	}

	var total model.TCPMetrics
	for _, v := range values {
		total.ClosedTotal += v.ClosedTotal
		total.ShortLivedTotal += v.ShortLivedTotal
		total.ConnectAttemptTotal += v.ConnectAttemptTotal
		total.ConnectFailedTotal += v.ConnectFailedTotal
		total.RetransTotal += v.RetransTotal
		total.RetransBurstTotal += v.RetransBurstTotal
		total.RTTSampleTotal += v.RttSampleTotal
		total.CwndSampleTotal += v.CwndSampleTotal
	}
	return total
}

// sumDNSCounters aggregates the per-CPU DNS counter struct into one total.
func sumDNSCounters(m *ebpf.Map) model.DNSMetrics {
	values, err := lookupPerCPU[bpfNdDnsCounters](m, 0)
	if err != nil {
		log.Printf("collect nd_dns_counters: %v", err)
		return model.DNSMetrics{}
	}

	var total model.DNSMetrics
	for _, v := range values {
		total.QueryTotal += v.QueryTotal
		total.ResponseTotal += v.ResponseTotal
		total.RcodeErrorTotal += v.RcodeErrorTotal
		total.SlowTotal += v.SlowTotal
	}
	return total
}

// sumUDPCounters aggregates the per-CPU UDP counter struct into one total.
func sumUDPCounters(m *ebpf.Map) model.UDPMetrics {
	values, err := lookupPerCPU[bpfNdUdpCounters](m, 0)
	if err != nil {
		log.Printf("collect nd_udp_counters: %v", err)
		return model.UDPMetrics{}
	}

	var total model.UDPMetrics
	for _, v := range values {
		total.PacketTotal += v.PacketTotal
		total.LongFlowTotal += v.LongFlowTotal
	}
	return total
}

func sumHistogram(m *ebpf.Map) [model.HistBuckets]uint64 {
	var counts [model.HistBuckets]uint64
	for i := uint32(0); i < model.HistBuckets; i++ {
		counts[i] = sumU64PerCPU(m, i)
	}
	return counts
}

// sumU64PerCPU reads a per-CPU uint64 array value and sums all CPU slots.
func sumU64PerCPU(m *ebpf.Map, key uint32) uint64 {
	values, err := lookupPerCPU[uint64](m, key)
	if err != nil {
		log.Printf("collect %s[%d]: %v", m.String(), key, err)
		return 0
	}

	var total uint64
	for _, v := range values {
		total += v
	}
	return total
}

// lookupPerCPU handles per-CPU map reads for generated BPF value types. The
// runtime.NumCPU-sized slice works on most systems; the fallback lets cilium/ebpf
// size the slice dynamically if the kernel reports a different CPU count.
func lookupPerCPU[T any](m *ebpf.Map, key uint32) ([]T, error) {
	values := make([]T, runtime.NumCPU())
	if err := m.Lookup(&key, &values); err == nil {
		return values, nil
	}

	var dynamic []T
	if err := m.Lookup(&key, &dynamic); err != nil {
		return nil, err
	}
	return dynamic, nil
}
