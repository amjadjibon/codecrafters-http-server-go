package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Request struct {
	Method   string
	Headers  map[string]string
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

	// parse the headers
	request.Headers = make(map[string]string)
	for _, line := range lines[1:] {
		if line == "" {
			break // end of headers section
		}
		header := strings.Split(line, ": ")
		request.Headers[header[0]] = header[1]
	}

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
	} else if strings.HasPrefix(request.URI, "/echo") {
		// get the message from the URI after the /echo
		message := strings.TrimPrefix(request.URI, "/echo/")
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)
		_, _ = conn.Write([]byte(response))
	} else if request.URI == "/user-agent" {
		userAgentHeader := request.Headers["User-Agent"]
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
			len(userAgentHeader), userAgentHeader)
		_, _ = conn.Write([]byte(response))
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
