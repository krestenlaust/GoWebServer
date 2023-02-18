package main

import (
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

/*
type FileHandler struct{}

func (FileHandler) MakeResponse(req Request) Response {
	res := new(Response)

}*/
