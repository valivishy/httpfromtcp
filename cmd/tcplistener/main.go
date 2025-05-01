package main

import (
	"fmt"
	"github.com/valivishy/httpfromtcp/internal/request"
	"io"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}
	defer safeClose(listener)

	for {
		accept, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Println("Connection accepted")

		fromReader, _ := request.FromReader(accept)
		fmt.Println("Request line:")
		line := fromReader.RequestLine
		fmt.Printf("- Method: %s\n", line.Method)
		fmt.Printf("- Target: %s\n", line.RequestTarget)
		fmt.Printf("- Version: %s\n", line.HttpVersion)

		fmt.Println("Headers:")
		for name, value := range fromReader.Headers {
			fmt.Printf("- %s: %s\n", name, value)
		}

		fmt.Println("Body:")
		fmt.Println(string(fromReader.Body))

		fmt.Println("Connection closed")
		safeClose(accept)
	}
}

func safeClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		fmt.Printf("Error closing: %v\n", err)
	}
}
