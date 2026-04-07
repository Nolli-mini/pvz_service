package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Технические метрики
var (
	// Количество запросов
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Время ответа
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

// Бизнес метрики
var (
	// Количество созданных ПВЗ
	PVZsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pvzs_created_total",
			Help: "Total number of PVZs created",
		},
	)

	// Количество созданных приёмок
	ReceptionsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "receptions_created_total",
			Help: "Total number of receptions created",
		},
	)

	// Количество добавленных товаров
	ProductsAddedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_added_total",
			Help: "Total number of products added",
		},
	)
)

// Функции для инкремента бизнес метрик
func IncPVZsCreated() {
	PVZsCreatedTotal.Inc()
}

func IncReceptionsCreated() {
	ReceptionsCreatedTotal.Inc()
}

func IncProductsAdded() {
	ProductsAddedTotal.Inc()
}
