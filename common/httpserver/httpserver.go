package httpserver

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type HttpServer struct {
	listener net.Listener
	closing  bool
	Router   *mux.Router
	wg       sync.WaitGroup
}

func NewHttpServer() *HttpServer {
	server := &HttpServer{Router: mux.NewRouter()}
	server.Router.StrictSlash(true)
	return server
}

func (httpServer *HttpServer) Start(netAddr string) error {
	l, err := net.Listen("tcp", netAddr)
	if err != nil {
		return err
	}
	httpServer.listener = l

	httpServer.wg.Add(1)
	go func() {
		httpServer.serve()
		httpServer.wg.Done()
	}()
	return nil
}

func (httpServer *HttpServer) serve() {
	err := http.Serve(httpServer.listener, httpServer.Router)
	fmt.Println("HttpServer http.Serve error: " + err.Error())
	if !httpServer.closing && err != nil {
		fmt.Println("http serve error: " + err.Error())
	}
}

func (httpServer *HttpServer) Stop() {
	httpServer.closing = true
	if httpServer.listener != nil {
		httpServer.listener.Close()
	}
	httpServer.wg.Wait()
}

// 回调注册（两种方式）
func (httpServer *HttpServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	httpServer.Router.HandleFunc(pattern, handler)
}

func (httpServer *HttpServer) Handle(pattern string, handler http.Handler) {
	httpServer.Router.Handle(pattern, handler)
}
