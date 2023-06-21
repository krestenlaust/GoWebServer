package main

import (
	"fmt"
	"net"
	"strconv"
)

const REQUEST_TOTAL_TIMEOUT = 10000
const REQUEST_TRANSFER_TIMEOUT = 1000

func ListenHttp(host string, port int, handler RequestHandler) error {
	hostport := net.JoinHostPort(host, strconv.Itoa(port))
	l, err := net.Listen("tcp", hostport)

	if err != nil {
		return err
	}

	defer l.Close()
	fmt.Printf("Listening on %s\n", hostport)

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			continue
		}

		go handleConnection(conn, handler)
	}
}

func handleConnection(conn net.Conn, handler RequestHandler) {
	defer conn.Close()
	defer fmt.Printf("[%s] Client disconnected\n", conn.RemoteAddr().String())

	ch := make(chan Request)
	go readRequests(conn, ch)

	for {
		// TODO: can't remember what to do with channels.

		req := <-ch // breakpoint
		res := handler.MakeResponse(req)

		conn.Write([]byte(res.String()))

		if res.connectionStatus == "close" {
			return
		}
	}
}
