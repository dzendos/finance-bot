package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func CollectMetrics(logger *zap.Logger) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			logger.Fatal("error starting http server", zap.Error(err))
		}
	}()
}
