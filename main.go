package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("Hello, world")
}

type Request struct {
	// General header
	contentLength    int
	body             []byte
	connectionStatus string // 'close', 'keep-alive' etc.
	date             time.Time

	// Request header
	method     string
	requestUri string
	userAgent  string
	host       string
	originator net.IP
}

type Response struct {
	// General header
	contentLength    int
	body             []byte
	connectionStatus string // 'close', 'keep-alive' etc.
	date             time.Time

	// Response header
	statusCode     int
	statusResponse string
	contentType    string
	content        string
}

func (res Response) String() string {
	return fmt.Sprintf("HTTP/1.1 %d %s\r\n", res.statusCode, res.statusResponse) +
		fmt.Sprintf("Date: %s\r\n", res.date.Format("Mon, 01 Jan 2006 15:04:05")) +
		fmt.Sprintf("Connection: %s\r\n", res.connectionStatus) +
		fmt.Sprintf("Content-Length: %d\r\n", res.contentLength) +
		fmt.Sprintf("Content-Type: %s\r\n\r\n", res.contentType) +
		res.content
}
