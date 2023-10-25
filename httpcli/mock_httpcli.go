package httpcli

import (
	"context"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) InjectBehavior(mfn func(*MockClient)) *MockClient {
	mfn(m)
	return m
}

func (m *MockClient) Request(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return nil, nil
	}

	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockClient) Head(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	args := m.Called(ctx, url, headers, params)

	if args.Get(0) == nil {
		return nil, nil
	}

	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockClient) Get(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	args := m.Called(ctx, url, headers, params)

	if args.Get(0) == nil {
		return nil, nil
	}

	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockClient) Patch(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	args := m.Called(ctx, url, headers, params)

	if args.Get(0) == nil {
		return nil, nil
	}

	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockClient) Post(ctx context.Context, url string, headers, params map[string]string, body io.Reader) (resp *http.Response, err error) {
	args := m.Called(ctx, url, headers, params, body)

	if args.Get(0) == nil {
		return nil, nil
	}

	return args.Get(0).(*http.Response), args.Error(1)

}

func (m *MockClient) Put(ctx context.Context, url string, headers, params map[string]string, body io.Reader) (resp *http.Response, err error) {
	args := m.Called(ctx, url, headers, params, body)

	if args.Get(0) == nil {
		return nil, nil
	}

	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockClient) Delete(ctx context.Context, url string, headers, params map[string]string) (resp *http.Response, err error) {
	args := m.Called(ctx, url, headers, params)

	if args.Get(0) == nil {
		return nil, nil
	}

	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockClient) PostJson(ctx context.Context, url string, headers, params map[string]string, jsonBody []byte) (respContent []byte, err error) {
	args := m.Called(ctx, url, headers, params, jsonBody)

	if args.Get(0) == nil {
		return nil, nil
	}

	return args.Get(0).([]byte), args.Error(1)
}
