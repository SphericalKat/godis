package main

import (
	"log"
	"net"
	"os"
)

func main() {
	// listen on port 6379
	lis, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatal("something went wrong:", err)
		os.Exit(1)
	}

	defer lis.Close()

	log.Println("listening for connections on port 6379")

	// start accepting incoming connections
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		log.Println("got connection:", conn.RemoteAddr())
		go handleIncomingRequest(conn)
	}
}

func handleIncomingRequest(conn net.Conn) {
	// store incoming data
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	// respond
	conn.Write([]byte("Hi back!\n"))

	// close conn
	conn.Close()
}
