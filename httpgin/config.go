package httpgin

type Config struct {
	EnableAccess      bool   `yaml:"enable_access" default:"true"`
	AccessLog         string `yaml:"access_log" default:"log"`
	EnablePprof       bool   `yaml:"enable_pprof" default:"true"`
	EnableMetrics     bool   `yaml:"enable_metrics" default:"true"`
	EnableTraffic     bool   `yaml:"enable_traffic" default:"true"`
	EnableHealthCheck bool   `yaml:"enable_health_check" default:"true"`
	EnableTimeout     bool   `yaml:"enable_timeout" default:"true"`
	Timeout           int    `yaml:"timeout" default:"10"` // seconds
}
