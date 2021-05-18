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
	PathParams	map[string]string
	QueryParams	url.Values
	Headers     map[string]string
	Body        []byte
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

func (s *Server) handle(conn net.Conn) {

	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Println(cerr)
		}
	}()

	var req Request
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
	
	data := buf[:n]
	requestDelimetr := []byte{'\r', '\n'}
	requestLine := bytes.Index(data, requestDelimetr)
	if requestLine == -1 {
		log.Print("requestLineErr: ", requestLine)
		return
	}
	
	headerDelimetr := []byte{'\r', '\n', '\r', '\n'}
	headerLine := bytes.Index(data, headerDelimetr)
	if headerLine == -1 {
		log.Print("headerLineEnd: ", headerLine)
		return
	}

	headersLine := string(data[requestLine:headerLine])
	header := strings.Split(headersLine, "\r\n")

	paramMap := make(map[string]string)
	for _, v := range header {
		if v != "" {
			eachHeaderLine := strings.Split(v, ": ")
			paramMap[eachHeaderLine[0]] = eachHeaderLine[1]
		}
	}
	req.Headers = paramMap
	
	body := string(data[headerLine:])
	bodyLine := strings.Trim(body, "\r\n")
	req.Body = []byte(bodyLine)
	
	request := string(data[:requestLine])
	parts := strings.Split(request, " ")
	if len(parts) != 3 {
		log.Print("Parts: ", parts)
		return
	}

	path, version := parts[1], parts[2]
	if version != "HTTP/1.1" {
		log.Print("Version Not HTTP/1.1: ", version)
		return
	}

	decoded, err := url.PathUnescape(path)
	if err != nil {
		log.Print(err)
		return
	}
	log.Println(decoded)

	uri, err := url.ParseRequestURI(decoded)
	if err != nil {
		log.Print(err)
		return
	}
	log.Print(uri.Path)
	log.Print(uri.Query())

	req.Conn = conn
	req.QueryParams = uri.Query()

	handler := func(req *Request) {
		req.Conn.Close()
	}
	s.mu.RLock()
	pathPar, ok := s.findPath(uri.Path)
	if ok != nil {
		req.PathParams = pathPar
		handler = ok
	}
	s.mu.RUnlock()
	handler(&req)
}


func (s *Server) findPath(path string) (map[string]string, HandlerFunc) {

	registRoutes := make([]string, len(s.handlers))
	i := 0
	for k := range s.handlers {
		registRoutes[i] = k
		i++
	}

	paramMap := make(map[string]string)

	for i := 0; i < len(registRoutes); i++ {
		flag := false
		eachRegistRoutes := registRoutes[i]
		partsOfRegistRoutes := strings.Split(eachRegistRoutes, "/")
		partsOfClientRoutes := strings.Split(path, "/")

		for j, v := range partsOfRegistRoutes {
			if v != "" {
				f := v[0:1]
				l := v[len(v)-1:]
				if f == "{" && l == "}" {
					paramMap[v[1:len(v)-1]] = partsOfClientRoutes[j] 
					flag = true
				} else if partsOfClientRoutes[j] != v {

					strs := strings.Split(v, "{")
					if len(strs) > 0 {
						key := strs[1][:len(strs[1])-1] 
						paramMap[key] = partsOfClientRoutes[j][len(strs[0]):]
						flag = true
					} else {
						flag = false
						break
					}
				}
				flag = true
			}
		}
		if flag {
			if function, status := s.handlers[eachRegistRoutes]; status {
				return paramMap, function
			}
			break
		}
	}

	return nil, nil
}


func (s *Server) Response(body string) string {
	return "HTTP/1.1 200 OK\r\n" +
		"Content-Length: " + strconv.Itoa(len(body)) + lineCRLF +
		"Content-Type: text/html\r\n" +
		"Connection: close\r\n" +
		lineCRLF + body
}