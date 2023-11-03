package httpgin

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func Test_applyTimeout(t *testing.T) {
	t.Run("when timeout is 0 then return non nil", func(t *testing.T) {
		cfg := &Config{
			Timeout: 0,
		}
		got := applyTimeout(cfg)
		if got == nil {
			t.Errorf("applyTimeout() = %v, want not nil", got)
		}
	})
	t.Run("when timeout is 0 then return nil", func(t *testing.T) {
		cfg := &Config{
			Timeout: 0,
		}
		timeFunc := applyTimeout(cfg)
		if timeFunc == nil {
			t.Errorf("applyTimeout() = %v, want not nil", timeFunc)
		}

		c := &gin.Context{}
		timeFunc(c)

	})
}
