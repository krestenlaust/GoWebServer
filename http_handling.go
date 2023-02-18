package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const SUPPORTED_HTTPVERSION_MAJOR = 1
const SUPPORTED_HTTPVERSION_MINOR = 1
const REQUEST_BUFFERSIZE = 384 // 3/4 * 512

type RequestHandler interface {
	MakeResponse(Request) Response
}

type Response struct {
	// General header
	//body             []byte // only 'content' is used for now.
	connectionStatus string
	fields           map[string]string

	// Response header
	date           time.Time
	statusCode     int
	statusResponse string
	content        string
}

func NewResponse() Response {
	res := new(Response)

	res.fields = make(map[string]string)

	return *res
}

func (res *Response) SetContentText(text string) {
	res.fields["content-type"] = "text/html; charset=utf-8"
	res.content = text
}

func (res Response) String() string {
	resStr := fmt.Sprintf("HTTP/1.1 %d %s\r\n", res.statusCode, res.statusResponse) +
		fmt.Sprintf("Date: %s\r\n", res.date.Format("Mon, 01 Jan 2006 15:04:05")) +
		fmt.Sprintf("Connection: %s\r\n", res.connectionStatus) +
		fmt.Sprintf("Content-Length: %d\r\n", len(res.content))

	for key, element := range res.fields {
		resStr += fmt.Sprintf("%s: %s\r\n", PascalifyShishkebabCase(key), element)
	}

	return resStr + "\r\n" + res.content
}

func readRequests(conn net.Conn, ch chan Request) {
	defer close(ch)

	initiatedTime := time.Now()

	// Handle requests from client until closed
	for {
		fmt.Printf("[%s] Request received\n", conn.RemoteAddr().String())

		//dataTransferStartTime := time.Now()

		// TODO: proper timeout, taking both timeouts into account. Choose the one with the shortest timeout
		conn.SetDeadline(initiatedTime.Add(time.Millisecond * time.Duration(REQUEST_TOTAL_TIMEOUT)))
		req, err := NewRequestByConn(conn)

		netErr, ok := err.(net.Error)

		if err != nil {
			if ok {
				if netErr.Timeout() {
					// TODO: Request timed out. Close stream
					return
				}
			} else {
				// TODO: 400 bad request
				fmt.Printf("[%s] Couldn't read or parse request: %s\n", conn.RemoteAddr().String(), err.Error())
				continue
			}
		}

		ch <- req

		if req.connectionStatus == "close" {
			return
		}

		// TODO: wait for more data
	}
}

type Request struct {
	// General header
	body             []byte
	connectionStatus string // 'close', 'keep-alive' etc.
	fields           map[string]string

	// Request header
	method           string
	httpVersionMajor int
	httpVersionMinor int
	requestUri       string
	originator       net.Addr
}

func NewRequest(raw string, origin net.Addr) (Request, error) {
	req := new(Request)
	req.fields = make(map[string]string)
	err := req.parseHeader(raw)

	req.connectionStatus = "close" // closes by default
	req.originator = origin

	return *req, err
}

func NewRequestByConn(conn net.Conn) (Request, error) {
	req := new(Request)
	req.fields = make(map[string]string)

	req.connectionStatus = "close" // closes by default
	req.originator = conn.RemoteAddr()

	reader := bufio.NewReader(conn)
	err := req.parseHeaderByReader(*reader)

	return *req, err
}

func (req *Request) parseHeaderByReader(reader bufio.Reader) error {
	firstLine := true

	// Read header fields
	for {
		lineWithCarriageReturn, err := reader.ReadString('\n')

		if err != nil {
			return err
		}

		line := strings.TrimSuffix(lineWithCarriageReturn, "\r\n")

		if line == "" {
			// End of header
			break
		}

		if firstLine {
			firstLine = false

			// First line is method and version
			words := strings.Split(line, " ")
			req.method = strings.ToLower(words[0])
			req.requestUri = words[1]

			major, minor, err := parseHttpVersion(words[2])

			if err != nil {
				// TODO: 400 bad request
				fmt.Printf("[?] Bad request: %s", err.Error())
				return err
			}

			req.httpVersionMajor = major
			req.httpVersionMinor = minor

			// Not sure in what way they are backwards compatible, so only support HTTP/1.0 and HTTP/1.1
			if SUPPORTED_HTTPVERSION_MAJOR != req.httpVersionMajor && req.httpVersionMinor <= SUPPORTED_HTTPVERSION_MINOR {
				// TODO: 505 Http version not supported
				return errors.New("http version not supported")
			}

			continue
		}

		keyValuePair := strings.SplitN(line, ":", 2)

		// Not a field
		if len(keyValuePair) == 1 {
			// TODO: 400 bad request
			fmt.Printf("[%s] Bad request: %s", req.originator.String(), err.Error())
			return errors.New("bad request")
		}

		req.fields[strings.ToLower(keyValuePair[0])] = strings.TrimLeft(keyValuePair[1], " ")
	}

	contentLengthStr, ok := req.fields["content-length"]

	// Request has body
	if ok {
		contentLength, err := strconv.Atoi(contentLengthStr)

		if err != nil {
			// TODO: 400 bad request, invalid content-length value
			return nil
		}

		req.body = make([]byte, contentLength)
		_, err = reader.Read(req.body)

		netErr, ok := err.(net.Error)

		if ok {
			if netErr.Timeout() {
				// TODO: timeout
				fmt.Print("Timeout")
			}
		} else {
			panic(err)
		}
	}

	return nil
}

func (req *Request) parseHeader(rawHeader string) error {
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
