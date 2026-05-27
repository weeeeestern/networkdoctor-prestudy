package metrics

import (
	"math"

	"github.com/prometheus/client_golang/prometheus"

	"networkdoctor-agent/internal/model"
)

// emitLog2Histogram converts the BPF program's fixed log2 bucket array into a
// Prometheus histogram. Bucket boundaries are powers of two.
func emitLog2Histogram(ch chan<- prometheus.Metric, desc *prometheus.Desc, counts [model.HistBuckets]uint64) {
	var total uint64
	var estimatedSum float64

	for i, count := range counts {
		total += count

		// The BPF program stores only log2 buckets, not exact samples. Use the
		// bucket upper bound as a conservative sum estimate for _sum.
		estimatedSum += float64(count) * float64(log2UpperBound(uint32(i)))
	}

	cumulative := make(map[float64]uint64, model.HistBuckets-1)
	var running uint64
	for i := 0; i < model.HistBuckets-1; i++ {
		running += counts[i]
		cumulative[float64(log2UpperBound(uint32(i)))] = running
	}

	ch <- prometheus.MustNewConstHistogram(desc, total, estimatedSum, cumulative)
}

// log2UpperBound returns the inclusive upper bound represented by a log2 bucket
// index used in the BPF histogram arrays.
func log2UpperBound(bucket uint32) uint64 {
	if bucket >= 63 {
		return math.MaxUint64
	}
	return uint64(1) << bucket
}
