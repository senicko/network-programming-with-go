package ch11

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEchoServerTLS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverAddress := "localhost:34443"
	maxIdle := time.Second
	server := NewTLSServer(ctx, serverAddress, maxIdle, nil)
	done := make(chan struct{})

	go func() {
		err := server.ListenAndServeTLS("cert.pem", "key.pem")
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Error(err)
			return
		}
		done <- struct{}{}
	}()
	server.Ready()

	cert, err := os.ReadFile("cert.pem")
	if err != nil {
		t.Fatal(err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		t.Fatal("failed to append certificate to pool")
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		MinVersion:       tls.VersionTLS12,
		RootCAs:          certPool,
	}

	// Connect to the server and initiate a tls handshake
	conn, err := tls.Dial("tcp", serverAddress, tlsConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Write "hello" message
	hello := []byte("hello")
	_, err = conn.Write(hello)
	if err != nil {
		t.Fatal(err)
	}

	// Read the ponged "hello" message
	b := make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		t.Fatal(err)
	}

	// Check if retrieved bytes are equal to the sent ones
	if actual := b[:n]; !bytes.Equal(hello, actual) {
		t.Fatalf("expected %q; actual %q", hello, actual)
	}

	// Wait for connection to be closed because of too long idle
	time.Sleep(2 * maxIdle)

	// Try to read the message (should result in io.EOF error)
	_, err = conn.Read(b)
	if err != io.EOF {
		t.Fatal(err)
	}

	// Close the connection
	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}

	cancel()
	<-done
}
