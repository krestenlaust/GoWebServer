package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
    "utf8"

	"github.com/badgerodon/collections"
)

func main() {
	fmt.Println("Hello, world")

	listenHttp("127.0.0.1", 8080)
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

const REQUEST_TOTAL_TIMEOUT = 123
const REQUEST_TRANSFER_TIMEOUT = 123

// TODO: set write deadline to counteract slow-loris attack.
// I think it can be done by simply setting the deadline
// as the remaining timeout time
func handleConnection(conn net.Conn) {
    TERMINATOR_CONSTANT := [4]byte{ 13, 10, 13, 10 };
	fmt.Println("[%s] Request recieved", conn.RemoteAddr().String())

    defer conn.Close()
    defer fmt.Println("[%s] Client disconnected", conn.RemoteAddr().String())

	initiatedTime := time.Now()

	for {
		dataTransferStartTime := time.Now()
		requestData := collections.NewQueue[byte]()

		endCharsIndex := 0

		requestFinished := false
		for !requestFinished {
            if time.Since(initiatedTime).Seconds() > REQUEST_TOTAL_TIMEOUT || time.Since(dataTransferStartTime) > REQUEST_TRANSFER_TIMEOUT
            {
                fmt.Println("[%s] Client timed out", conn.RemoteAddr().String())
                return
            }

            dataTransferStartTime = time.Now()

            newByte := make([]byte, 1)
            _, err := conn.Read(newByte)

            requestData.Push(newByte[0])

            if newByte == TERMINATOR_CONSTANT[endCharsIndex] {
                endCharsIndex++
            }else{
                endCharsIndex = 0
            }

            if endCharsIndex == 4 {
                break
            }
		}

        requestDataBytes := make(byte[requestData.Size()])
        for i := 0; i < requestData.Size(); i++ {
            requestDataBytes[i] = requestData[i]
        }

        requestString := string(requestData[:])

        // Generate request
        req := NewRequest(requestString, conn.RemoteAddr())
        if req.contentLength > 0 {
            req.body = make([]byte, req.contentLength)
            conn.Read(req.body, req.contentLength)
        }

        res, err := handleRequest(req)

        if err != nil {
            fmt.Println("[%s] Couldn't handle request: %s", conn.RemoteAddr().String(), err.Error())
            return
        }

        conn.Write([]byte(res.String()))

        if req.connectionStatus == "close" {
            return
        }
    }
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
		res.content = "<span>Denne hjemmeside kører på hjemmelavet serversoftware</span>"

		return *res, nil
	}

	return *res, errors.New("server cannot handle the desired request method")
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
	originator net.Addr
}

func NewRequest(raw string, origin net.Addr) Request {
    req := new(Request)
    req.connectionStatus = "close" // closes by default
    req.originator = origin
    ParseRequestHeader(raw)
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
				return *res, errors.New("invalid request header, 'Content-Length' value not integer")
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
