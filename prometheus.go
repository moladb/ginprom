package ginprom

import (
	"fmt"

	"github.com/gin-gonic/gin"
	prom "github.com/prometheus/client_golang/prometheus"
)

var DefaultInstrument = NewInstrument()

func init() {
	prom.MustRegister(DefaultInstrument.handledCounter)
}

type Instrument struct {
	handledCounter   *prom.CounterVec
	handledHistogram *prom.HistogramVec
}

func NewInstrument() *Instrument {
	return &Instrument{
		handledCounter: prom.NewCounterVec(prom.CounterOpts{
			Name: "gin_rest_handled_total",
			Help: "Total number of api handled on the server, regardless of success or failure",
		}, []string{"method", "path", "status_code"}),
		handledHistogram: nil,
	}
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
	return func(c *gin.Context) {
		handler(c)
		DefaultInstrument.handledCounter.WithLabelValues(c.Request.Method, path, fmt.Sprintf("%d", c.Writer.Status())).Inc()
	}
}

func (i *Instrument) WithMetrics(path string, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c)
		i.handledCounter.WithLabelValues(c.Request.Method, path, fmt.Sprintf("%d", c.Writer.Status())).Inc()
	}
}
