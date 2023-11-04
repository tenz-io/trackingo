package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync"
	"time"
)

type item struct {
	raw    []byte
	expire int64
}

type local struct {
	m    map[string]*item
	lock sync.RWMutex
}

func NewLocal() Manager {
	return &local{
		m: make(map[string]*item),
	}
}

func (l *local) active() bool {
	if l == nil || l.m == nil {
		return false
	}
	return true
}

func (l *local) Get(ctx context.Context, key string) (raw string, err error) {
	if !l.active() {
		return "", ErrInActive
	}

	l.lock.RLock()

	it, found := l.m[key]
	if !found {
		defer l.lock.RUnlock()
		return "", ErrNotFound
	}

	if it == nil || time.Now().Unix() > it.expire {
		l.lock.RUnlock()

		l.lock.Lock()
		defer l.lock.Unlock()
		delete(l.m, key)
		return "", ErrNotFound
	} else {
		defer l.lock.RUnlock()
		return string(it.raw), nil
	}

}

func (l *local) Set(ctx context.Context, key string, raw string, expire time.Duration) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	l.m[key] = &item{
		raw:    []byte(raw),
		expire: time.Now().Add(expire).Unix(),
	}
	return nil
}

func (l *local) SetNx(ctx context.Context, key string, raw string, expire time.Duration) (existing bool, err error) {
	if !l.active() {
		return false, ErrInActive
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if _, ok := l.m[key]; ok {
		return true, nil
	} else {
		l.m[key] = &item{
			raw:    []byte(raw),
			expire: time.Now().Add(expire).Unix(),
		}
		return false, nil
	}
}

func (l *local) GetBlob(ctx context.Context, key string, output any) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.lock.RLock()
	it, found := l.m[key]
	if !found {
		defer l.lock.RUnlock()
		return ErrNotFound
	}

	if it != nil && time.Now().Unix() < it.expire {
		defer l.lock.RUnlock()

		r := bytes.NewReader(it.raw)
		decoder := gob.NewDecoder(r)
		if err = decoder.Decode(output); err != nil {
			return fmt.Errorf("decode error: %w", err)
		}
		return nil
	} else {
		l.lock.RUnlock()

		l.lock.Lock()
		defer l.lock.Unlock()
		delete(l.m, key)
		return ErrNotFound
	}

}

func (l *local) SetBlob(ctx context.Context, key string, val any, expire time.Duration) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err = encoder.Encode(val); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	l.m[key] = &item{
		raw:    buf.Bytes(),
		expire: time.Now().Add(expire).Unix(),
	}
	return nil

}

func (l *local) Del(ctx context.Context, key string) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if _, ok := l.m[key]; ok {
		delete(l.m, key)
	}
	return nil
}

func (l *local) Expire(ctx context.Context, key string, expire time.Duration) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.lock.Lock()
	defer l.lock.Unlock()
	if it, ok := l.m[key]; ok && it != nil {
		it.expire = time.Now().Add(expire).Unix()
		return nil
	} else {
		return ErrNotFound
	}

}

func (l *local) Eval(ctx context.Context, script string, keys []string, args ...any) (val any, err error) {
	// ignore
	return nil, fmt.Errorf("not support")
}
