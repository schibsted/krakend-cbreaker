//
// Copyright 2011 - 2018 Schibsted Products & Technology AS.
// Licensed under the terms of the Apache 2.0 license. See LICENSE in the project root.
//
package cbreaker

import (
	"context"
	"errors"
	"testing"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"

	"net/http"
)

var Backend500Error error = errors.New("Backend500Error")

func TestNewMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r != proxy.ErrTooManyProxies {
			t.Errorf("The code did not panic\n")
		}
	}()
	NewMiddleware(&config.Backend{})(proxy.NoopProxy, proxy.NoopProxy)
}

func TestNewMiddleware_zeroConfig(t *testing.T) {
	cfg := &config.Backend{
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{
				"command_name":             "test_cmd",
				"sleep_window":             10.0,
				"max_concurrent_requests":  15.0,
				"error_percent_threshold":  50.0,
				"request_volume_threshold": 2.0,
				"timeout":                  10.0,
			},
		},
	}
	data := ConfigGetter(cfg.ExtraConfig).(Config)

	if data.CommandName != "test_cmd" {
		t.Errorf("Comand name was expected to be CommandName, but it's %s", data.CommandName)
	}
	if data.SleepWindow != 10 {
		t.Errorf("SleepWindow was expected to be 10, but it's %i", data.SleepWindow)
	}
	if data.MaxConcurrentRequests != 15 {
		t.Errorf("MaxConcurrentRequests was expected to be 15, but it's %i", data.MaxConcurrentRequests)
	}
	if data.ErrorPercentThreshold != 50 {
		t.Errorf("ErrorPercentThreshold was expected to be 50, but it's %i", data.ErrorPercentThreshold)
	}
	if data.RequestVolumeThreshold != 2 {
		t.Errorf("SleepWindow was expected to be 2, but it's %i", data.RequestVolumeThreshold)
	}
	if data.Timeout != 10 {
		t.Errorf("Timeout was expected to be 10, but it's %i", data.Timeout)
	}
}

func TestNewMiddleware_Config(t *testing.T) {
	for _, cfg := range []*config.Backend{
		{},
		{ExtraConfig: map[string]interface{}{Namespace: 42}},
	} {
		resp := proxy.Response{}
		mdw := NewMiddleware(cfg)
		p := mdw(dummyProxy(&resp, nil))

		request := proxy.Request{
			Path: "/tupu",
		}

		for i := 0; i < 100; i++ {
			r, err := p(context.Background(), &request)
			if err != nil {
				t.Error(err.Error())
				return
			}
			if &resp != r {
				t.Fail()
			}
		}
	}
}

func TestNewMiddleware_ko_with_200(t *testing.T) {
	resp := proxy.Response{Metadata: proxy.Metadata{StatusCode: http.StatusOK}}
	mdw := NewMiddleware(&config.Backend{
		ExtraConfig: map[string]interface{}{
			Namespace: map[interface{}]interface{}{
				"command_name":            "test200",
				"timeout":                 1000.0,
				"max_concurrent_requests": 100.0,
				"error_percent_threshold": 1.0,
			},
		},
	})
	p := mdw(dummyProxy(&resp, nil))

	request := proxy.Request{
		Path: "/tupu",
	}

	for i := 0; i < 100; i++ {
		r, err := p(context.Background(), &request)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if &resp != r {
			t.Fail()
		}

		if r.Metadata.StatusCode != http.StatusOK {
			t.Error("response with status 200 expected")
		}
	}
}

func TestNewMiddleware_ko_with_500(t *testing.T) {
	mdw := NewMiddleware(&config.Backend{
		ExtraConfig: map[string]interface{}{
			Namespace: map[interface{}]interface{}{
				"command_name":             "test500",
				"sleep_window":             1000.0,
				"max_concurrent_requests":  10.0,
				"error_percent_threshold":  1.0,
				"request_volume_threshold": 1.0,
				"timeout":                  1000.0,
			},
		},
	})
	p := mdw(func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, Backend500Error
	})

	request := proxy.Request{
		Path: "/tupu",
	}

	_, actualErr := p(context.Background(), &request)
	if actualErr != Backend500Error {
		t.Error("error expected")
	}

	_, actualErr = p(context.Background(), &request)
	if actualErr != Backend500Error {
		t.Error("error expected")
	}

}

func TestNewMiddleware_ko_with_404(t *testing.T) {
	mdw := NewMiddleware(&config.Backend{
		ExtraConfig: map[string]interface{}{
			Namespace: map[interface{}]interface{}{
				"command_name":             "test404",
				"sleep_window":             1000.0,
				"max_concurrent_requests":  10.0,
				"error_percent_threshold":  1.0,
				"request_volume_threshold": 1.0,
				"timeout":                  1000.0,
			},
		},
	})
	p := mdw(func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{Metadata: proxy.Metadata{StatusCode: http.StatusNotFound}}, nil
	})

	request := proxy.Request{
		Path: "/tupu",
	}

	r, actualErr := p(context.Background(), &request)
	validateProxyResult(r, t, actualErr)

	r, actualErr = p(context.Background(), &request)
	validateProxyResult(r, t, actualErr)

}
func validateProxyResult(r *proxy.Response, t *testing.T, actualErr error) {
	if r == nil {
		t.Error("response expected")
	}
	if actualErr != nil {
		t.Error("error unexpected")
	}
	if r.Metadata.StatusCode != http.StatusNotFound {
		t.Error("response with status 404 expected")
	}
}
