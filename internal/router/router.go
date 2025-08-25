package router

import (
	"fmt"
	"net"
	"sync"

	"neo4j-proxy/pkg/config"
)

// Router handles routing connections to appropriate Neo4j backends
type Router struct {
	config *config.Config
	mu     sync.RWMutex
}

// New creates a new router instance
func New(cfg *config.Config) *Router {
	return &Router{
		config: cfg,
	}
}

// RouteConnection establishes connection to the appropriate backend for the tenant
func (r *Router) RouteConnection(tenantID string) (net.Conn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantConfig, exists := r.config.Tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant %s not found", tenantID)
	}

	// Connect to the backend Neo4j instance
	address := fmt.Sprintf("%s:%d", tenantConfig.Host, tenantConfig.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to backend %s: %w", address, err)
	}

	return conn, nil
}

// GetTenantConfig returns configuration for a specific tenant
func (r *Router) GetTenantConfig(tenantID string) (*config.TenantConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, exists := r.config.Tenants[tenantID]
	return &cfg, exists
}

// ListTenants returns all configured tenant IDs
func (r *Router) ListTenants() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenants := make([]string, 0, len(r.config.Tenants))
	for tenantID := range r.config.Tenants {
		tenants = append(tenants, tenantID)
	}
	return tenants
}

// UpdateTenantConfig updates configuration for a tenant
func (r *Router) UpdateTenantConfig(tenantID string, cfg config.TenantConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.config.Tenants[tenantID] = cfg
}

// RemoveTenant removes a tenant configuration
func (r *Router) RemoveTenant(tenantID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.config.Tenants, tenantID)
}