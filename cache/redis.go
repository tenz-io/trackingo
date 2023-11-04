package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/tenz-io/trackingo/common"
	"github.com/tenz-io/trackingo/logger"
	"github.com/tenz-io/trackingo/monitor"
	"time"
)

type Opt func(m *manager)
type Options []Opt

func NewManager(
	client *redis.Client,
	opts Options,
) Manager {
	m := &manager{
		client: client,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

type manager struct {
	client        *redis.Client
	enableMetrics bool
	enableTraffic bool
}

func WithMetrics(enable bool) Opt {
	return func(m *manager) {
		m.enableMetrics = enable
	}
}

func WithTraffic(enable bool) Opt {
	return func(m *manager) {
		m.enableTraffic = enable
	}
}

func (m *manager) active() bool {
	if m == nil || m.client == nil {
		return false
	}
	return true
}

func (m *manager) Get(ctx context.Context, key string) (raw string, err error) {
	if m.enableMetrics {
		rec := monitor.BeginRecord(ctx, "cache_get")
		defer func() {
			rec.EndWithError(err)
		}()
	}

	if m.enableTraffic {
		trafficRec := logger.StartTrafficRec(ctx, &logger.TrafficReq{
			Cmd: "cache_get",
			Req: key,
		}, logger.Fields{})
		defer func() {
			trafficRec.End(&logger.TrafficResp{
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
				Resp: raw,
			}, logger.Fields{})
		}()
	}

	if !m.active() {
		return "", ErrInActive
	}
	raw, err = m.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound
		}
		return "", err
	}

	return raw, nil
}

func (m *manager) Set(ctx context.Context, key string, raw string, expire time.Duration) (err error) {

	if m.enableMetrics {
		rec := monitor.BeginRecord(ctx, "cache_set")
		defer func() {
			rec.EndWithError(err)
		}()
	}

	if m.enableTraffic {
		trafficRec := logger.StartTrafficRec(ctx, &logger.TrafficReq{
			Cmd: "cache_set",
			Req: key,
		}, logger.Fields{
			"expire": fmt.Errorf("%v", expire),
		})
		defer func() {
			trafficRec.End(&logger.TrafficResp{
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
				Resp: raw,
			}, logger.Fields{})
		}()
	}

	if !m.active() {
		return ErrInActive
	}

	err = m.client.Set(ctx, key, raw, expire).Err()
	return
}

func (m *manager) SetNx(ctx context.Context, key string, raw string, expire time.Duration) (existing bool, err error) {

	if m.enableMetrics {
		rec := monitor.BeginRecord(ctx, "cache_setnx")
		defer func() {
			rec.EndWithError(err)
		}()
	}

	if m.enableTraffic {
		trafficRec := logger.StartTrafficRec(ctx, &logger.TrafficReq{
			Cmd: "cache_setnx",
			Req: key,
		}, logger.Fields{
			"expire": fmt.Errorf("%v", expire),
		})
		defer func() {
			trafficRec.End(&logger.TrafficResp{
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
				Resp: raw,
			}, logger.Fields{
				"existing": existing,
			})
		}()
	}

	if !m.active() {
		return false, ErrInActive
	}

	existing, err = m.client.SetNX(ctx, key, raw, expire).Result()
	return
}

func (m *manager) GetBlob(ctx context.Context, key string, output any) (err error) {
	if m.enableMetrics {
		rec := monitor.BeginRecord(ctx, "cache_get_blob")
		defer func() {
			rec.EndWithError(err)
		}()
	}

	if m.enableTraffic {
		trafficRec := logger.StartTrafficRec(ctx, &logger.TrafficReq{
			Cmd: "cache_get_blob",
			Req: key,
		}, logger.Fields{})
		defer func() {
			trafficRec.End(&logger.TrafficResp{
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
				Resp: output,
			}, logger.Fields{})
		}()
	}

	if !m.active() {
		return ErrInActive
	}

	bs, err := m.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNotFound
		}
		return err
	}

	r := bytes.NewReader(bs)
	decoder := gob.NewDecoder(r)
	if err = decoder.Decode(output); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

func (m *manager) SetBlob(ctx context.Context, key string, val any, expire time.Duration) (err error) {
	if m.enableMetrics {
		rec := monitor.BeginRecord(ctx, "cache_set_blob")
		defer func() {
			rec.EndWithError(err)
		}()
	}

	if m.enableTraffic {
		trafficRec := logger.StartTrafficRec(ctx, &logger.TrafficReq{
			Cmd: "cache_set_blob",
			Req: key,
		}, logger.Fields{
			"expire": fmt.Errorf("%v", expire),
		})
		defer func() {
			trafficRec.End(&logger.TrafficResp{
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
				Resp: val,
			}, logger.Fields{})
		}()
	}

	if !m.active() {
		return ErrInActive
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err = encoder.Encode(val); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	// expire is 0, then set no expire
	// expire is -1, then set default expire
	if err = m.client.Set(ctx, key, buf.Bytes(), expire).Err(); err != nil {
		return fmt.Errorf("set error: %w", err)
	}
	return nil

}

func (m *manager) Del(ctx context.Context, key string) (err error) {
	if m.enableMetrics {
		rec := monitor.BeginRecord(ctx, "cache_del")
		defer func() {
			rec.EndWithError(err)
		}()
	}

	if m.enableTraffic {
		trafficRec := logger.StartTrafficRec(ctx, &logger.TrafficReq{
			Cmd: "cache_del",
			Req: key,
		}, logger.Fields{})
		defer func() {
			trafficRec.End(&logger.TrafficResp{
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
			}, logger.Fields{})
		}()
	}

	if !m.active() {
		return ErrInActive
	}

	err = m.client.Del(ctx, key).Err()
	return
}

func (m *manager) Expire(ctx context.Context, key string, expire time.Duration) (err error) {
	if m.enableMetrics {
		rec := monitor.BeginRecord(ctx, "cache_expire")
		defer func() {
			rec.EndWithError(err)
		}()
	}

	if m.enableTraffic {
		trafficRec := logger.StartTrafficRec(ctx, &logger.TrafficReq{
			Cmd: "cache_expire",
			Req: key,
		}, logger.Fields{
			"expire": fmt.Errorf("%v", expire),
		})
		defer func() {
			trafficRec.End(&logger.TrafficResp{
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
			}, logger.Fields{})
		}()
	}

	if !m.active() {
		return ErrInActive
	}

	err = m.client.Expire(ctx, key, expire).Err()
	return
}

func (m *manager) Eval(ctx context.Context, script string, keys []string, args ...any) (val any, err error) {
	if m.enableMetrics {
		rec := monitor.BeginRecord(ctx, "cache_eval")
		defer func() {
			rec.EndWithError(err)
		}()
	}

	if m.enableTraffic {
		trafficRec := logger.StartTrafficRec(ctx, &logger.TrafficReq{
			Cmd: "cache_eval",
			Req: script,
		}, logger.Fields{
			"keys": keys,
			"args": args,
		})
		defer func() {
			trafficRec.End(&logger.TrafficResp{
				Code: common.ErrorCode(err),
				Msg:  common.ErrorMsg(err),
				Resp: val,
			}, logger.Fields{})
		}()
	}

	if !m.active() {
		return nil, ErrInActive
	}

	val, err = m.client.Eval(ctx, script, keys, args...).Result()
	return
}
