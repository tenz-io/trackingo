package httpcli

import (
	"bytes"
	"context"
	"fmt"
	"github.com/tenz-io/trackingo/common"
	"github.com/tenz-io/trackingo/logger"
	"github.com/tenz-io/trackingo/monitor"
	"io"
	"net/http"
	"time"
)

type HttpClient interface {
	Request(ctx context.Context, req *http.Request) (resp *http.Response, err error)
	Head(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error)
	Get(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error)
	Patch(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error)
	Post(ctx context.Context, url string, headers, params map[string]string, body io.Reader) (resp *http.Response, err error)
	Put(ctx context.Context, url string, headers, params map[string]string, body io.Reader) (resp *http.Response, err error)
	Delete(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error)
	PostJson(ctx context.Context, url string, headers, params map[string]string, jsonBody []byte) (respContent []byte, err error)
}

func NewHttpClient(
	client *http.Client,
	opts ...Option,
) HttpClient {
	hc := &httpClient{
		client: client,
	}

	for _, opt := range opts {
		opt(hc)
	}
	return hc
}

type httpClient struct {
	client        *http.Client
	enableMetrics bool
	enableTrace   bool
}

type Option func(*httpClient)

func WithMetrics(enable bool) Option {
	return func(hc *httpClient) {
		hc.enableMetrics = enable
	}
}

func WithTrace(enable bool) Option {
	return func(hc *httpClient) {
		hc.enableTrace = enable
	}
}

func (hc *httpClient) Head(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	req, err := hc.newRequest(ctx, http.MethodHead, url, nil, headers, params)
	if err != nil {
		return nil, err
	}

	return hc.Request(ctx, req)
}

func (hc *httpClient) Get(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	req, err := hc.newRequest(ctx, http.MethodGet, url, nil, headers, params)
	if err != nil {
		return nil, err
	}

	return hc.Request(ctx, req)
}

func (hc *httpClient) Patch(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	req, err := hc.newRequest(ctx, http.MethodPatch, url, nil, headers, params)
	if err != nil {
		return nil, err
	}

	return hc.Request(ctx, req)
}

func (hc *httpClient) Delete(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	req, err := hc.newRequest(ctx, http.MethodDelete, url, nil, headers, params)
	if err != nil {
		return nil, err
	}

	return hc.Request(ctx, req)
}

func (hc *httpClient) Post(ctx context.Context, url string, headers, params map[string]string, body io.Reader) (resp *http.Response, err error) {
	req, err := hc.newRequest(ctx, http.MethodPost, url, body, headers, params)
	if err != nil {
		return nil, err
	}

	return hc.Request(ctx, req)
}

func (hc *httpClient) Put(ctx context.Context, url string, headers, params map[string]string, body io.Reader) (resp *http.Response, err error) {
	req, err := hc.newRequest(ctx, http.MethodPut, url, body, headers, params)
	if err != nil {
		return nil, err
	}

	return hc.Request(ctx, req)
}

func (hc *httpClient) PostJson(ctx context.Context, url string, headers, params map[string]string, jsonBody []byte) (respContent []byte, err error) {
	body := bytes.NewBuffer(jsonBody)
	resp, err := hc.Post(ctx, url, headers, params, body)
	if err != nil {
		return nil, fmt.Errorf("error sending POST request: %w", err)
	}
	//defer func() {
	//	_ = resp.Body.Close()
	//}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response with status: %d", resp.StatusCode)
	}

	respContent, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return respContent, nil
}

func (hc *httpClient) Request(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	var (
		begin = time.Now()
		code  = 0
		path  = req.URL.Path
		rec   = monitor.BeginRecord(ctx, path)
	)

	defer func() {
		code = common.ErrorCode(err)
		if hc.enableMetrics {
			rec.EndWithCode(code)
			mon := monitor.GetSingleFlight(ctx)
			mon.Sample(ctx, path, code, float64(req.ContentLength), "reqLen")
			if resp != nil {
				mon.Sample(ctx, path, code, float64(resp.ContentLength), "respLen")
			}

		}
		if hc.enableTrace {
			logger.TrafficEntryFromContext(ctx).DataWith(&logger.Traffic{
				Typ:  logger.TrafficTypRequest,
				Cmd:  path,
				Cost: time.Since(begin),
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
			}, logger.Fields{
				"method":     req.Method,
				"host":       req.Host,
				"path":       path,
				"reqHeader":  req.Header,
				"reqQuery":   req.URL.Query(),
				"reqLen":     req.ContentLength,
				"respCode":   getRespField(resp, func(resp *http.Response) any { return resp.StatusCode }),
				"respHeader": getRespField(resp, func(resp *http.Response) any { return resp.Header }),
				"respLen":    getRespField(resp, func(resp *http.Response) any { return resp.ContentLength }),
			})
		}
	}()

	resp, err = hc.client.Do(req)
	if err != nil {
		return resp, common.NewValError(1, fmt.Errorf("error sending request: %w", err))
	}

	if code != http.StatusOK {
		return resp, common.NewValError(code, fmt.Errorf("response with status: %d", code))
	}

	return resp, nil
}

func (hc *httpClient) newRequest(ctx context.Context,
	method,
	url string,
	body io.Reader,
	headers,
	params map[string]string,
) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating %s request: %w", method, err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

func getRespField(resp *http.Response, fn func(*http.Response) any) any {
	if resp == nil {
		return nil
	}

	return fn(resp)
}
