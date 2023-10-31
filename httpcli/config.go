package httpcli

import "time"

type Config struct {
	MaxConnsPerHost int           `yaml:"max_conns_per_host" json:"max_conns_per_host" default:"100"`
	IdleConnTimeout time.Duration `yaml:"idle_conn_timeout" json:"idle_conn_timeout" default:"60s"`
	ReadBufferSize  int           `yaml:"read_buffer_size" json:"read_buffer_size" default:"4096"`
	MaxTimeout      time.Duration `yaml:"max_timeout" json:"max_timeout" default:"120s"`
	EnableMetrics   bool          `yaml:"enable_metrics" json:"enable_metrics" default:"true"`
	EnableTraffic   bool          `yaml:"enable_traffic" json:"enable_traffic" default:"true"`
}
