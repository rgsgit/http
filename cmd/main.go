package main

import (
	"log"
	"net"
	"os"

	"github.com/rgsgit/http/pkg/server"
)

func main() {
	host := "0.0.0.0"
	port := "9999"

	if err := execute(host, port); err != nil {
		os.Exit(1)
	}
}

func execute(host string, port string) (err error) {
	srv := server.NewServer(net.JoinHostPort(host, port))

	/*srv.Register("/", func(conn net.Conn) {
		body := "Welcome to our web-site"

		_, err = conn.Write([]byte(srv.Response(body)))
		if err != nil {
			log.Print(err)
		}
	})

	srv.Register("/about", func(conn net.Conn) {
		body := "About Golang Academy"

		_, err = conn.Write([]byte(srv.Response(body)))
		if err != nil {
			log.Print(err)
		}
	})*/

	log.Print(0)
	return srv.Start()
}
