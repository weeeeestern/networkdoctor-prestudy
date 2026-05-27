package model

const HistBuckets = 32

type Snapshot struct {
	TCP      TCPMetrics
	DNS      DNSMetrics
	UDP      UDPMetrics
	TCPState TCPStateMetrics
	Hist     Histograms
}

type TCPMetrics struct {
	ClosedTotal         uint64
	ShortLivedTotal     uint64
	ConnectAttemptTotal uint64
	ConnectFailedTotal  uint64
	RetransTotal        uint64
	RetransBurstTotal   uint64
	RTTSampleTotal      uint64
	CwndSampleTotal     uint64
}

type DNSMetrics struct {
	QueryTotal      uint64
	ResponseTotal   uint64
	RcodeErrorTotal uint64
	SlowTotal       uint64
	RcodeTotal      [16]uint64
}

type UDPMetrics struct {
	PacketTotal   uint64
	LongFlowTotal uint64
}

type TCPStateMetrics struct {
	Established int64
	TimeWait    int64
}

type Histograms struct {
	TCPSRTTUs     [HistBuckets]uint64
	TCPCwnd       [HistBuckets]uint64
	DNSLatencyUs  [HistBuckets]uint64
	RunqlatUs     [HistBuckets]uint64
	UDPDurationUs [HistBuckets]uint64
}

type Event struct {
	Type    uint32
	PID     uint32
	Ifindex uint32

	SrcIP   string
	DstIP   string
	SrcPort uint16
	DstPort uint16

	ValueUs uint64
	Aux     uint64
}

type GCStats struct {
	DNSPendingDeleted int
	UDPFlowsDeleted   int
}
