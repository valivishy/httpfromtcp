package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}
	defer closer(listener)

	for {
		accept, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Println("Connection accepted")
		channel := getLinesChannel(accept)

		for value := range channel {
			fmt.Println(value)
		}
		fmt.Println("Connection closed")
		conCloser(accept)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	channel := make(chan string)

	go func() {
		bytes := make([]byte, 8)
		storage := ""
		for {
			n, err := f.Read(bytes)
			if n > 0 {
				storage += string(bytes[:n])
				if strings.Contains(storage, "\n") {
					split := strings.Split(storage, "\n")
					channel <- split[0]
					storage = split[1]
				}
				if n < 8 {
					if len(storage) > 0 {
						channel <- storage
					}

					close(channel)
					return
				}
			}
			if err == io.EOF {
				if len(storage) > 0 {
					channel <- storage
				}

				close(channel)
				return
			}
			if err != nil {
				panic(err)
			}
		}
	}()
	return channel
}

func closer(target net.Listener) {
	err := target.Close()
	if err != nil {
		panic(err)
	}
}

func conCloser(target net.Conn) {
	err := target.Close()
	if err != nil {
		panic(err)
	}
}
