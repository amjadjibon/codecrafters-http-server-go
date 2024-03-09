package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
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

func handleConnection(conn net.Conn, directory string) {
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
		message := strings.TrimPrefix(request.URI, "/echo/")
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)
		_, _ = conn.Write([]byte(response))
	} else if request.URI == "/user-agent" {
		userAgentHeader := request.Headers["User-Agent"]
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
			len(userAgentHeader), userAgentHeader)
		_, _ = conn.Write([]byte(response))
	} else if strings.HasPrefix(request.URI, "/files/") {
		filePath := strings.TrimPrefix(request.URI, "/files/")
		file, err := os.Open(directory + "/" + filePath)
		if err != nil {
			_, _ = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}

		// read the file

		fileBody, err := io.ReadAll(file)
		if err != nil {
			_, _ = conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}

		response := fmt.Sprintf(
			"HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
			len(fileBody), fileBody,
		)
		_, _ = conn.Write([]byte(response))
	} else {
		_, _ = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}
}

func parseFlags(args []string) map[string]string {
	flags := make(map[string]string)
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") {
			if strings.Contains(arg, "=") {
				flag := strings.Split(arg, "=")
				flags[flag[0][2:]] = flag[1]
			} else {
				flags[args[i][2:]] = args[i+1]
			}
		}
	}
	return flags
}

func main() {
	var addr = "localhost:4221"
	log.Printf("Server started on: %s", addr)

	flags := parseFlags(os.Args)
	directory, ok := flags["directory"]
	if ok {
		log.Printf("Serving files from: %s", directory)
	} else {
		log.Printf("Serving files from: %s", ".")
	}

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn, directory)
	}
}
