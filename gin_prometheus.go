package gin_prometheus

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricsPath = "/metrics"
	faviconPath = "/favicon.ico"
)

type handlerPath sync.Map

func (h *handlerPath) get(uri string) string {
	v, ok := h.Load(handler)
	if !ok {
		return ""
	}
	return v.(string)
}

func (h *handlerPath)set(uri gin.RouteInfo) {
	h.Store(uri.Handler, uri.Path)
}

type GinPrometheus struct {
	serverName string
	engine  *gin.Engine
	ignored map[string]bool
	pathMap *handlerPath
	updated bool
}

type Option func(*GinPrometheus)

func Ignore(path ...string) Option {
	return func(gp *GinPrometheus) {
		for _, p := range path {
			gp.ignored[p] = true
		}
	}
}

// New gin_prometheus
func New(engine *gin.Engine, options ...Option) *GinPrometheus {
	gp := &GinPrometheus{
		engine: engine,
		ignored: map[string]bool{
			metricsPath: true,
			faviconPath: true,
		},
		pathMap: &handlerPath{},
	}

	for _, o := options {
		o(gp)
	}
	engine.RouterGroup.GET("/metrics", prometheus.Handler())
	return gp
}

func (gp *GinPrometheus) updatePath() {
	gp.updated = true
	for _, uri := range gp.engine.Routes() {
		gp.pathMap.set(uri)
	}
}

func (gp *GinPrometheus)Middleware() gin.HandlerFunc{
	return func (c *gin.Context) {
		if !gp.updated {
			gp.updatePath()
		}

		if gp.ignored[c.Request.URL.String()] {
			c.Next()
		}

		start := time.Now()
		c.Next()

		// TODO
	}
}
