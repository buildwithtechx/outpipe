package telemetry

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type MetricsExporter struct {
	mu             sync.RWMutex
	counters       map[string]*int64
	gauges         map[string]*int64
	workerStatuses map[string]string
}

func NewMetricsExporter() *MetricsExporter {
	return &MetricsExporter{
		counters:       make(map[string]*int64),
		gauges:         make(map[string]*int64),
		workerStatuses: make(map[string]string),
	}
}

func (m *MetricsExporter) IncCounter(name string, val int64) {
	m.mu.Lock()
	ptr, ok := m.counters[name]

	if !ok {
		var v int64
		ptr = &v
		m.counters[name] = ptr
	}

	m.mu.Unlock()
	atomic.AddInt64(ptr, val)
}

func (m *MetricsExporter) SetGauge(name string, val int64) {
	m.mu.Lock()
	ptr, ok := m.gauges[name]

	if !ok {
		var v int64
		ptr = &v
		m.gauges[name] = ptr
	}

	m.mu.Unlock()
	atomic.StoreInt64(ptr, val)
}

func (m *MetricsExporter) SetWorkerStatus(jobName string, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workerStatuses[jobName] = status
}

func (m *MetricsExporter) ExportPrometheus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("# HELP outpipe_metrics System operational metrics\n")

	counters := make([]string, 0, len(m.counters))
	for name := range m.counters {
		counters = append(counters, name)
	}
	sort.Strings(counters)
	for _, name := range counters {
		ptr := m.counters[name]
		sb.WriteString(fmt.Sprintf("# TYPE %s counter\n", name))
		sb.WriteString(fmt.Sprintf("%s %d\n", name, atomic.LoadInt64(ptr)))
	}

	gauges := make([]string, 0, len(m.gauges))
	for name := range m.gauges {
		gauges = append(gauges, name)
	}
	sort.Strings(gauges)
	for _, name := range gauges {
		ptr := m.gauges[name]
		sb.WriteString(fmt.Sprintf("# TYPE %s gauge\n", name))
		sb.WriteString(fmt.Sprintf("%s %d\n", name, atomic.LoadInt64(ptr)))
	}

	for job, status := range m.workerStatuses {
		val := 1

		if status != "succeeded" && status != "running" && status != "idle" {
			val = 0
		}

		sb.WriteString(fmt.Sprintf("worker_health{job=\"%s\",status=\"%s\"} %d\n", escapeLabel(job), escapeLabel(status), val))
	}

	return sb.String()
}

func escapeLabel(value string) string {
	return strings.NewReplacer(`\\`, `\\\\`, `"`, `\\"`, "\n", `\n`).Replace(value)
}
