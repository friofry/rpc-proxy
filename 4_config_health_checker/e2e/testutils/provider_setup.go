package testutils

import (
	"sync"
)

// ProviderSetup manages multiple mock RPC servers
type ProviderSetup struct {
	servers []*MockRPCServer
	wg      sync.WaitGroup
}

// NewProviderSetup creates a new provider setup
func NewProviderSetup() *ProviderSetup {
	return &ProviderSetup{
		servers: make([]*MockRPCServer, 0),
	}
}

// AddProvider adds a new mock provider
func (p *ProviderSetup) AddProvider(port int, responses map[string]map[string]interface{}) *MockRPCServer {
	server := NewMockRPCServer(port)
	for method, response := range responses {
		server.AddResponse(method, response)
	}
	p.servers = append(p.servers, server)
	return server
}

// StartAll starts all mock providers
func (p *ProviderSetup) StartAll() error {
	for _, server := range p.servers {
		p.wg.Add(1)
		go func(s *MockRPCServer) {
			defer p.wg.Done()
			s.Start()
		}(server)
	}
	return nil
}

// StopAll stops all mock providers
func (p *ProviderSetup) StopAll() error {
	for _, server := range p.servers {
		if err := server.Stop(); err != nil {
			return err
		}
	}
	p.wg.Wait()
	return nil
}
