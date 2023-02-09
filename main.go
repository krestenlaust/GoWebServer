package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("Hello, world")
}

func listenHttp(host string, port int) error {
	hostport := net.JoinHostPort(host, strconv.Itoa(port))
	l, err := net.Listen("tcp", hostport)

	if err != nil {
		return err
	}

	defer l.Close()
	fmt.Println("Listening on " + hostport)

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error: " + err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

}

// TODO: should receive the network connection instead of a parsed request,
// this is because the main blocking action would be the parsing and retrieving.
func handleRequest(req Request) (Response, error) {
	res := new(Response)

	switch req.method {
	case "get":
		res.statusCode = 420
		res.statusResponse = "Enhance your calm"
		res.date = time.Now()
		res.connectionStatus = req.connectionStatus
		res.contentType = "text/html"
		res.content = "<span>Hello, world</span>"

		return *res, nil
	}

	return *res, errors.New("Server cannot handle the desired request method")
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

func ParseRequestHeader(rawHeader string) (Request, error) {
	res := new(Request)
	lines := strings.Split(rawHeader, "\r\n")

	for i := 0; i < len(lines); i++ {
		if strings.HasSuffix(strings.ToLower(lines[i]), "http/1.1") {
			words := strings.Split(lines[i], " ")

			res.method = strings.ToLower(words[0])
			res.requestUri = words[1]
			continue
		}

		keyValuePair := strings.SplitN(lines[i], ":", 2)

		switch strings.ToLower(keyValuePair[0]) {
		case "user-agent":
			res.userAgent = keyValuePair[1]
		case "host":
			res.host = keyValuePair[1]
		case "content-length":
			i, err := strconv.Atoi(strings.TrimSpace(keyValuePair[i]))

			if err != nil {
				return *res, errors.New("Invalid request header, 'Content-Length' value not integral")
			}

			res.contentLength = i
		}
	}

	return *res, nil
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
