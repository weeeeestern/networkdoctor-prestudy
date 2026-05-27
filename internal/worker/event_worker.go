//go:build linux

package worker

import (
	"context"
	"errors"
	"log"

	"networkdoctor-agent/internal/bpf"
)

// RunEventWorker consumes anomaly records emitted by the BPF ring buffer and
// logs them for now. A later backend can replace this with structured ingestion.
func RunEventWorker(ctx context.Context, rt *bpf.Runtime) {
	for {
		evt, err := rt.ReadEvent(ctx)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, bpf.ErrEventReaderClosed) {
				return
			}
			log.Print(err)
			continue
		}

		log.Printf(
			"anomaly event type=%d pid=%d ifindex=%d src=%s:%d dst=%s:%d value_us=%d aux=%d",
			evt.Type,
			evt.PID,
			evt.Ifindex,
			evt.SrcIP,
			evt.SrcPort,
			evt.DstIP,
			evt.DstPort,
			evt.ValueUs,
			evt.Aux,
		)
	}
}
