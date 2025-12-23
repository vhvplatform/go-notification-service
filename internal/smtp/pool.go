package smtp

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"sync"
)

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
}

// SMTPPool manages a pool of SMTP connections
type SMTPPool struct {
	connections chan *smtp.Client
	config      SMTPConfig
	size        int
	mu          sync.Mutex
	closed      bool
}

// NewSMTPPool creates a new SMTP connection pool
func NewSMTPPool(config SMTPConfig, size int) (*SMTPPool, error) {
	pool := &SMTPPool{
		connections: make(chan *smtp.Client, size),
		config:      config,
		size:        size,
		closed:      false,
	}

	// Initialize pool with connections
	for i := 0; i < size; i++ {
		client, err := pool.createConnection()
		if err != nil {
			// Close any already created connections
			pool.Close()
			return nil, fmt.Errorf("failed to initialize connection pool: %w", err)
		}
		pool.connections <- client
	}

	return pool, nil
}

// createConnection creates a new SMTP connection
func (p *SMTPPool) createConnection() (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	var client *smtp.Client
	var err error

	if p.config.UseTLS {
		tlsConfig := &tls.Config{
			ServerName:         p.config.Host,
			InsecureSkipVerify: false, // Always verify certificates in production
			MinVersion:         tls.VersionTLS12,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to dial TLS: %w", err)
		}
		client, err = smtp.NewClient(conn, p.config.Host)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to dial SMTP: %w", err)
		}
	}

	// Authenticate if credentials are provided
	if p.config.Username != "" && p.config.Password != "" {
		auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)
		if err := client.Auth(auth); err != nil {
			client.Quit()
			return nil, fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	return client, nil
}

// Get retrieves a connection from the pool
func (p *SMTPPool) Get() (*smtp.Client, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, fmt.Errorf("connection pool is closed")
	}
	p.mu.Unlock()

	select {
	case client := <-p.connections:
		// Test connection with NOOP
		if err := client.Noop(); err != nil {
			// Connection dead, close it and create new one
			client.Quit()
			newClient, err := p.createConnection()
			if err != nil {
				return nil, fmt.Errorf("failed to create new connection: %w", err)
			}
			return newClient, nil
		}
		return client, nil
	default:
		// Pool empty, create new connection temporarily
		return p.createConnection()
	}
}

// Put returns a connection to the pool
func (p *SMTPPool) Put(client *smtp.Client) {
	if client == nil {
		return
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		client.Quit()
		return
	}
	p.mu.Unlock()

	select {
	case p.connections <- client:
		// Successfully returned to pool
	default:
		// Pool full, close connection
		client.Quit()
	}
}

// Close closes all connections in the pool
func (p *SMTPPool) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	close(p.connections)
	for client := range p.connections {
		if client != nil {
			client.Quit()
		}
	}
}

// Size returns the pool size
func (p *SMTPPool) Size() int {
	return p.size
}
