package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/senicko/network-programming-with-go/ch12/housework/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var addr, certFn, keyFn string

func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "listen address")
	flag.StringVar(&certFn, "cert", "cert.pem", "certificate filename")
	flag.StringVar(&keyFn, "key", "key.pem", "private key filename")
}

func main() {
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(certFn, keyFn)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(
		&tls.Config{
			Certificates:     []tls.Certificate{cert},
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			MinVersion:       tls.VersionTLS12,
		},
	)))

	rosie := new(Rosie)
	housework.RegisterRobotMaidServer(server, rosie.Service())

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Listening for TLS connection on %s ...\n", addr)

	log.Fatal(server.Serve(listener))
}
