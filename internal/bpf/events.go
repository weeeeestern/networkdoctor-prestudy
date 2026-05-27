//go:build linux

package bpf

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/cilium/ebpf/ringbuf"

	"networkdoctor-agent/internal/model"
)

var ErrEventReaderClosed = errors.New("event reader closed")

// ReadEvent consumes and decodes one anomaly record emitted by the BPF ring
// buffer.
func (r *Runtime) ReadEvent(ctx context.Context) (model.Event, error) {
	record, err := r.reader.Read()
	if err != nil {
		if ctx.Err() != nil || errors.Is(err, ringbuf.ErrClosed) {
			return model.Event{}, ErrEventReaderClosed
		}
		return model.Event{}, fmt.Errorf("read ringbuf event: %w", err)
	}

	var evt bpfNdEvent
	if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &evt); err != nil {
		return model.Event{}, fmt.Errorf("decode ringbuf event: %w", err)
	}

	return model.Event{
		Type:    evt.Type,
		PID:     evt.Pid,
		Ifindex: evt.Ifindex,
		SrcIP:   ipv4String(evt.SrcIp),
		DstIP:   ipv4String(evt.DstIp),
		SrcPort: evt.SrcPort,
		DstPort: evt.DstPort,
		ValueUs: evt.ValueUs,
		Aux:     evt.Aux,
	}, nil
}

// ipv4String formats a uint32 IPv4 address from BPF event/map data for logs.
func ipv4String(ip uint32) string {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], ip)
	return net.IP(b[:]).String()
}
