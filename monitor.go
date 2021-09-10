package monitor

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

var (
	monitor      *Monitor
	once         sync.Once
	pushInterval time.Duration = 10 * time.Minute
)

type Monitor struct {
	engine               *gin.Engine
	pusher               *push.Pusher
	lastWriteMetricsTime *time.Timer // 写倒计时，超过10分钟，push一次，每次写重置倒计时，仅push模式使用

	MetricsList []*Metric
}

type Option func(*Monitor)

func GinEngine(engine *gin.Engine, serverName string) Option {
	return func(m *Monitor) {
		m.engine = engine
		p := ginprometheus.NewPrometheus(serverName)
		p.Use(engine)
	}
}

func SetPushWay(prometheusAddr, serverName string) Option {
	return func(m *Monitor) {
		m.engine = nil
		m.pusher = push.New(prometheusAddr, serverName)
		m.pusher.Gatherer(prometheus.DefaultGatherer)
		m.lastWriteMetricsTime = time.AfterFunc(pushInterval, m.Push)
	}
}

func NewMonitor(opts ...Option) *Monitor {
	once.Do(func() {
		monitor = &Monitor{}
		for _, o := range opts {
			o(monitor)
		}
	})
	return monitor
}

type Metric struct {
	MetricCollector prometheus.Collector
	ID              string
	Name            string
	Description     string
	Type            string
	Args            []string
}

// NewMetric associates prometheus.Collector based on Metric.Type
func (monitor *Monitor) NewMetric(m *Metric, subsystem string) (prometheus.Collector, error) {
	var metric prometheus.Collector
	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "counter":
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "gauge_vec":
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "gauge":
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "histogram_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "histogram":
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "summary":
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}

	if err := prometheus.Register(metric); err != nil {
		return nil, errors.Wrapf(err, "%s could not be registered in Prometheus", m.Name)
	}
	m.MetricCollector = metric

	monitor.registerMetric(m, subsystem)
	return metric, nil
}

func (monitor *Monitor) Push() {
	if monitor.pusher == nil {
		return
	}

	monitor.pusher.Push()
	monitor.Reset()
}

func (monitor *Monitor) Reset() {
	if monitor.pusher == nil {
		return
	}

	monitor.lastWriteMetricsTime.Reset(pushInterval)
}

func (monitor *Monitor) registerMetric(m *Metric, subsystem string) {
	for _, metricDef := range monitor.MetricsList {
		if metricDef.Name == m.Name {
			return
		}
	}
	monitor.MetricsList = append(monitor.MetricsList, m)
	if monitor.pusher != nil {
		monitor.pusher.Collector(m.MetricCollector)
	}
}

func (monitor *Monitor) CollectorCounterVec(name string) *prometheus.CounterVec {
	for _, m := range monitor.MetricsList {
		if m.Name == name {
			return m.MetricCollector.(*prometheus.CounterVec)
		}
	}
	return nil
}

// func (monitor *Monitor) CollectorCounter(name string) *prometheus.CounterVec {
// 	for _, m := range monitor.MetricsList {
// 		if m.Name == name {
// 			return m.MetricCollector.(*prometheus.CounterVec)
// 		}
// 	}
// 	return nil
// }

func (monitor *Monitor) CollectorHistogramVec(name string) *prometheus.HistogramVec {
	for _, m := range monitor.MetricsList {
		if m.Name == name {
			return m.MetricCollector.(*prometheus.HistogramVec)
		}
	}
	return nil
}

// func (monitor *Monitor) CollectorHistogram(name string) *prometheus.Histogram {
// 	for _, m := range monitor.MetricsList {
// 		if m.Name == name {
// 			return m.MetricCollector.(*prometheus.Histogram)
// 		}
// 	}
// 	return nil
// }

func (monitor *Monitor) CollectorGaugeVec(name string) *prometheus.GaugeVec {
	for _, m := range monitor.MetricsList {
		if m.Name == name {
			return m.MetricCollector.(*prometheus.GaugeVec)
		}
	}
	return nil
}

// func (monitor *Monitor) CollectorGauge(name string) *prometheus.Gauge {
// 	for _, m := range monitor.MetricsList {
// 		if m.Name == name {
// 			return m.MetricCollector.(*prometheus.Gauge)
// 		}
// 	}
// 	return nil
// }

func (monitor *Monitor) CollectorSummaryVec(name string) *prometheus.SummaryVec {
	for _, m := range monitor.MetricsList {
		if m.Name == name {
			return m.MetricCollector.(*prometheus.SummaryVec)
		}
	}
	return nil
}

// func (monitor *Monitor) CollectorSummary(name string) *prometheus.Summary {
// 	for _, m := range monitor.MetricsList {
// 		if m.Name == name {
// 			return m.MetricCollector.(*prometheus.Summary)
// 		}
// 	}
// 	return nil
// }
