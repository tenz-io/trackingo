package httpcli

import (
	"bytes"
	"context"
	"fmt"
	"github.com/tenz-io/trackingo/common"
	"github.com/tenz-io/trackingo/logger"
	"github.com/tenz-io/trackingo/monitor"
	"github.com/tenz-io/trackingo/util"
	"io"
	"net/http"
	"time"
)

type Client interface {
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
	cli *http.Client,
	opts Options,
) Client {
	hc := &client{
		cli: cli,
	}

	for _, opt := range opts {
		opt(hc)
	}
	return hc
}

type client struct {
	cli           *http.Client
	enableMetrics bool
	enableTrace   bool
}

type Option func(*client)
type Options []Option

func WithMetrics(enable bool) Option {
	return func(hc *client) {
		hc.enableMetrics = enable
	}
}

func WithTrace(enable bool) Option {
	return func(hc *client) {
		hc.enableTrace = enable
	}
}

func (c *client) Head(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	req, err := c.newRequest(ctx, http.MethodHead, url, nil, headers, params)
	if err != nil {
		return nil, err
	}

	return c.Request(ctx, req)
}

func (c *client) Get(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	req, err := c.newRequest(ctx, http.MethodGet, url, nil, headers, params)
	if err != nil {
		return nil, err
	}

	return c.Request(ctx, req)
}

func (c *client) Patch(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	req, err := c.newRequest(ctx, http.MethodPatch, url, nil, headers, params)
	if err != nil {
		return nil, err
	}

	return c.Request(ctx, req)
}

func (c *client) Delete(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil, headers, params)
	if err != nil {
		return nil, err
	}

	return c.Request(ctx, req)
}

func (c *client) Post(ctx context.Context, url string, headers, params map[string]string, body io.Reader) (resp *http.Response, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, url, body, headers, params)
	if err != nil {
		return nil, err
	}

	return c.Request(ctx, req)
}

func (c *client) Put(ctx context.Context, url string, headers, params map[string]string, body io.Reader) (resp *http.Response, err error) {
	req, err := c.newRequest(ctx, http.MethodPut, url, body, headers, params)
	if err != nil {
		return nil, err
	}

	return c.Request(ctx, req)
}

func (c *client) PostJson(ctx context.Context, url string, headers, params map[string]string, jsonBody []byte) (respContent []byte, err error) {
	body := bytes.NewBuffer(jsonBody)
	resp, err := c.Post(ctx, url, headers, params, body)
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

func (c *client) Request(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	var (
		begin   = time.Now()
		code    = 0
		path    = util.If(req.URL.Path != "", req.URL.Path, "/")
		reqBody = util.CaptureRequest(req)
		rec     = monitor.BeginRecord(ctx, path)
	)

	defer func() {
		code = common.ErrorCode(err)
		respBody := util.CaptureResponse(resp)
		if c.enableMetrics {
			rec.EndWithCode(code)
			mon := monitor.GetSingleFlight(ctx)
			mon.Sample(ctx, path, code, float64(len(reqBody)), "reqLen")
			if resp != nil {
				mon.Sample(ctx, path, code, float64(len(respBody)), "respLen")
			}

		}
		if c.enableTrace {
			logger.TrafficEntryFromContext(ctx).DataWith(&logger.Traffic{
				Typ:  logger.TrafficTypRequest,
				Cmd:  path,
				Cost: time.Since(begin),
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
				Req:  util.ReadableHttpBody(util.RequestContentType(req), reqBody),
				Resp: util.ReadableHttpBody(util.ResponseContentType(resp), respBody),
			}, logger.Fields{
				"method":     req.Method,
				"url":        req.URL.String(),
				"reqHeader":  req.Header,
				"reqQuery":   req.URL.Query(),
				"reqLen":     len(reqBody),
				"respCode":   getRespField(resp, func(resp *http.Response) any { return resp.StatusCode }),
				"respHeader": getRespField(resp, func(resp *http.Response) any { return resp.Header }),
				"respLen":    len(respBody),
			})
		}
	}()

	resp, err = c.cli.Do(req)
	if err != nil {
		return resp, common.NewValError(1, fmt.Errorf("error sending request: %w", err))
	}

	if resp.StatusCode != http.StatusOK {
		return resp, common.NewValError(code, fmt.Errorf("response with status: %d", resp.StatusCode))
	}

	return resp, nil
}

func (c *client) newRequest(ctx context.Context,
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
