package ch04

import (
	"io"
	"log"
	"net"
	"os"
)

// Monitor embeds a log.Logger meant for logging network traffic.
type Monitor struct {
	*log.Logger
}

// Write implmements the io.Writer interface
func (m *Monitor) Write(p []byte) (int, error) {
	err := m.Output(2, string(p))
	if err != nil {
		log.Println(err)
	}

	// In order to not interrupt network connection we always return nil error from the logger write method.
	return len(p), nil
}

func ExampleMonitor() {
	monitor := &Monitor{
		Logger: log.New(os.Stdout, "monitor: ", 0),
	}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		monitor.Fatal(err)
	}

	done := make(chan struct{})

	go func() {
		defer close(done)

		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		b := make([]byte, 1024)
		r := io.TeeReader(conn, monitor)

		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}

		w := io.MultiWriter(conn, monitor)

		_, err = w.Write(b[:n])
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		monitor.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("Test\n"))
	if err != nil {
		monitor.Fatal(err)
	}

	<-done
}
