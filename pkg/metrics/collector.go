package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	FleetReplicas = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gfd_fleet_replicas",
		Help: "Current replica count per fleet.",
	}, []string{"fleet", "namespace"})

	FleetPlayersTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gfd_fleet_players_total",
		Help: "Total players across all servers in a fleet.",
	}, []string{"fleet", "namespace"})

	FleetBufferAvailable = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gfd_fleet_buffer_available",
		Help: "Number of warm, unallocated servers in the buffer pool.",
	}, []string{"fleet", "namespace"})

	ScaleUpTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gfd_scale_up_total",
		Help: "Total number of scale-up events.",
	}, []string{"fleet", "namespace"})

	ScaleDownTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gfd_scale_down_total",
		Help: "Total number of scale-down events.",
	}, []string{"fleet", "namespace"})

	AllocationTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gfd_allocation_total",
		Help: "Total allocation requests by fleet, strategy, and phase.",
	}, []string{"fleet", "strategy", "phase"})

	AllocationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gfd_allocation_latency_seconds",
		Help:    "Allocation request latency distribution.",
		Buckets: prometheus.DefBuckets,
	}, []string{"fleet", "strategy"})

	DrainDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gfd_drain_duration_seconds",
		Help:    "Drain duration distribution.",
		Buckets: []float64{10, 30, 60, 120, 300, 600, 900, 1800},
	}, []string{"fleet"})

	ReconcileDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "gfd_reconcile_duration_seconds",
		Help: "Reconcile loop duration distribution.",
		Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 15, 30},
	}, []string{"controller"})

	CircuitBreakerTriggered = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gfd_circuit_breaker_triggered",
		Help: "Number of times the circuit breaker opened.",
	}, []string{"fleet", "namespace"})

	ReconcileErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gfd_reconcile_errors_total",
		Help: "Total number of reconcile errors.",
	}, []string{"controller"})
)

// SetFleetMetrics updates all fleet-level gauges.
func SetFleetMetrics(fleet, namespace string, replicas, players, buffer int32) {
	FleetReplicas.WithLabelValues(fleet, namespace).Set(float64(replicas))
	FleetPlayersTotal.WithLabelValues(fleet, namespace).Set(float64(players))
	FleetBufferAvailable.WithLabelValues(fleet, namespace).Set(float64(buffer))
}

// RecordScaleUp increments the scale-up counter.
func RecordScaleUp(fleet, namespace string) {
	ScaleUpTotal.WithLabelValues(fleet, namespace).Inc()
}

// RecordScaleDown increments the scale-down counter.
func RecordScaleDown(fleet, namespace string) {
	ScaleDownTotal.WithLabelValues(fleet, namespace).Inc()
}

// RecordAllocation increments the allocation counter and observes latency.
func RecordAllocation(fleet, strategy, phase string, latencySeconds float64) {
	AllocationTotal.WithLabelValues(fleet, strategy, phase).Inc()
	AllocationLatency.WithLabelValues(fleet, strategy).Observe(latencySeconds)
}
