package test

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

func TCPConn() {
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listen.Close()
	for {
		Conn, err := listen.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go handFunc(&Conn)
	}
}

func handFunc(conn *net.Conn) {
	reader := bufio.NewReader(*conn)
	for {
		readString, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connect close")
			} else {
				fmt.Println(err)
			}
			return
		}
		fmt.Println(readString)
		(*conn).Write([]byte(readString))
	}
}
