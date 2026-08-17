package relay

import (
	"fmt"
	"sync/atomic"
)

type Metrics struct {
	connections atomic.Int64
	tunnels     atomic.Int64
	frames      atomic.Int64
	bytes       atomic.Int64
	errors      atomic.Int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) AddConnection(delta int64) { m.connections.Add(delta) }
func (m *Metrics) AddTunnel(delta int64)     { m.tunnels.Add(delta) }
func (m *Metrics) AddFrame(bytes int64)      { m.frames.Add(1); m.bytes.Add(bytes) }
func (m *Metrics) AddError()                 { m.errors.Add(1) }

func (m *Metrics) Snapshot() map[string]int64 {
	return map[string]int64{"connections": m.connections.Load(), "tunnels": m.tunnels.Load(), "frames": m.frames.Load(), "bytes": m.bytes.Load(), "errors": m.errors.Load()}
}

func (m *Metrics) Prometheus() string {
	snapshot := m.Snapshot()
	return fmt.Sprintf("outpipe_connections %d\noutpipe_tunnels %d\noutpipe_frames %d\noutpipe_bytes %d\noutpipe_errors %d\n", snapshot["connections"], snapshot["tunnels"], snapshot["frames"], snapshot["bytes"], snapshot["errors"])
}
