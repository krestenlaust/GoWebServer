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

	listenHttp("0.0.0.0", 80)
}

func listenHttp(host string, port int) error {
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

		go handleConnection(conn)
	}
}

const REQUEST_TOTAL_TIMEOUT = 10000
const REQUEST_TRANSFER_TIMEOUT = 1000

func handleConnection(conn net.Conn) {
	defer conn.Close()
	defer fmt.Printf("[%s] Client disconnected\n", conn.RemoteAddr().String())

	fmt.Printf("[%s] Request received\n", conn.RemoteAddr().String())

	initiatedTime := time.Now()

	// Handle requests from client until closed
	for {
		dataTransferStartTime := time.Now()
		requestString := ""

		// Receive request
		for {
			if (time.Since(initiatedTime).Milliseconds() > REQUEST_TOTAL_TIMEOUT) || (time.Since(dataTransferStartTime).Milliseconds() > REQUEST_TRANSFER_TIMEOUT) {
				fmt.Printf("[%s] Client timed out\n", conn.RemoteAddr().String())
				return
			}

			dataTransferStartTime = time.Now()

			dataBuffer := make([]byte, 256)

			// set deadline as the timeout duration, to prevent client stalling
			conn.SetDeadline(initiatedTime.Add(time.Millisecond * time.Duration(REQUEST_TOTAL_TIMEOUT)))
			i, err := conn.Read(dataBuffer)

			if err != nil {
				fmt.Printf("[%s] Error occured: %s\n", conn.RemoteAddr().String(), err.Error())
				//return
			}

			requestString += string(dataBuffer[:i])

			fmt.Printf("Read %d bytes \n", i)

			// More data to read?
			if i == len(dataBuffer) {
				continue
			}

			if strings.HasSuffix(requestString, "\r\n\r\n") {
				break
			}
		}

		req, err := NewRequest(requestString, conn.RemoteAddr())

		if err != nil {
			fmt.Printf("[%s] Couldn't parse request: %s\n", conn.RemoteAddr().String(), err.Error())
		}

		if req.contentLength > 0 {
			req.body = make([]byte, req.contentLength)
			conn.Read(req.body)
		}

		res, err := handleRequest(req)

		if err != nil {
			fmt.Printf("[%s] Couldn't handle request: %s\n", conn.RemoteAddr().String(), err.Error())
			return
		}

		conn.Write([]byte(res.String()))

		if req.connectionStatus == "close" {
			break
		}
	}
}

func handleRequest(req Request) (Response, error) {
	res := new(Response)

	switch req.method {
	case "get":
		res.statusCode = 418
		res.statusResponse = "I'm a teapot"
		res.date = time.Now()
		res.connectionStatus = req.connectionStatus
		res.contentType = "text/html; charset=utf-8"
		res.content = "<span>Denne hjemmeside kører på hjemmelavet serversoftware</span>"
		res.contentLength = len(res.content)

		return *res, nil
	}

	return *res, errors.New("server cannot handle the desired request method")
}

type Request struct {
	// General header
	contentLength    int
	body             []byte
	connectionStatus string // 'close', 'keep-alive' etc.

	// Request header
	httpVersionMajor int
	httpVersionMinor int
	requestUri       string
	originator       net.Addr
	method           string
	fields           map[string]string
}

func NewRequest(raw string, origin net.Addr) (Request, error) {
	req := new(Request)
	req.fields = make(map[string]string)
	err := req.ParseHeader(raw)

	req.connectionStatus = "close" // closes by default
	req.originator = origin

	return *req, err
}

// Parses version string, e.g. "http/1.1", "HTTP/2" or similar
func parseHttpVersion(versionString string) (major int, minor int, err error) {
	parts := strings.Split(strings.ToLower(versionString), "/")

	if len(parts) != 2 || parts[0] != "http" {
		return 0, 0, errors.New("invalid version format")
	}

	majorMinor := strings.Split(parts[1], ".")

	majorValue, err := strconv.Atoi(majorMinor[0])

	if err != nil {
		return 0, 0, errors.New("invalid version format")
	}

	// Only major
	if len(majorMinor) == 1 {
		return majorValue, 0, nil
	}

	minorValue, err := strconv.Atoi(majorMinor[1])

	if err != nil {
		return 0, 0, errors.New("invalid version format")
	}

	return majorValue, minorValue, nil
}

const SUPPORTED_HTTPVERSION_MAJOR = 1
const SUPPORTED_HTTPVERSION_MINOR = 1

func (req *Request) ParseHeader(rawHeader string) error {
	lines := strings.Split(rawHeader, "\r\n")

	// First line is method and version
	words := strings.Split(lines[0], " ")
	req.method = strings.ToLower(words[0])
	req.requestUri = words[1]

	major, minor, err := parseHttpVersion(words[2])

	if err != nil {
		// TODO: 400 bad request
		fmt.Printf("[%s] Bad request: %s", req.originator.String(), err.Error())
		return err
	}

	req.httpVersionMajor = major
	req.httpVersionMinor = minor

	// Not sure in what way they are backwards compatible, so only support HTTP/1.0 and HTTP/1.1
	if SUPPORTED_HTTPVERSION_MAJOR != req.httpVersionMajor && req.httpVersionMinor <= SUPPORTED_HTTPVERSION_MINOR {
		// TODO: 505 Http version not supported
		return errors.New("http version not supported")
	}

	for i := 0; i < len(lines); i++ {
		keyValuePair := strings.SplitN(lines[i], ":", 2)

		// Not a field
		if len(keyValuePair) == 1 {
			continue
		}

		req.fields[strings.ToLower(keyValuePair[0])] = strings.TrimLeft(keyValuePair[1], " ")
	}

	return nil
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
