package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"networkdoctor-agent/internal/model"
)

type SnapshotSource interface {
	Snapshot() model.Snapshot
}

// networkDoctorCollector is a custom Prometheus collector backed by a snapshot
// source. Each scrape reads the latest values and emits metrics.
type networkDoctorCollector struct {
	source SnapshotSource

	tcpActive          *prometheus.Desc
	tcpTimeWait        *prometheus.Desc
	tcpClosed          *prometheus.Desc
	tcpConnectFailed   *prometheus.Desc
	tcpShortLived      *prometheus.Desc
	tcpConnectAttempts *prometheus.Desc
	tcpRetrans         *prometheus.Desc
	tcpRetransBurst    *prometheus.Desc
	tcpRTTSamples      *prometheus.Desc
	tcpCwndSamples     *prometheus.Desc

	dnsQueries     *prometheus.Desc
	dnsResponses   *prometheus.Desc
	dnsRcodeErrors *prometheus.Desc
	dnsSlow        *prometheus.Desc
	dnsRcode       *prometheus.Desc

	udpPackets   *prometheus.Desc
	udpLongFlows *prometheus.Desc

	tcpSRTTHist     *prometheus.Desc
	tcpCwndHist     *prometheus.Desc
	dnsLatencyHist  *prometheus.Desc
	runqlatHist     *prometheus.Desc
	udpDurationHist *prometheus.Desc
}

// newNetworkDoctorCollector defines metric descriptors once. The Collect
// method later supplies concrete values for each scrape.
func newNetworkDoctorCollector(source SnapshotSource) *networkDoctorCollector {
	return &networkDoctorCollector{
		source: source,

		tcpActive:          prometheus.NewDesc("ebpf_tcp_connections_active", "Current TCP ESTABLISHED connection delta summed across CPUs.", nil, nil),
		tcpTimeWait:        prometheus.NewDesc("ebpf_tcp_connections_time_wait", "Current TCP TIME_WAIT connection delta summed across CPUs.", nil, nil),
		tcpClosed:          prometheus.NewDesc("ebpf_tcp_connections_closed_total", "TCP connections observed leaving ESTABLISHED.", nil, nil),
		tcpConnectFailed:   prometheus.NewDesc("ebpf_tcp_connect_failed_total", "TCP connect attempts that failed before ESTABLISHED.", nil, nil),
		tcpShortLived:      prometheus.NewDesc("ebpf_tcp_connections_short_lived_total", "TCP connections shorter than the configured threshold.", nil, nil),
		tcpConnectAttempts: prometheus.NewDesc("ebpf_tcp_connect_attempts_total", "TCP SYN-like state transitions observed.", nil, nil),
		tcpRetrans:         prometheus.NewDesc("ebpf_tcp_retransmits_total", "TCP retransmission events observed.", nil, nil),
		tcpRetransBurst:    prometheus.NewDesc("ebpf_tcp_retransmit_bursts_total", "TCP retransmission bursts exceeding the configured threshold.", nil, nil),
		tcpRTTSamples:      prometheus.NewDesc("ebpf_tcp_rtt_samples_total", "TCP RTT samples observed by tcp_rcv_established.", nil, nil),
		tcpCwndSamples:     prometheus.NewDesc("ebpf_tcp_cwnd_samples_total", "TCP cwnd samples observed by tcp_rcv_established.", nil, nil),

		dnsQueries:     prometheus.NewDesc("ebpf_dns_queries_total", "DNS queries observed on UDP port 53.", nil, nil),
		dnsResponses:   prometheus.NewDesc("ebpf_dns_responses_total", "DNS responses observed on UDP port 53.", nil, nil),
		dnsRcodeErrors: prometheus.NewDesc("ebpf_dns_rcode_errors_total", "DNS responses with non-zero rcode.", nil, nil),
		dnsSlow:        prometheus.NewDesc("ebpf_dns_slow_total", "DNS responses slower than the configured threshold.", nil, nil),
		dnsRcode:       prometheus.NewDesc("ebpf_dns_rcode_total", "DNS response count by rcode.", []string{"rcode"}, nil),

		udpPackets:   prometheus.NewDesc("ebpf_udp_packets_total", "UDP packets observed by TC programs.", nil, nil),
		udpLongFlows: prometheus.NewDesc("ebpf_udp_long_flows_total", "UDP flows exceeding the configured duration threshold.", nil, nil),

		tcpSRTTHist:     prometheus.NewDesc("ebpf_tcp_srtt_microseconds", "TCP SRTT samples exposed from log2 BPF buckets.", nil, nil),
		tcpCwndHist:     prometheus.NewDesc("ebpf_tcp_cwnd", "TCP congestion window samples exposed from log2 BPF buckets.", nil, nil),
		dnsLatencyHist:  prometheus.NewDesc("ebpf_dns_query_latency", "DNS query latency samples exposed from log2 BPF buckets.", nil, nil),
		runqlatHist:     prometheus.NewDesc("ebpf_runqlat", "Scheduler run queue latency samples exposed from log2 BPF buckets.", nil, nil),
		udpDurationHist: prometheus.NewDesc("ebpf_udp_session_duration", "Long UDP session duration samples exposed from log2 BPF buckets.", nil, nil),
	}
}

