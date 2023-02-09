package main

import (
	"fmt"
    "time"
    "net"
)

func main() {
    fmt.Println("Hello, world")
}

type Request struct {
    // General header
    int contentLength
    []byte body
    string connectionStatus // 'close', 'keep-alive' etc.
    Time date

    // Request header
    string method
    string requestUri
    string userAgent
    string host
    IP originator
}

type Response struct {
    // General header
    int contentLength
    []byte body
    string connectionStatus
    Time date

    // Response header
    int statusCode
    string statusResponse
    string contentType
    string content
}

func (res Response) String() string {
    return fmt.Sprintf("HTTP/1.1 %d %s\r\n", res.statusCode, res.statusResponse) +
    	fmt.Sprintf("Date: \r\n", res.date.Format("ddd, dd MMM yyyy T")) +
	fmt.Sprintf("Connection: %s\r\n", res.connectionStatus) +
	fmt.Sprintf("Content-Length: %s\r\n", res.contentLength) +
	fmt.Sprintf("Content-Type: %s\r\n\r\n", res.contentType) +
	res.content
}
