package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, world")

	ListenHttp("0.0.0.0", 80, new(FileHandler))
}