// Describe sends every metric descriptor to Prometheus during collector
// registration.
func (c *networkDoctorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.tcpActive
	ch <- c.tcpTimeWait
	ch <- c.tcpClosed
	ch <- c.tcpConnectFailed
	ch <- c.tcpShortLived
	ch <- c.tcpConnectAttempts
	ch <- c.tcpRetrans
	ch <- c.tcpRetransBurst
	ch <- c.tcpRTTSamples
	ch <- c.tcpCwndSamples
	ch <- c.dnsQueries
	ch <- c.dnsResponses
	ch <- c.dnsRcodeErrors
	ch <- c.dnsSlow
	ch <- c.dnsRcode
	ch <- c.udpPackets
	ch <- c.udpLongFlows
	ch <- c.tcpSRTTHist
	ch <- c.tcpCwndHist
	ch <- c.dnsLatencyHist
	ch <- c.runqlatHist
	ch <- c.udpDurationHist
}

// Collect is called on every /metrics scrape. It converts snapshot counters and
// histograms into Prometheus metric samples.
func (c *networkDoctorCollector) Collect(ch chan<- prometheus.Metric) {
	snap := c.source.Snapshot()

	ch <- prometheus.MustNewConstMetric(c.tcpActive, prometheus.GaugeValue, float64(snap.TCPState.Established))
	ch <- prometheus.MustNewConstMetric(c.tcpTimeWait, prometheus.GaugeValue, float64(snap.TCPState.TimeWait))

	ch <- prometheus.MustNewConstMetric(c.tcpClosed, prometheus.CounterValue, float64(snap.TCP.ClosedTotal))
	ch <- prometheus.MustNewConstMetric(c.tcpConnectFailed, prometheus.CounterValue, float64(snap.TCP.ConnectFailedTotal))
	ch <- prometheus.MustNewConstMetric(c.tcpShortLived, prometheus.CounterValue, float64(snap.TCP.ShortLivedTotal))
	ch <- prometheus.MustNewConstMetric(c.tcpConnectAttempts, prometheus.CounterValue, float64(snap.TCP.ConnectAttemptTotal))
	ch <- prometheus.MustNewConstMetric(c.tcpRetrans, prometheus.CounterValue, float64(snap.TCP.RetransTotal))
	ch <- prometheus.MustNewConstMetric(c.tcpRetransBurst, prometheus.CounterValue, float64(snap.TCP.RetransBurstTotal))
	ch <- prometheus.MustNewConstMetric(c.tcpRTTSamples, prometheus.CounterValue, float64(snap.TCP.RTTSampleTotal))
	ch <- prometheus.MustNewConstMetric(c.tcpCwndSamples, prometheus.CounterValue, float64(snap.TCP.CwndSampleTotal))

	ch <- prometheus.MustNewConstMetric(c.dnsQueries, prometheus.CounterValue, float64(snap.DNS.QueryTotal))
	ch <- prometheus.MustNewConstMetric(c.dnsResponses, prometheus.CounterValue, float64(snap.DNS.ResponseTotal))
	ch <- prometheus.MustNewConstMetric(c.dnsRcodeErrors, prometheus.CounterValue, float64(snap.DNS.RcodeErrorTotal))
	ch <- prometheus.MustNewConstMetric(c.dnsSlow, prometheus.CounterValue, float64(snap.DNS.SlowTotal))

	ch <- prometheus.MustNewConstMetric(c.udpPackets, prometheus.CounterValue, float64(snap.UDP.PacketTotal))
	ch <- prometheus.MustNewConstMetric(c.udpLongFlows, prometheus.CounterValue, float64(snap.UDP.LongFlowTotal))

	collectRcodeCounters(ch, c.dnsRcode, snap.DNS.RcodeTotal)
	emitLog2Histogram(ch, c.tcpSRTTHist, snap.Hist.TCPSRTTUs)
	emitLog2Histogram(ch, c.tcpCwndHist, snap.Hist.TCPCwnd)
	emitLog2Histogram(ch, c.dnsLatencyHist, snap.Hist.DNSLatencyUs)
	emitLog2Histogram(ch, c.runqlatHist, snap.Hist.RunqlatUs)
	emitLog2Histogram(ch, c.udpDurationHist, snap.Hist.UDPDurationUs)
}

// collectRcodeCounters exports one Prometheus sample per DNS rcode bucket.
func collectRcodeCounters(ch chan<- prometheus.Metric, desc *prometheus.Desc, totals [16]uint64) {
	for rcode := 0; rcode < len(totals); rcode++ {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(totals[rcode]), fmt.Sprint(rcode))
	}
}
