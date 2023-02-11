package main

import (
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

			dataBuffer := make([]byte, REQUEST_BUFFERSIZE)

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
			// 400 bad request
			fmt.Printf("[%s] Couldn't parse request: %s\n", conn.RemoteAddr().String(), err.Error())
			continue
		}

		contentLength, err := strconv.Atoi(req.fields["content-length"])

		if err != nil {
			// 400 bad request, invalid content-length value
			continue
		}

		if contentLength > 0 {
			req.body = make([]byte, contentLength)
			conn.Read(req.body)
		}

		ch <- req

		if req.connectionStatus == "close" {
			return
		}
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
	err := req.ParseHeader(raw)

	req.connectionStatus = "close" // closes by default
	req.originator = origin

	return *req, err
}

func (req *Request) ParseHeader(rawHeader string) error {
	lines := strings.Split(rawHeader, "\r\n")

	// First line is method and version
	words := strings.Split(lines[0], " ")
	req.method = strings.ToLower(words[0])
	req.requestUri = words[1]

	major, minor, err := ParseHttpVersion(words[2])

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
