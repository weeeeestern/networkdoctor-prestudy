package bpf

// bpf2go compiles the kernel-side eBPF C program and generates Go wrappers
// such as bpfObjects, loadBpfObjects, and bpfNdEvent used by this package.
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -no-strip -target amd64 -type nd_config -type nd_event -type nd_tcp_counters -type nd_dns_counters -type nd_udp_counters -type nd_flow4_key -type nd_dns_key -type nd_dns_pending_value -type nd_udp_flow_value bpf ../../bpf/networkdoctor.bpf.c -- -I../../bpf -I../..
