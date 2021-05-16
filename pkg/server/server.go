package server

import (
	"net"
	"sync"
)

type HandlerFunc func(con net.Conn)

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

	return nil
}