package monitor

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMonitor_CollectorCounterVec(t *testing.T) {
	type args struct {
		name     string
		num      int
		interval time.Duration
	}
	tests := []struct {
		name string
		args args
		want *prometheus.CounterVec
	}{
		{
			name: "test_push",
			args: args{
				name:     "push_test_counter",
				num:      10,
				interval: 10 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMonitor(SetPushWay("http://127.0.0.1:9091", "test_push"))
			m.NewMetric(&Metric{
				ID:          tt.args.name,
				Name:        tt.args.name,
				Description: "this is test metric",
				Type:        "counter_vec",
				Args:        []string{"push_count"},
			}, "testing")
			for i := 0; i < tt.args.num; i++ {
				got := m.CollectorCounterVec(tt.args.name)
				if got == nil {
					t.Errorf("Monitor.CollectorCounterVec() = %v, want %v", got, tt.want)
				}
				got.WithLabelValues("push_count").Inc()
				time.Sleep(tt.args.interval)
			}
		})
	}
}
