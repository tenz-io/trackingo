package httpgin

type Config struct {
	EnableAccess      bool   `yaml:"enable_access" json:"enable_access" default:"true"`
	AccessLog         string `yaml:"access_log" json:"access_log" default:"log"`
	EnablePprof       bool   `yaml:"enable_pprof" json:"enable_pprof" default:"true"`
	EnableMetrics     bool   `yaml:"enable_metrics" json:"enable_metrics" default:"true"`
	EnableTraffic     bool   `yaml:"enable_traffic" json:"enable_traffic" default:"true"`
	EnableHealthCheck bool   `yaml:"enable_health_check" json:"enable_health_check" default:"true"`
	Timeout           int    `yaml:"timeout" json:"timeout" default:"0"` // seconds, 0 means no timeout
}
