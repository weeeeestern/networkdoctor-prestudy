//go:build linux

package bpf

import "github.com/cilium/ebpf"

func writeConfig(objs *bpfObjects, opts Options) error {
	key := uint32(0)
	cfg := bpfNdConfig{
		SlowTcpRttUs:         opts.SlowTCPRTTUs,
		ShortTcpConnUs:       opts.ShortTCPConnUs,
		SlowDnsUs:            opts.SlowDNSUs,
		LongUdpFlowUs:        opts.LongUDPFlowUs,
		HighRunqlatUs:        opts.HighRunqlatUs,
		RetransBurstWindowNs: opts.RetransBurstWindowNs,
		EventCooldownNs:      opts.EventCooldownNs,
		RetransBurstCount:    opts.RetransBurstCount,
		CwndDropPercent:      opts.CwndDropPercent,
	}
	return objs.NdConfig.Update(&key, &cfg, ebpf.UpdateAny)
}
