package httpcli

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"net/http"
	"reflect"
	"testing"
)

func Test_client_Request(t *testing.T) {
	type fields struct {
		sender Sender
		cfg    *Config
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
				sender: new(MockSender),
				cfg:    &Config{},
			},
			behavior: func(fields fields) {
				var (
					senderMock = fields.sender.(*MockSender)
				)

				senderMock.On("Do", mock.Anything).Return(
					nil,
					fmt.Errorf("some error"),
				)
			},
			args: args{
				ctx: func() context.Context {
					return context.Background()
				}(),
				req: &http.Request{},
			},
			wantResp: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				sender: tt.fields.sender,
				cfg:    tt.fields.cfg,
			}
			gotResp, err := c.Request(tt.args.ctx, tt.args.req)
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
