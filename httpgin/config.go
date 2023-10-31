package httpgin

type Config struct {
	EnableAccess    bool   `yaml:"enable_access" json:"enable_access" default:"true"`
	AccessLogbase   string `yaml:"access_log" json:"access_logbase" default:"log"`
	EnablePprof     bool   `yaml:"enable_pprof" json:"enable_pprof" default:"true"`
	EnableMetrics   bool   `yaml:"enable_metrics" json:"enable_metrics" default:"true"`
	MetricsEndpoint string `yaml:"metrics_endpoint" json:"metrics_endpoint" default:"/metrics"`
	EnableTraffic   bool   `yaml:"enable_traffic" json:"enable_traffic" default:"true"`
	EnableCheck     bool   `yaml:"enable_check" json:"enable_check" default:"true"`
	CheckEndpoint   string `yaml:"check_endpoint" json:"check_endpoint" default:"/health"`
	Timeout         int    `yaml:"timeout" json:"timeout" default:"0"` // seconds, 0 means no timeout
}
