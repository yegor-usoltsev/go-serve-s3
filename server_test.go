package main

import (
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	t.Parallel()
	cfg := Config{
		ServerHost: "2001:db8::1234:5678",
		ServerPort: 8080,
	}
	handler := http.NewServeMux()

	server := NewServer(cfg, handler)

	assert.Equal(t, "[2001:db8::1234:5678]:8080", server.Addr)
	assert.Equal(t, handler, server.Handler)
	assert.Equal(t, readTimeout, server.ReadTimeout)
	assert.Equal(t, writeTimeout, server.WriteTimeout)
	assert.Equal(t, idleTimeout, server.IdleTimeout)
}

func TestServer_StartAsync_Stop(t *testing.T) {
	t.Parallel()
	cfg := Config{
		ServerHost: "localhost",
		ServerPort: 0,
	}
	handler := http.NewServeMux()
	server := NewServer(cfg, handler)

	assert.NotPanics(t, func() { server.StartAsync() })
	assert.NotPanics(t, func() { server.Stop() })
}

func TestServer_StartAsync_AddressAlreadyInUse(t *testing.T) {
	t.Parallel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	defer func() { _ = listener.Close() }()
	require.NoError(t, err)
	addr := listener.Addr().(*net.TCPAddr)

	cfg := Config{
		ServerHost: addr.IP.String(),
		ServerPort: uint16(addr.Port),
	}
	handler := http.NewServeMux()
	server := NewServer(cfg, handler)

	assert.Panics(t, func() { server.StartAsync() })
}
