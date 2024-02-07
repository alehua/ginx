package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusBuilder struct {
	Namespace  string
	Subsystem  string
	Name       string
	Help       string
	InstanceID string
}

// BuildResponseTime 响应时间统计
func (p *PrometheusBuilder) BuildResponseTime() gin.HandlerFunc {
	// method: 方法, pattern: 路由, status: 状态码
	labels := []string{"method", "pattern", "status"}
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: p.Namespace,
		Subsystem: p.Subsystem,
		Name:      p.Name + "_resp_time",
		Help:      p.Help,
		ConstLabels: map[string]string{
			"instance_id": p.InstanceID,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(vector)
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		start := time.Now()
		defer func() {
			vector.WithLabelValues(method, ctx.FullPath(),
				strconv.Itoa(ctx.Writer.Status())).
				// 执行时间
				Observe(float64(time.Since(start).Milliseconds()))
		}()
		ctx.Next()
	}
}

// BuildActiveRequest 活跃请求数统计
func (p *PrometheusBuilder) BuildActiveRequest() gin.HandlerFunc {
	// 关心总的活跃请求数
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: p.Namespace,
		Subsystem: p.Subsystem,
		Name:      p.Name + "_active_req",
		Help:      p.Help,
		ConstLabels: map[string]string{
			"instance_id": p.InstanceID,
		},
	})
	prometheus.MustRegister(gauge)
	return func(ctx *gin.Context) {
		gauge.Inc()
		defer gauge.Dec()
		ctx.Next()
	}
}
