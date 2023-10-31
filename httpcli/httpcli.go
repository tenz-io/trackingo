package httpcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tenz-io/trackingo/common"
	"github.com/tenz-io/trackingo/logger"
	"github.com/tenz-io/trackingo/monitor"
	"github.com/tenz-io/trackingo/util"
	"io"
	"net/http"
	"strings"
	"time"
)

type (
	Params  map[string][]string
	Headers map[string]string
)

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
	cfg *Config,
) Client {
	return &client{
		cfg: cfg,
		cli: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost: cfg.MaxConnsPerHost,
				IdleConnTimeout: cfg.IdleConnTimeout,
				ReadBufferSize:  cfg.ReadBufferSize,
			},
			Timeout: cfg.MaxTimeout,
		},
	}
}

type client struct {
	cli *http.Client
	cfg *Config
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
		path    = req.URL.Path
		cmd     = util.If(path == "", "/", path)
		reqBody = capture(req.Body)
		code    = 0
		rec     = monitor.BeginRecord(ctx, cmd)
	)

	defer func() {
		code = common.ErrorCode(err)
		respBody := capture(resp.Body)
		if c.cfg.EnableMetrics {
			rec.EndWithCode(code)
			mon := monitor.FromContext(ctx)
			mon.Sample(ctx, cmd, code, float64(len(reqBody)), "reqLen")
			if resp != nil {
				mon.Sample(ctx, cmd, code, float64(len(respBody)), "respLen")
			}

		}
		if c.cfg.EnableTraffic {
			var (
				respHeader http.Header
				respCode   int
			)
			if resp != nil {
				respHeader = resp.Header
				respCode = resp.StatusCode
			}
			logger.TrafficEntryFromContext(ctx).DataWith(&logger.Traffic{
				Typ:  logger.TrafficTypRequest,
				Cmd:  cmd,
				Cost: time.Since(begin),
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
				Req:  printPayload(req.Header, reqBody),
				Resp: printPayload(respHeader, respBody),
			}, logger.Fields{
				"method":     req.Method,
				"reqUrl":     req.URL.String(),
				"reqHeader":  req.Header,
				"reqQuery":   req.URL.Query(),
				"reqLen":     len(reqBody),
				"respCode":   respCode,
				"respHeader": respHeader,
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

// getContentType returns the content type of the http header.
func getContentType(head http.Header) string {
	if head == nil {
		return ""
	}
	return head.Get("Content-Type")
}

func capture(body io.ReadCloser) []byte {
	if body == nil {
		return nil
	}

	bs, err := io.ReadAll(body)
	if err != nil {
		return nil
	}

	_ = body.Close()
	bsCopy := bytes.Clone(bs)
	defer func() {
		body = io.NopCloser(bytes.NewBuffer(bs))
	}()
	return bsCopy
}

// printPayload print the payload of the http request or response.
func printPayload(header http.Header, payload []byte) any {
	contentType := getContentType(header)
	if contentType == "" || len(payload) == 0 {
		return nil
	}

	contentType = strings.ToLower(contentType)

	if !(strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
		strings.HasPrefix(contentType, "text/xml") ||
		strings.HasPrefix(contentType, "text/html")) {
		// if not json, xml, form, html, return nil
		return fmt.Sprintf("<not support contentType: %s>", contentType)
	}

	if strings.HasPrefix(contentType, "application/json") {
		var reqMap map[string]any
		if err := json.Unmarshal(payload, &reqMap); err != nil {
			return nil
		}

		return reqMap
	} else if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		var reqMap map[string]string
		if err := json.Unmarshal(payload, &reqMap); err != nil {
			return nil
		}

		return reqMap
	} else {
		s := string(payload)
		return s
	}

}
