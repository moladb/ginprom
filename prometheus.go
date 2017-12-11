package ginprom

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	prom "github.com/prometheus/client_golang/prometheus"
)

var DefaultInstrument = NewInstrument()

func init() {
	prom.MustRegister(DefaultInstrument.handledCounter)
}

type Instrument struct {
	apiGroup         string
	handledCounter   *prom.CounterVec
	handledHistogram *prom.HistogramVec
}

type InstrumentOption func(i *Instrument)

func WithAPIGroup(apiGroup string) InstrumentOption {
	return func(i *Instrument) { i.apiGroup = apiGroup }
}

func WithHistogram(buckets []float64) InstrumentOption {
	return func(i *Instrument) {
		i.handledHistogram = prom.NewHistogramVec(prom.HistogramOpts{
			Name: "gin_handled_latency",
			Help: "Histogram of response latency (seconds) handled by the server",
		}, []string{"method", "path", "status_code"})
		prom.MustRegister(i.handledHistogram)
	}
}

func NewInstrument(opts ...InstrumentOption) *Instrument {
	i := &Instrument{
		handledCounter: prom.NewCounterVec(prom.CounterOpts{
			Name: "gin_handled_total",
			Help: "Total number of request handled by the server, regardless of success or failure",
		}, []string{"method", "path", "status_code"}),
		handledHistogram: nil,
	}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

func (i *Instrument) Describe(ch chan<- *prom.Desc) {
	i.handledCounter.Describe(ch)
	if i.handledHistogram != nil {
		i.handledHistogram.Describe(ch)
	}
}

func (i *Instrument) Collect(ch chan<- prom.Metric) {
	i.handledCounter.Collect(ch)
	if i.handledHistogram != nil {
		i.handledHistogram.Collect(ch)
	}
}

func WithMetrics(path string, handler gin.HandlerFunc) gin.HandlerFunc {
	h := DefaultInstrument.WithMetrics(path, handler)
	return func(c *gin.Context) {
		h(c)
	}
}

func (i *Instrument) WithMetrics(path string, handler gin.HandlerFunc) gin.HandlerFunc {
	fullPath := fmt.Sprintf("%s%s", i.apiGroup, path)
	return func(c *gin.Context) {
		if i.handledHistogram == nil {
			handler(c)
			i.handledCounter.WithLabelValues(c.Request.Method, fullPath, fmt.Sprintf("%d", c.Writer.Status())).Inc()
		} else {
			startTime := time.Now()
			handler(c)
			status := fmt.Sprintf("%d", c.Writer.Status())
			i.handledCounter.WithLabelValues(c.Request.Method, fullPath, status).Inc()
			i.handledHistogram.WithLabelValues(c.Request.Method, fullPath, status).Observe(time.Since(startTime).Seconds())
		}
	}
}
