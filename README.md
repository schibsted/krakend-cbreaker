# krakend-cbreaker

Krankend cirbuit breaker implementation based on [github.com/afex/hystrix-go](github.com/afex/hystrix-go)

## Implementation
It's based on a new backend factory which can be used by any ProxyFactory

## Usage
See an usage example [here](./proxy_integration_test.go)

```go

import (
	"log"
	"os"
	"testing"
	"time"

  cbreaker "github.com/schibsted/krakend-cbreaker"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
  kgin "github.com/devopsfaith/krakend/router/gin"
)

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

	routerFactory := kgin.DefaultFactory(proxy.NewDefaultFactory(cbreaker.BackendFactory(proxy.CustomHTTPProxyFactory(proxy.NewHTTPClient)), logger), logger)
	routerFactory.New().Run(serviceConfig)
}
```

## Configuration
The configuration should be added in the backend extra_config.

### Options
- **command_name:** 
it's the hystrix command name

- **sleep_window:**
This property sets the amount of time, after tripping the circuit, to reject requests before allowing attempts again to determine if the circuit should again be closed.

- **max_concurrent_requests:**
This property sets the maximum number of requests allowed to a HystrixCommand.run() method when you are using ExecutionIsolationStrategy.SEMAPHORE.
If this maximum concurrent limit is hit then subsequent requests will be rejected.

- **error_percent_threshold:**
This property sets the error percentage at or above which the circuit should trip open and start short-circuiting requests to fallback logic.

- **request_volume_threshold:**
This property sets the minimum number of requests in a rolling window that will trip the circuit.
For example, if the value is 20, then if only 19 requests are received in the rolling window (say a window of 10 seconds) the circuit will not trip open even if all 19 failed.

- **timeout:**
This property sets the time in milliseconds after which the caller will observe a timeout and walk away from the command execution. Hystrix marks the HystrixCommand as a TIMEOUT, and performs fallback logic. Note that there is configuration for turning off timeouts per-command, if that is desired (see command.timeout.enabled)

### More detail
For a more detailed description please, visit: [https://github.com/Netflix/Hystrix/wiki/Configuration](https://github.com/Netflix/Hystrix/wiki/Configuration)

### Example
```yml
{
  "version": 2,
  "max_idle_connections": 250,
  "timeout": "3000ms",
  "read_timeout": "0s",
  "write_timeout": "0s",
  "idle_timeout": "0s",
  "read_header_timeout": "0s",
  "name": "Test",
  "endpoints": [
    {
      "endpoint": "/cbcrash",
      "method": "GET",
      "backend": [
        {
          "url_pattern": "/crash",
          "host": [
            "http://localhost:8000"
          ],
          "extra_config": {
            "github.com/schibsted/krakend-cbreaker": {
              "command_name": "crash",
              "sleep_window": 10000.0,
              "max_concurrent_requests": 1.0,
              "error_percent_threshold": 1.0,
              "request_volume_threshold": 1.0,
              "timeout": 1000.0
            }
          }
        }
      ],
      "timeout": "1500ms",
      "max_rate": "10000"
    }
  ]
}
```


