package ch09

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/senicko/network-programming-with-go/ch09/handlers"
)

func getHttpsClientForCerts(certPath string) (*http.Client, error) {
	cert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{Transport: httpTransport}, nil
}

func TestSimpleHTTPServer(t *testing.T) {
	certPath := "./certs/localhost+1.pem"
	keyPath := "./certs/localhost+1-key.pem"

	srv := &http.Server{
		Addr: "127.0.0.1:8443",
		Handler: http.TimeoutHandler(
			handlers.DefaultMethodsHandler(),
			2*time.Minute,
			"",
		),
		IdleTimeout:       5 * time.Minute,
		ReadHeaderTimeout: time.Minute,
		TLSConfig: &tls.Config{

			InsecureSkipVerify: true,
		},
	}

	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := srv.ServeTLS(l, certPath, keyPath)
		if err != http.ErrServerClosed {
			t.Error(err)
		}
	}()

	testCases := []struct {
		method   string
		body     io.Reader
		code     int
		response string
	}{
		{http.MethodGet, nil, http.StatusOK, "Hello, friend!"},
		{http.MethodPost, bytes.NewBufferString("<world>"), http.StatusOK, "Hello, &lt;world&gt;!"},
		{http.MethodHead, nil, http.StatusMethodNotAllowed, ""},
	}

	client := new(http.Client)
	path := fmt.Sprintf("https://%s/", srv.Addr)

	for i, c := range testCases {
		r, err := http.NewRequest(c.method, path, c.body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		res, err := client.Do(r)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		if res.StatusCode != c.code {
			t.Errorf("%d: unexpected status code: %q", i, res.Status)
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		_ = res.Body.Close()

		if c.response != string(b) {
			t.Errorf("%d: expected %q; actual %q", i, c.response, b)
		}
	}

	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}
}
