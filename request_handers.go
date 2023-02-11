package main

import (
	"time"
)

type BasicHandler struct{}

func (BasicHandler) MakeResponse(req Request) Response {
	res := new(Response)
	res.date = time.Now()

	switch req.method {
	case "get":
		res.statusCode = 418
		res.statusResponse = "I'm a teapot"
		res.connectionStatus = req.connectionStatus
		res.SetContentText("<span>Denne hjemmeside kører på hjemmelavet serversoftware</span>")
	default:
		res.statusCode = 501
		res.statusResponse = "Not implemented"
		res.connectionStatus = req.connectionStatus
	}

	// 501 method not supported
	return *res //, errors.New("server cannot handle the desired request method")
}

/*
type FileHandler struct{}

func (FileHandler) MakeResponse(req Request) Response {
	res := new(Response)

}*/
