package ch03

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextCancelFanOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			conn.Close()
		}
	}()

	dial := func(ctx context.Context, address string, responseCh chan int, id int, wg *sync.WaitGroup) {
		defer wg.Done()

		var d net.Dialer
		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}
		c.Close()

		select {
		case <-ctx.Done():
		case responseCh <- id:
		}
	}

	responseCh := make(chan int)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), responseCh, i+1, &wg)
	}

	response := <-responseCh
	cancel()
	wg.Wait()
	close(responseCh)

	if ctx.Err() != context.Canceled {
		t.Errorf("expected cancelled context, actual: %s", ctx.Err())
	}

	t.Logf("dialer %d retrieved the resource", response)
}
