package httpcli

import (
	"context"
	"io"
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
		params := map[string]string{
			"q": "test",
		}

		resp, err := hc.Get(context.Background(), url, header, params)
		if err != nil {
			t.Errorf("error sending GET request: %v", err)
			return
		}

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("error reading response body: %v", err)
			return
		}

		t.Logf("response content: %s", content)
		time.Sleep(time.Second)

	})
}
