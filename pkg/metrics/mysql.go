package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tapclap/db-connect-checker/pkg/mysqlcheck"
	"github.com/tapclap/db-connect-checker/pkg/types"
)

// Пакет metrics предоставляет экспортеры метрик для мониторинга подключений к MySQL.
//
// Метрики:
//   - mysql_connection_available: доступность подключения (1 = доступно, 0 = недоступно)
//   - mysql_connection_duration_seconds: время выполнения проверки подключения в секундах
//
// Пример использования для нескольких баз данных:
//
//	configs := []types.MysqlConfig{...}
//	exporter := metrics.NewMultiMySQLExporter(configs, 30*time.Second)
//	exporter.Start() // запускает периодические проверки
//	prometheus.MustRegister(exporter)
//	http.Handle("/metrics", promhttp.Handler())
//	http.ListenAndServe(":8080", nil)

type MultiMySQLExporter struct {
	configs            []types.MysqlConfig
	availabilityMetric *prometheus.GaugeVec
	durationMetric     *prometheus.GaugeVec
	checkInterval      time.Duration
	mu                 sync.RWMutex
	ctx                context.Context
	cancel             context.CancelFunc
}

func NewMultiMySQLExporter(configs []types.MysqlConfig, checkInterval time.Duration) *MultiMySQLExporter {
	ctx, cancel := context.WithCancel(context.Background())

	if checkInterval == 0 {
		checkInterval = 30 * time.Second
	}

	return &MultiMySQLExporter{
		configs:       configs,
		checkInterval: checkInterval,
		ctx:           ctx,
		cancel:        cancel,
		availabilityMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mysql_connection_available",
				Help: "MySQL connection availability (1 = available, 0 = unavailable)",
			},
			[]string{"host", "port", "database"},
		),
		durationMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mysql_connection_duration_seconds",
				Help: "MySQL connection check duration in seconds",
			},
			[]string{"host", "port", "database"},
		),
	}
}

func (e *MultiMySQLExporter) Describe(ch chan<- *prometheus.Desc) {
	e.availabilityMetric.Describe(ch)
	e.durationMetric.Describe(ch)
}

func (e *MultiMySQLExporter) Start() {
	e.performChecks()

	go func() {
		ticker := time.NewTicker(e.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				e.performChecks()
			case <-e.ctx.Done():
				return
			}
		}
	}()
}

func (e *MultiMySQLExporter) Stop() {
	e.cancel()
}

func (e *MultiMySQLExporter) performChecks() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.availabilityMetric.Reset()
	e.durationMetric.Reset()

	var wg sync.WaitGroup
	for _, config := range e.configs {
		wg.Add(1)
		go func(cfg types.MysqlConfig) {
			defer wg.Done()

			startTime := time.Now()
			err := mysqlcheck.CheckConnection(cfg)

			duration := time.Since(startTime).Seconds()
			labels := prometheus.Labels{
				"host":     cfg.Host,
				"port":     cfg.Port,
				"database": cfg.Name,
			}

			if err != nil {
				e.availabilityMetric.With(labels).Set(0)
			} else {
				e.availabilityMetric.With(labels).Set(1)
			}

			e.durationMetric.With(labels).Set(duration)
		}(config)
	}
	wg.Wait()
}

func (e *MultiMySQLExporter) Collect(ch chan<- prometheus.Metric) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	e.availabilityMetric.Collect(ch)
	e.durationMetric.Collect(ch)
}
