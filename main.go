package main

import (
	"fmt"
)

func main() {
	fmt.Println("Server starting...")
	
	ListenHttp("0.0.0.0", 80, new(FileHandler))
}
