package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		panic(err)
	}

	udp, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer safeClose(udp)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Println(err.Error())
		}
		_, err = udp.Write(line)
		if err != nil {
			log.Println(err.Error())
		}

	}
}

func safeClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		panic(err)
	}
}
