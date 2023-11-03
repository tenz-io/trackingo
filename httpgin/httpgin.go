package httpgin

import (
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ginFunc func(*Config) gin.HandlerFunc

type Manager interface {
	// GetEngine returns the gin.Engine.
	GetEngine() *gin.Engine
	// Use adds middleware to the chain which is run before router.
	Use(gin.HandlerFunc)
	// Run a http server.
	Run(addr ...string) error
}

func NewManager(cfg *Config) Manager {
	m := &manager{
		cfg:    cfg,
		engine: gin.New(),
	}

	for _, fn := range buildInMiddlewares {
		m.Use(fn(cfg))
	}

	return m
}

type manager struct {
	cfg    *Config
	engine *gin.Engine
}

func (m *manager) GetEngine() *gin.Engine {
	return m.engine
}

func (m *manager) Use(fn gin.HandlerFunc) {
	m.engine.Use(fn)
}

func (m *manager) Run(addr ...string) error {
	m.register()

	err := m.engine.Run(addr...)
	if err != nil {
		return fmt.Errorf("failed to run http server: %w", err)
	}
	return nil
}

// register registers the endpoints.
func (m *manager) register() {

	if m.cfg.EnablePprof {
		pprof.Register(m.engine)
	}

	if m.cfg.EnableMetrics {
		if m.cfg.MetricsEndpoint == "" {
			m.cfg.MetricsEndpoint = "/metrics"
		}

		m.engine.GET(m.cfg.MetricsEndpoint, gin.WrapH(promhttp.Handler()))
	}

	if m.cfg.EnableCheck {
		if m.cfg.CheckEndpoint == "" {
			m.cfg.CheckEndpoint = "/health"
		}
		m.engine.GET(m.cfg.CheckEndpoint, func(c *gin.Context) {
			c.String(200, "ok")
		})
	}

}
