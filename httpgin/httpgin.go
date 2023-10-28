package httpgin

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sync"
)

type ginFunc func(*Config) gin.HandlerFunc

type Manager interface {
	// GetEngine returns the gin.Engine.
	GetEngine() *gin.Engine
}

func NewManager(cfg *Config) Manager {
	m := &manager{
		cfg:    cfg,
		engine: gin.New(),
	}

	for _, fn := range ginFuncs {
		m.register(fn(cfg))
	}
	m.registerEndpoints()

	return m
}

type manager struct {
	cfg    *Config
	engine *gin.Engine
	lock   sync.RWMutex
}

func (m *manager) GetEngine() *gin.Engine {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.engine
}

func (m *manager) register(fn gin.HandlerFunc) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.engine.Use(fn)
}

func (m *manager) registerEndpoints() {

	if m.cfg.EnablePprof {
		pprof.Register(m.engine)
	}

	if m.cfg.EnableMetrics {
		m.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	if m.cfg.EnableHealthCheck {
		m.engine.GET("/health", func(c *gin.Context) {
			c.String(200, "ok")
		})
	}

}
