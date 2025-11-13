package metrics_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tapclap/db-connect-checker/pkg/metrics"
	"github.com/tapclap/db-connect-checker/pkg/types"
)

// ExampleNewMultiMySQLExporter демонстрирует создание экспортера для нескольких баз данных
func ExampleNewMultiMySQLExporter() {
	// Создаем конфигурации для нескольких баз данных
	configs := []types.MysqlConfig{
		{
			Name: "db1",
			User: "root",
			Pass: "password",
			Host: "localhost",
			Port: "3306",
			TLS:  false,
		},
		{
			Name: "db2",
			User: "root",
			Pass: "password",
			Host: "localhost",
			Port: "3307",
			TLS:  false,
		},
	}

	// Создаем экспортер для нескольких БД с интервалом проверки 30 секунд
	exporter := metrics.NewMultiMySQLExporter(configs, 30*time.Second)

	// Запускаем фоновые проверки
	exporter.Start()
	defer exporter.Stop()

	// Регистрируем экспортер в Prometheus
	prometheus.MustRegister(exporter)

	// Настраиваем HTTP сервер для метрик
	http.Handle("/metrics", promhttp.Handler())

	// Запускаем HTTP сервер
	fmt.Println("Metrics available at http://localhost:8080/metrics")
	// http.ListenAndServe(":8080", nil)
}
