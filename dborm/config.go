package dborm

import (
	"fmt"
	"time"
)

type Config struct {
	Username        string        `yaml:"username" json:"username"`
	Password        string        `yaml:"password" json:"password"`
	Dbname          string        `yaml:"dbname" json:"dbname"`
	Host            string        `yaml:"host" json:"host"`
	Port            int           `yaml:"port" json:"port"`
	MaxOpenConn     int           `yaml:"max_open_conn" json:"max_open_conn" default:"10"`
	MaxIdleConn     int           `yaml:"max_idle_conn" json:"max_idle_conn" default:"5"`
	MaxLifetime     time.Duration `yaml:"max_lifetime" json:"max_lifetime" default:"300s"`
	EnableTracking  bool          `yaml:"enable_tracking" json:"enable_tracking" default:"true"`
	TrackingLogbase string        `yaml:"tracking_logbase" json:"tracking_logbase" default:"log"`
}

func (dc *Config) GetDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dc.Username,
		dc.Password,
		dc.Host,
		dc.Port,
		dc.Dbname,
	)
}
