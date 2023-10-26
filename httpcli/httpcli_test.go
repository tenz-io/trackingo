package httpcli

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewHttpClient(t *testing.T) {
	t.Run("test new http cli", func(t *testing.T) {
		hc := NewHttpClient(&http.Client{}, Options{
			WithMetrics(true),
			WithTrace(true),
		})

		url := "https://www.google.com"
		header := map[string]string{}
		params := map[string][]string{
			"q": {"golang"},
		}

		respBody, err := hc.Get(context.Background(), url, params, header)
		if err != nil {
			t.Errorf("error sending GET request: %v", err)
			return
		}

		t.Logf("response content: %s", respBody)
		time.Sleep(time.Second)

	})
}
