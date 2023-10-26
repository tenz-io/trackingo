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

type Params map[string][]string
type Headers map[string]string

type Client interface {
	// Request sends an HTTP request and returns an HTTP response, following
	Request(ctx context.Context, req *http.Request) (resp *http.Response, err error)
	// Head sends a HEAD request and returns the response.
	Head(ctx context.Context, url string, params Params, headers Headers) (err error)
	// Delete sends a DELETE request and returns the response.
	Delete(ctx context.Context, url string, params Params, headers Headers) (err error)
	// Get sends a GET request and returns the response body as a byte slice.
	Get(ctx context.Context, url string, params Params, headers Headers) (respBody []byte, err error)
	// Post sends a POST request and returns the response body as a byte slice.
	Post(ctx context.Context, url string, params Params, headers Headers, reqBody []byte) (respBody []byte, err error)
	// Put sends a PUT request and returns the response body as a byte slice.
	Put(ctx context.Context, url string, params Params, headers Headers, reqBody []byte) (respBody []byte, err error)
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

func (c *client) Head(
	ctx context.Context,
	url string,
	params Params,
	headers Headers,
) (err error) {
	req, err := c.newRequest(ctx, http.MethodHead, url, params, headers, nil)
	if err != nil {
		return err
	}

	_, err = c.Request(ctx, req)
	return err
}

func (c *client) Delete(
	ctx context.Context,
	url string,
	params Params,
	headers Headers,
) (err error) {
	req, err := c.newRequest(ctx, http.MethodDelete, url, params, headers, nil)
	if err != nil {
		return err
	}

	_, err = c.Request(ctx, req)
	return err
}

func (c *client) Get(
	ctx context.Context,
	url string,
	params Params,
	headers Headers,
) (respBody []byte, err error) {
	req, err := c.newRequest(ctx, http.MethodGet, url, params, headers, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Request(ctx, req)
	if err != nil {
		return nil, err
	}

	return c.readResponseBody(resp)
}

func (c *client) Post(
	ctx context.Context,
	url string,
	params Params,
	headers Headers,
	reqBody []byte,
) (respBody []byte, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, url, params, headers, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	resp, err := c.Request(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.readResponseBody(resp)
}

func (c *client) Put(
	ctx context.Context,
	url string,
	params Params,
	headers Headers,
	reqBody []byte,
) (respBody []byte, err error) {
	req, err := c.newRequest(ctx, http.MethodPut, url, params, headers, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	resp, err := c.Request(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.readResponseBody(resp)
}

func (c *client) Request(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	var (
		begin   = time.Now()
		code    = 0
		cmd     = req.Method
		reqBody = util.CaptureRequest(req)
		rec     = monitor.BeginRecord(ctx, cmd)
	)

	defer func() {
		code = common.ErrorCode(err)
		respBody := util.CaptureResponse(resp)
		if c.enableMetrics {
			rec.EndWithCode(code)
			mon := monitor.GetSingleFlight(ctx)
			mon.Sample(ctx, cmd, code, float64(len(reqBody)), "reqLen")
			if resp != nil {
				mon.Sample(ctx, cmd, code, float64(len(respBody)), "respLen")
			}

		}
		if c.enableTrace {
			logger.TrafficEntryFromContext(ctx).DataWith(&logger.Traffic{
				Typ:  logger.TrafficTypRequest,
				Cmd:  cmd,
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
	method string,
	url string,
	params Params,
	headers Headers,
	body io.Reader,
) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating %s request: %w", method, err)
	}

	if len(params) > 0 {
		q := req.URL.Query()
		for k, vars := range params {
			for _, v := range vars {
				q.Add(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return req, nil
}

func (c *client) readResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return bs, nil
}

func getRespField(resp *http.Response, fn func(*http.Response) any) any {
	if resp == nil {
		return nil
	}

	return fn(resp)
}
