package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"neo4j-proxy/internal/auth"
	"neo4j-proxy/internal/router"
	"neo4j-proxy/pkg/bolt"
	"neo4j-proxy/pkg/config"
)

// Proxy represents the Neo4j multi-tenant proxy server
type Proxy struct {
	config        *config.Config
	router        *router.Router
	authenticator *auth.Authenticator
	listener      net.Listener
	wg            sync.WaitGroup
}

// New creates a new proxy instance
func New(cfg *config.Config) *Proxy {
	return &Proxy{
		config:        cfg,
		router:        router.New(cfg),
		authenticator: auth.New(auth.NewUsernameBasedExtractor()),
	}
}

// Start starts the proxy server
func (p *Proxy) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", p.config.ProxyPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	p.listener = listener

	log.Printf("Proxy listening on %s", addr)

	go func() {
		<-ctx.Done()
		p.listener.Close()
	}()

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
		}

		p.wg.Add(1)
		go p.handleConnection(ctx, conn)
	}
}

// Stop stops the proxy server gracefully
func (p *Proxy) Stop() error {
	if p.listener != nil {
		p.listener.Close()
	}
	p.wg.Wait()
	return nil
}

// handleConnection handles a single client connection
func (p *Proxy) handleConnection(ctx context.Context, clientConn net.Conn) {
	defer p.wg.Done()
	defer clientConn.Close()

	log.Printf("New connection from %s", clientConn.RemoteAddr())

	// Wrap the connection with Bolt protocol handler
	boltConn := bolt.NewConnection(clientConn)

	// Perform Bolt handshake with client
	if err := boltConn.Handshake(); err != nil {
		log.Printf("Handshake failed with client %s: %v", clientConn.RemoteAddr(), err)
		return
	}

	log.Printf("Bolt handshake successful with client %s, version: %d", 
		clientConn.RemoteAddr(), boltConn.GetVersion())

	// Read the first message to determine tenant
	firstMsg, err := boltConn.ReadMessage()
	if err != nil {
		log.Printf("Failed to read first message from client %s: %v", clientConn.RemoteAddr(), err)
		return
	}

	// Extract tenant ID from the connection/message
	tenantID, err := p.determineTenant(boltConn, firstMsg)
	if err != nil {
		log.Printf("Failed to determine tenant for client %s: %v", clientConn.RemoteAddr(), err)
		return
	}

	log.Printf("Client %s routed to tenant: %s", clientConn.RemoteAddr(), tenantID)

	// Establish connection to backend
	backendConn, err := p.router.RouteConnection(tenantID)
	if err != nil {
		log.Printf("Failed to route to tenant %s: %v", tenantID, err)
		return
	}
	defer backendConn.Close()

	log.Printf("Connected to backend for tenant %s", tenantID)

	// Create backend Bolt connection
	backendBolt := bolt.NewConnection(backendConn)

	// Perform handshake with backend
	if err := backendBolt.Handshake(); err != nil {
		log.Printf("Backend handshake failed for tenant %s: %v", tenantID, err)
		return
	}

	// Forward the first message to backend
	if err := backendBolt.WriteMessage(firstMsg); err != nil {
		log.Printf("Failed to forward first message to backend for tenant %s: %v", tenantID, err)
		return
	}

	// Start bidirectional proxy
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	// Forward from client to backend
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := p.forwardData(clientConn, backendConn, "client->backend"); err != nil {
			log.Printf("Client->Backend forwarding error for tenant %s: %v", tenantID, err)
		}
	}()

	// Forward from backend to client  
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := p.forwardData(backendConn, clientConn, "backend->client"); err != nil {
			log.Printf("Backend->Client forwarding error for tenant %s: %v", tenantID, err)
		}
	}()

	// Wait for context cancellation or forwarding to complete
	select {
	case <-ctx.Done():
		log.Printf("Connection context cancelled for tenant %s", tenantID)
	}

	wg.Wait()
	log.Printf("Connection closed for client %s, tenant %s", clientConn.RemoteAddr(), tenantID)
}

// determineTenant determines which tenant this connection should be routed to
func (p *Proxy) determineTenant(conn *bolt.Connection, firstMsg *bolt.Message) (string, error) {
	// For now, use simple extraction from the Bolt connection
	// In production, this would parse INIT message authentication details
	tenantID, err := conn.ExtractTenantID(firstMsg)
	if err != nil {
		// Fallback to simple round-robin or default tenant
		tenants := p.router.ListTenants()
		if len(tenants) > 0 {
			return tenants[0], nil
		}
		return "", fmt.Errorf("no tenants configured")
	}

	return tenantID, nil
}

// forwardData forwards data from source to destination
func (p *Proxy) forwardData(src, dst net.Conn, direction string) error {
	_, err := io.Copy(dst, src)
	return err
}