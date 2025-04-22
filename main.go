package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		panic(err)
	}

	bytes := make([]byte, 8)
	for {
		n, err := file.Read(bytes)
		if err != nil {
			panic(err)
		}
		fmt.Printf("read: %s\n", string(bytes[:n]))
		if n < 8 {
			break
		}
	}

	defer closer(file)
}

func closer(open *os.File) {
	err := open.Close()
	if err != nil {
		panic(err)
	}
}
