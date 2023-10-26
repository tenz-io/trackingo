package httpcli

import (
	"context"
	"github.com/stretchr/testify/mock"
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

	if args.Get(0) != nil {
		resp = args.Get(0).(*http.Response)
	}

	err = args.Error(1)

	return

}

func (m *MockClient) Head(ctx context.Context, url string, params Params, headers Headers) (err error) {
	args := m.Called(ctx, url, params, headers)

	err = args.Error(0)

	return
}

func (m *MockClient) Delete(ctx context.Context, url string, params Params, headers Headers) (err error) {
	args := m.Called(ctx, url, params, headers)

	err = args.Error(0)

	return
}

func (m *MockClient) Get(ctx context.Context, url string, params Params, headers Headers) (respBody []byte, err error) {
	args := m.Called(ctx, url, params, headers)

	if args.Get(0) != nil {
		respBody = args.Get(0).([]byte)
	}
	err = args.Error(1)

	return
}

func (m *MockClient) Post(ctx context.Context, url string, params Params, headers Headers, reqBody []byte) (respBody []byte, err error) {
	args := m.Called(ctx, url, params, headers, reqBody)

	if args.Get(0) != nil {
		respBody = args.Get(0).([]byte)
	}
	err = args.Error(1)

	return
}

func (m *MockClient) Put(ctx context.Context, url string, params Params, headers Headers, reqBody []byte) (respBody []byte, err error) {
	args := m.Called(ctx, url, params, headers, reqBody)

	if args.Get(0) != nil {
		respBody = args.Get(0).([]byte)
	}
	err = args.Error(1)

	return
}
