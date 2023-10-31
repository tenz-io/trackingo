package dborm

import "fmt"

type Config struct {
	Username        string `yaml:"username" json:"username"`
	Password        string `yaml:"password" json:"password"`
	Dbname          string `yaml:"dbname" json:"dbname"`
	Host            string `yaml:"host" json:"host"`
	Port            int    `yaml:"port" json:"port"`
	MaxOpenConn     int    `yaml:"max_open_conn" json:"max_open_conn"`
	MaxIdleConn     int    `yaml:"max_idle_conn" json:"max_idle_conn"`
	MaxLifetime     int    `yaml:"max_lifetime" json:"max_lifetime"`
	EnableTracking  bool   `yaml:"enable_tracking" json:"enable_tracking"`
	TrackingLogbase string `yaml:"tracking_logbase" json:"tracking_logbase" default:"log"`
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
