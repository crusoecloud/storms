package loadbalancer

import (
	"errors"
	"net"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

var errDialError = errors.New("failed to dial all downstream services")

type Algorithm string

const AlgorithmRoundRobin Algorithm = "round-robin"

// The LoadBalancer type can be used to transparently add fault tolerance to any client connection.
// The LoadBalancer interface supports multiple load-balancing algorithms, but currently only the
// "round-robin" strategy is supported.
type LoadBalancer struct {
	algorithm Algorithm
	services  []*DownstreamService
	dialCount uint64 // The number of dial requests. This counter must be accessed atomically.
}

func NewLoadBalancer(algorithm Algorithm, addrs ...net.Addr) *LoadBalancer {
	defaultDialFunc := net.Dial
	services := make([]*DownstreamService, len(addrs))
	for i, addr := range addrs {
		services[i] = newDownstreamService(addr, defaultDialFunc)
	}

	return &LoadBalancer{
		algorithm: algorithm,
		services:  services,
		dialCount: 0,
	}
}

func (lb *LoadBalancer) DialCount() uint64 {
	return atomic.LoadUint64(&lb.dialCount)
}

func (lb *LoadBalancer) Dial() (*Conn, error) {
	serviceCount := uint64(len(lb.services))
	dialCount := atomic.AddUint64(&lb.dialCount, 1)
	for i := uint64(0); i < serviceCount; i++ {
		service := lb.services[(i+dialCount-1)%serviceCount]
		conn, err := service.Dial()
		activeCount := service.ActiveCount()
		errorCount := service.ErrorCount()
		if err == nil {
			log.Info().
				Uint64("active_count", activeCount).
				Uint64("error_count", errorCount).
				Msg("Connected to downstream service")

			return conn, nil
		}

		log.Error().
			Uint64("active_count", activeCount).
			Uint64("error_count", errorCount).
			Err(err).Msg("Failed to connect to downstream service")
	}

	return nil, errDialError
}
