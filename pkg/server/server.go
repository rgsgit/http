package server

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const lineCRLF string = "\r\n"

type HandlerFunc func(req *Request)

type Request struct{
	Conn		net.Conn
	//PathParams	map[string]string
	QueryParams	url.Values
}

type Server struct{
	addr 		string
	mu			sync.RWMutex
	handlers	map[string]HandlerFunc
}

func NewServer(adr string) *Server{
	return &Server{addr: adr,handlers: make(map[string]HandlerFunc)}
}

func (s *Server) Register(path string, handler HandlerFunc)  {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

func (s *Server) Start() error{

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := listener.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go s.handle(conn)
	}
}

//handle - handles the connection
func (s *Server) handle(conn net.Conn) {

	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Println(cerr)
		}
	}()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err == io.EOF {
		log.Printf("Error EOF: %s", buf[:n])
		return
	}
	if err != nil {
		log.Println("Error not nil after reading", err)
		return
	}

	//Parsing...
	data := buf[:n]

	requestLineDelim := []byte{'\r', '\n'}
	requestLineEnd := bytes.Index(data, requestLineDelim)

	if requestLineEnd == -1 {
		log.Print("requestLineEndErr: ", requestLineEnd)
		return
	}

	requestLine := string(data[:requestLineEnd])
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		log.Print("Parts: ", parts)
		return
	}

	path, version := parts[1], parts[2]
	if version != "HTTP/1.1" {
		log.Print("Version Not HTTP/1.1: ", version)
		return
	}

	uri, err := url.ParseRequestURI(path)
	if err != nil {
		log.Print(err)
		return
	}
	log.Print(uri.Path)
	log.Print(uri.Query())

	s.mu.RLock()
	if handler, ok := s.handlers[uri.Path]; ok {
		s.mu.RUnlock()
		handler(&Request{
			Conn:        conn,
			QueryParams: uri.Query(),
		})
	}
}


func (s *Server) Response(body string) string {
	return "HTTP/1.1 200 OK\r\n" +
		"Content-Length: " + strconv.Itoa(len(body)) + lineCRLF +
		"Content-Type: text/html\r\n" +
		"Connection: close\r\n" +
		lineCRLF + body
}