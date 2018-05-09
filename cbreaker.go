//
// Copyright 2011 - 2018 Schibsted Products & Technology AS.
// Licensed under the terms of the Apache 2.0 license. See LICENSE in the project root.
//

package cbreaker

import (
	"github.com/afex/hystrix-go/hystrix"
	"github.com/devopsfaith/krakend/config"
)

// Namespace is the key to use to store and access the custom config data
const Namespace = "github.com/schibsted/krakend-cbreaker"

// Config is the custom config struct containing the params for the sony/gobreaker package
type Config struct {
	CommandName            string
	Timeout                int
	SleepWindow            int
	MaxConcurrentRequests  int
	ErrorPercentThreshold  int
	RequestVolumeThreshold int
}

// ZeroCfg is the zero value for the Config struct
var ZeroCfg = Config{}

// ConfigGetter implements the config.ConfigGetter interface. It parses the extra config for the
// gobreaker adapter and returns a ZeroCfg if something goes wrong.
func ConfigGetter(e config.ExtraConfig) interface{} {
	v, ok := e[Namespace]
	if !ok {
		return ZeroCfg
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return ZeroCfg
	}

	cfg := Config{}
	if v, ok := tmp["command_name"]; ok {
		cfg.CommandName = v.(string)
	}
	if v, ok := tmp["timeout"]; ok {
		cfg.Timeout = int(v.(float64))
	}
	if v, ok := tmp["max_concurrent_requests"]; ok {
		cfg.MaxConcurrentRequests = int(v.(float64))
	}
	if v, ok := tmp["error_percent_threshold"]; ok {
		cfg.ErrorPercentThreshold = int(v.(float64))
	}
	if v, ok := tmp["request_volume_threshold"]; ok {
		cfg.RequestVolumeThreshold = int(v.(float64))
	}
	if v, ok := tmp["sleep_window"]; ok {
		cfg.SleepWindow = int(v.(float64))
	}
	return cfg
}

// NewCommand builds a gobreaker circuit breaker with the injected config
func NewCommand(cfg Config) *HystrixCommand {
	hystrixConfig := hystrix.CommandConfig{
		Timeout:                cfg.Timeout,
		MaxConcurrentRequests:  cfg.MaxConcurrentRequests,
		ErrorPercentThreshold:  cfg.ErrorPercentThreshold,
		RequestVolumeThreshold: cfg.RequestVolumeThreshold,
	}
	hystrix.ConfigureCommand(cfg.CommandName, hystrixConfig)

	return &HystrixCommand{name: cfg.CommandName, cfg: hystrixConfig}
}

type HystrixCommand struct {
	name string
	cfg  hystrix.CommandConfig
}

func (hc *HystrixCommand) Execute(cmd func() error, fallback func(error) error) (err error) {
	return hystrix.Do(hc.name, cmd, fallback)
}
