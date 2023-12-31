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
)

type (
	Params  map[string][]string
	Headers map[string]string
)

//go:generate mockery --name sender --filename sender_mock.go --inpackage
type sender interface {
	Do(req *http.Request) (*http.Response, error)
}

type senderImpl struct {
	cli *http.Client
}

func (s *senderImpl) Do(req *http.Request) (*http.Response, error) {
	return s.cli.Do(req)
}

//go:generate mockery --name Client --filename Client_mock.go --inpackage
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

type Opt func(c *client)

type Opts []Opt

func NewClient(
	cli *http.Client,
	opts Opts,
) Client {
	hc := &client{
		sender: &senderImpl{
			cli: cli,
		},
	}

	for _, opt := range opts {
		opt(hc)
	}

	return hc
}

type client struct {
	sender        sender
	enableMetrics bool
	enableTraffic bool
}

func WithMetrics() Opt {
	return func(c *client) {
		c.enableMetrics = true
	}
}

func WithTraffic() Opt {
	return func(c *client) {
		c.enableTraffic = true
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
		path       = req.URL.Path
		cmd        = util.If(path == "", "/", path)
		code       = 0
		respHeader http.Header
		respCode   int
	)

	if c.enableMetrics {
		rec := monitor.BeginRecord(ctx, cmd)
		defer func() {
			rec.EndWithError(err)
		}()
	}

	if c.enableTraffic {
		reqBody := captureRequest(ctx, req)
		trafficRec := logger.StartTrafficRec(ctx, &logger.TrafficReq{
			Cmd: cmd,
			Req: printPayload(req.Header, reqBody),
		}, logger.Fields{
			"method":    req.Method,
			"req_url":   req.URL.String(),
			"header":    req.Header,
			"params":    req.URL.Query(),
			"body_size": len(reqBody),
		})
		defer func() {
			var (
				respBody = captureResponse(ctx, resp)
			)
			trafficRec.End(&logger.TrafficResp{
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
				Resp: printPayload(respHeader, respBody),
			}, logger.Fields{
				"code":      respCode,
				"header":    respHeader,
				"body_size": len(respBody),
			})
		}()
	}

	resp, err = c.sender.Do(req)
	if err != nil {
		return resp, common.NewValError(1, fmt.Errorf("error sending request: %w", err))
	}

	if resp.StatusCode != http.StatusOK {
		return resp, common.NewValError(code, fmt.Errorf("response with status: %d", resp.StatusCode))
	}

	respHeader = resp.Header
	respCode = resp.StatusCode

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

// captureRequest capture http body from http request
func captureRequest(ctx context.Context, req *http.Request) []byte {
	var (
		le = logger.FromContext(ctx)
	)
	if req == nil || req.Body == nil {
		le.Info("request or request body is nil")
		return nil
	}

	bs, err := io.ReadAll(req.Body)
	if err != nil {
		le.WithError(err).Warn("error reading request body")
		return nil
	}

	// clone body for reset body
	bsCopy := bytes.Clone(bs)
	req.Body = io.NopCloser(bytes.NewBuffer(bs))
	return bsCopy
}

// captureResponse capture response from http response
func captureResponse(ctx context.Context, resp *http.Response) []byte {
	var (
		le = logger.FromContext(ctx)
	)
	if resp == nil || resp.Body == nil {
		le.Info("response or response body is nil")
		return nil
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		le.WithError(err).Warn("error reading response body")
		return nil
	}

	// clone body for reset body
	bsCopy := bytes.Clone(bs)
	resp.Body = io.NopCloser(bytes.NewBuffer(bs))
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
