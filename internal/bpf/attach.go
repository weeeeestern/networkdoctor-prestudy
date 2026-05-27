//go:build linux

package bpf

import (
	"fmt"
	"io"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// attachPrograms attaches each generated BPF program to its kernel hook. The
// hook types mirror the SEC(...) declarations in networkdoctor.bpf.c.
func attachPrograms(objs *bpfObjects, iface string) ([]io.Closer, error) {
	var closers []io.Closer
	add := func(c io.Closer, err error) error {
		if err != nil {
			closeAll(closers)
			return err
		}
		closers = append(closers, c)
		return nil
	}

	rawTracepoints := []struct {
		name string
		prog *ebpf.Program
	}{
		{"inet_sock_set_state", objs.NdInetSockSetState},
		{"tcp_retransmit_skb", objs.NdTcpRetransmitSkb},
		{"sched_wakeup", objs.NdSchedWakeup},
		{"sched_wakeup_new", objs.NdSchedWakeupNew},
		{"sched_switch", objs.NdSchedSwitch},
	}
	for _, rt := range rawTracepoints {
		l, err := link.AttachRawTracepoint(link.RawTracepointOptions{
			Name:    rt.name,
			Program: rt.prog,
		})
		if err != nil {
			closeAll(closers)
			return nil, fmt.Errorf("attach raw tracepoint %s: %w", rt.name, err)
		}
		closers = append(closers, l)
	}

	if err := add(link.Kprobe("tcp_rcv_established", objs.NdTcpRcvEstablished, nil)); err != nil {
		return nil, fmt.Errorf("attach kprobe tcp_rcv_established: %w", err)
	}

	tcClosers, err := attachTC(iface, objs.NdTcIngress, objs.NdTcEgress)
	if err != nil {
		closeAll(closers)
		return nil, fmt.Errorf("attach TC programs to %s: %w", iface, err)
	}
	closers = append(closers, tcClosers...)
	return closers, nil
}
