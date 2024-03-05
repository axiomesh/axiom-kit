package pebble

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	diskSizeGauge            prometheus.Gauge // Gauge for tracking the size of all the levels in the database
	diskWriteThroughput      prometheus.Gauge // Gauge for measuring disk data written throughput
	walWriteThroughput       prometheus.Gauge // Gauge for measuring wal data written throughput
	effectiveWriteThroughput prometheus.Gauge // Gauge for measuring the kv effective amount of data written throughput
}

type MetricsOption func(pebbleMetrics *Metrics)

func WithDiskSizeGauge(namespace, subSystem, namePrefix string) MetricsOption {
	return func(pebbleMetrics *Metrics) {
		pebbleMetrics.diskSizeGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      namePrefix + "_kv_pebble_" + "disk_size",
			Help:      "disk size(MB)",
		})
		prometheus.MustRegister(pebbleMetrics.diskSizeGauge)
	}
}

func WithDiskWriteThroughput(namespace, subSystem, namePrefix string) MetricsOption {
	return func(pebbleMetrics *Metrics) {
		pebbleMetrics.diskWriteThroughput = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      namePrefix + "_kv_pebble_" + "disk_write_throughput",
			Help:      "disk write throughput, MB/s",
		})
		prometheus.MustRegister(pebbleMetrics.diskWriteThroughput)
	}
}

func WithWalWriteThroughput(namespace, subSystem, namePrefix string) MetricsOption {
	return func(pebbleMetrics *Metrics) {
		pebbleMetrics.walWriteThroughput = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      namePrefix + "_kv_pebble_" + "wal_write_throughput",
			Help:      "wal write throughput, MB/s",
		})
		prometheus.MustRegister(pebbleMetrics.walWriteThroughput)
	}
}

func WithEffectiveWriteThroughput(namespace, subSystem, namePrefix string) MetricsOption {
	return func(pebbleMetrics *Metrics) {
		pebbleMetrics.effectiveWriteThroughput = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      namePrefix + "_kv_pebble_" + "effective_write_throughput",
			Help:      "effective write throughput, MB/s",
		})
		prometheus.MustRegister(pebbleMetrics.effectiveWriteThroughput)
	}
}
