package httpcli

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func Test_client_Request(t *testing.T) {
	type fields struct {
		sender        sender
		enableMetrics bool
		enableTraffic bool
	}
	type behavior func(fields)
	type args struct {
		ctx context.Context
		req *http.Request
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		behavior behavior
		wantResp *http.Response
		wantErr  bool
	}{
		{
			name: "when sender.Do returns error then return error",
			fields: fields{
				sender:        new(mockSender),
				enableMetrics: true,
				enableTraffic: true,
			},
			behavior: func(fields fields) {
				var (
					senderMock = fields.sender.(*mockSender)
				)

				senderMock.On("Do", mock.Anything).Return(
					nil,
					fmt.Errorf("some error"),
				).Once()
			},
			args: args{
				ctx: func() context.Context {
					return context.Background()
				}(),
				req: &http.Request{
					Method: http.MethodPost,
					URL:    &url.URL{},
					Body:   http.NoBody,
				},
			},
			wantResp: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				sender:        tt.fields.sender,
				enableMetrics: true,
				enableTraffic: true,
			}

			tt.behavior(tt.fields)

			gotResp, err := c.Request(tt.args.ctx, tt.args.req)
			t.Logf("gotResp: %v", gotResp)
			t.Logf("err: %v", err)
			if (err != nil) != tt.wantErr {
				t.Errorf("Request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("Request() gotResp = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}
