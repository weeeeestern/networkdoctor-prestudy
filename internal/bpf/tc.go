//go:build linux

package bpf

import (
	"fmt"
	"io"

	"github.com/cilium/ebpf"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// attachTC installs a clsact qdisc and replaces ingress/egress BPF filters on
// the requested network interface.
func attachTC(iface string, ingressProg, egressProg *ebpf.Program) ([]io.Closer, error) {
	dev, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, fmt.Errorf("find interface: %w", err)
	}

	qdisc := &netlink.Clsact{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: dev.Attrs().Index,
			Handle:    netlink.MakeHandle(0xffff, 0),
			Parent:    netlink.HANDLE_CLSACT,
		},
	}
	if err := netlink.QdiscAdd(qdisc); err != nil && !isExists(err) {
		return nil, fmt.Errorf("add clsact qdisc: %w", err)
	}

	ingress := tcFilter(dev.Attrs().Index, netlink.HANDLE_MIN_INGRESS, 1, "nd_tc_ingress", ingressProg)
	egress := tcFilter(dev.Attrs().Index, netlink.HANDLE_MIN_EGRESS, 1, "nd_tc_egress", egressProg)

	for _, filter := range []*netlink.BpfFilter{ingress, egress} {
		if err := netlink.FilterReplace(filter); err != nil {
			return nil, fmt.Errorf("replace filter %s: %w", filter.Name, err)
		}
	}

	return []io.Closer{
		tcFilterCloser{filter: ingress},
		tcFilterCloser{filter: egress},
	}, nil
}

// tcFilter builds a direct-action BPF filter. Direct action means the BPF
// program returns TC_ACT_* values directly instead of using a separate action.
func tcFilter(linkIndex int, parent uint32, priority uint16, name string, prog *ebpf.Program) *netlink.BpfFilter {
	return &netlink.BpfFilter{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: linkIndex,
			Parent:    parent,
			Handle:    netlink.MakeHandle(0, priority),
			Protocol:  unix.ETH_P_ALL,
			Priority:  priority,
		},
		Fd:           prog.FD(),
		Name:         name,
		DirectAction: true,
	}
}

// tcFilterCloser adapts netlink.FilterDel to io.Closer so TC filters can share
// the same cleanup path as cilium/ebpf links.
type tcFilterCloser struct {
	filter netlink.Filter
}

func (c tcFilterCloser) Close() error {
	err := netlink.FilterDel(c.filter)
	if err != nil && !isNotFound(err) {
		return err
	}
	return nil
}
