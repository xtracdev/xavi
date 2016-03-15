package loadbalancer

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUp(t *testing.T) {
	ble := &LoadBalancerEndpoint{
		Address: "foo.com",
		PingURI: "/racy",
		Up:      true,
	}

	var wg sync.WaitGroup
	ble.MarkLoadBalancerEndpointUp(true)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			assert.True(t, ble.IsUp())
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestMarkLoadBalancerEndpointUp(t *testing.T) {
	ble := &LoadBalancerEndpoint{
		Address: "foo.com",
		PingURI: "/racy",
		Up:      true,
	}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			ble.MarkLoadBalancerEndpointUp(false)
			wg.Done()
		}()
	}
	wg.Wait()
	assert.False(t, ble.IsUp())
}
