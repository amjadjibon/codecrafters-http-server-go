package main

import (
	"log"
	"net"
	"strings"
)

type Request struct {
	Method   string
	URI      string
	Protocol string
}

func parseRequest(req []byte) Request {
	request := Request{}

	// split the request into lines
	lines := strings.Split(string(req), "\r\n")

	// split the first line
	firstLine := strings.Split(lines[0], " ")

	// parse the first line
	request.Method = firstLine[0]
	request.URI = firstLine[1]
	request.Protocol = firstLine[2]

	return request
}

func handleConnection(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	buf := make([]byte, 1024)
	req, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	// parse the request
	request := parseRequest(buf[:req])

	// handle the request based on the method and URI
	if request.Method != "GET" {
		_, _ = conn.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
		return
	}

	if request.URI == "/" || request.URI == "/index.html" {
		_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return
	} else {
		_, _ = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}
}

func main() {
	var addr = "localhost:4221"
	log.Printf("Server started on: %s", addr)

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		handleConnection(conn)
	}
}
