package main

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"time"
)

type BasicHandler struct{}

func (BasicHandler) MakeResponse(req Request) Response {
	res := NewResponse()
	res.date = time.Now()

	switch req.method {
	case "get":
		res.statusCode = 200
		res.statusResponse = "OK"
		res.connectionStatus = req.connectionStatus
		res.SetContentText("<span>Denne hjemmeside kører på hjemmelavet serversoftware</span>")
	default:
		// 501 method not supported
		res.statusCode = 501
		res.statusResponse = "Not implemented"
		res.connectionStatus = req.connectionStatus
	}

	return res
}

type FileHandler struct{}

func (FileHandler) MakeResponse(req Request) Response {
	res := NewResponse()
	res.date = time.Now()
	res.connectionStatus = req.connectionStatus

	switch req.method {
	case "get":
		targetFile := strings.TrimLeft(req.requestUri, "/")

		if targetFile == "" {
			targetFile = "index.html"
		}

		if !fs.ValidPath(targetFile) {
			res.statusCode = 400
			res.statusResponse = "Bad Request"
			return res
		}

		file, err := os.Open(targetFile)
		defer file.Close()

		if errors.Is(err, os.ErrNotExist) {
			res.statusCode = 404
			res.statusResponse = "Not Found"
			res.SetContentText("<span>404 - Jeg kan ikke finde den fil du leder efter. - Hr. Server</span>")
			return res
		} else if err != nil {

			//res.statusCode = 500
			//res.statusResponse = "Internal Server Error"
			//res.SetContentText("<span>500 - Det var ikke så godt. - Hr. Server</span>")
			//fmt.Println(err)
			//return res

			// All errors should be handled separately.
			panic(err)
		}

		fileStat, err := file.Stat()

		if err != nil {
			panic(err)
		}

		if fileStat.IsDir() {
			res.statusCode = 200
			res.statusResponse = "OK"
			res.SetContentText("<span>Dette er jo en mappe😶. - Hr. Server</span>")
			return res
		}

		var fileData []byte

		for {
			buffer := make([]byte, 512)
			i, err := file.Read(buffer)

			if i == 0 {
				break
			}

			if err != nil {
				panic(err)
			}

			newSlice := buffer[:i]

			fileData = append(fileData, newSlice...)
		}

		if strings.HasSuffix(targetFile, ".html") {
			res.fields["content-type"] = "text/html; charset=utf-8"
		}

		res.statusCode = 200
		res.statusResponse = "OK"
		res.content = fileData
	default:
		res.statusCode = 501
		res.statusResponse = "Not implemented"
	}

	return res
}
