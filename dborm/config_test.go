package dborm

import (
	"testing"
	"time"
)

func TestConfig_GetDSN(t *testing.T) {
	type fields struct {
		Username       string
		Password       string
		Dbname         string
		Host           string
		Port           int
		MaxOpenConn    int
		MaxIdleConn    int
		MaxLifetime    time.Duration
		EnableTracking bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "when all fields are set then return dsn",
			fields: fields{
				Username:       "username",
				Password:       "password",
				Dbname:         "dbname",
				Host:           "host",
				Port:           1234,
				MaxOpenConn:    10,
				MaxIdleConn:    5,
				MaxLifetime:    300 * time.Second,
				EnableTracking: true,
			},
			want: "username:password@tcp(host:1234)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &Config{
				Username:       tt.fields.Username,
				Password:       tt.fields.Password,
				Dbname:         tt.fields.Dbname,
				Host:           tt.fields.Host,
				Port:           tt.fields.Port,
				MaxOpenConn:    tt.fields.MaxOpenConn,
				MaxIdleConn:    tt.fields.MaxIdleConn,
				MaxLifetime:    tt.fields.MaxLifetime,
				EnableTracking: tt.fields.EnableTracking,
			}
			if got := dc.GetDSN(); got != tt.want {
				t.Errorf("GetDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}
