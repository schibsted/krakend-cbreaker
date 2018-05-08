//
// Copyright 2011 - 2018 Schibsted Products & Technology AS.
// Licensed under the terms of the Apache 2.0 license. See LICENSE in the project root.
//
package cbreaker

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	kgin "github.com/devopsfaith/krakend/router/gin"

	"github.com/gin-gonic/gin"

	"github.com/smartystreets/assertions"
	"github.com/tsenart/vegeta/lib"
)

func setup() {
	go DummyServer()
	go ApiGateway()
	time.Sleep(10 * time.Second)
}

func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	os.Exit(retCode)
}

func ApiGateway() {
	logger, err := logging.NewLogger("INFO", os.Stdout, "[KRAKEND]")
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

	parser := config.NewParser()
	serviceConfig, err := parser.Parse("./test.json")
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

	routerFactory := kgin.DefaultFactory(proxy.NewDefaultFactory(BackendFactory(proxy.CustomHTTPProxyFactory(proxy.NewHTTPClient)), logger), logger)
	routerFactory.New().Run(serviceConfig)
}

func DummyServer() {
	r := gin.Default()
	r.GET("/crash", func(c *gin.Context) {
		c.JSON(500, gin.H{
			"message": "boom!",
		})
	})
	r.Run(":8000")
}

func TestCircuitBreaker(t *testing.T) {
	rate := uint64(4) // per second
	duration := 2 * time.Second
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:8080/cbcrash",
	})
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration) {
		metrics.Add(res)
	}
	metrics.Close()
	equal := assertions.ShouldContainKey(metrics.StatusCodes, "500")
	if equal != "" {
		t.Errorf(equal)
	}
}
