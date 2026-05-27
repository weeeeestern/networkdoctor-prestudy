//go:build linux

package bpf

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

type Options struct {
	Iface string

	SlowTCPRTTUs         uint64
	ShortTCPConnUs       uint64
	SlowDNSUs            uint64
	LongUDPFlowUs        uint64
	HighRunqlatUs        uint64
	RetransBurstWindowNs uint64
	EventCooldownNs      uint64

	RetransBurstCount uint64
	CwndDropPercent   uint64
}

type Runtime struct {
	objs   bpfObjects
	links  []io.Closer
	reader *ringbuf.Reader

	closeOnce sync.Once
	closeErr  error
}

func Start(opts Options) (*Runtime, error) {
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("remove memlock limit: %w", err)
	}

	rt := &Runtime{}
	if err := loadBpfObjects(&rt.objs, nil); err != nil {
		return nil, fmt.Errorf("load BPF objects: %w", err)
	}

	if err := writeConfig(&rt.objs, opts); err != nil {
		_ = rt.objs.Close()
		return nil, fmt.Errorf("write nd_config: %w", err)
	}

	links, err := attachPrograms(&rt.objs, opts.Iface)
	if err != nil {
		_ = rt.objs.Close()
		return nil, err
	}
	rt.links = links

	rd, err := ringbuf.NewReader(rt.objs.NdEvents)
	if err != nil {
		closeAll(rt.links)
		_ = rt.objs.Close()
		return nil, fmt.Errorf("open nd_events ring buffer: %w", err)
	}
	rt.reader = rd

	return rt, nil
}

func (r *Runtime) Close() error {
	r.closeOnce.Do(func() {
		var errs []error
		if r.reader != nil {
			errs = append(errs, r.reader.Close())
		}
		closeAll(r.links)
		errs = append(errs, r.objs.Close())
		r.closeErr = errors.Join(errs...)
	})
	return r.closeErr
}
