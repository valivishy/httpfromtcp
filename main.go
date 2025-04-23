package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		panic(err)
	}

	channel := getLinesChannel(file)
	for value := range channel {
		fmt.Printf("read: %s\n", value)
	}

	defer closer(file)
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

func closer(open *os.File) {
	err := open.Close()
	if err != nil {
		panic(err)
	}
}
