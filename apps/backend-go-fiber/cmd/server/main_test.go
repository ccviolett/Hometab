package main

import (
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcquireListenerUsesRequestedPortWhenAvailable(t *testing.T) {
	port := freePort(t)

	listener, actualPort, err := acquireListener("127.0.0.1", port)
	require.NoError(t, err)
	defer listener.Close()

	assert.Equal(t, port, actualPort)
}

func TestAcquireListenerSkipsOccupiedPort(t *testing.T) {
	port := freePort(t)
	occupied, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	require.NoError(t, err)
	defer occupied.Close()

	listener, actualPort, err := acquireListener("127.0.0.1", port)
	require.NoError(t, err)
	defer listener.Close()

	assert.Greater(t, actualPort, port)
	assert.LessOrEqual(t, actualPort, port+maxPortSearchAttempts-1)
}

func TestAcquireListenerRejectsInvalidPort(t *testing.T) {
	listener, actualPort, err := acquireListener("127.0.0.1", 0)

	require.Error(t, err)
	assert.Nil(t, listener)
	assert.Equal(t, 0, actualPort)
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}
