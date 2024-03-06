package wsserver

import (
	"fmt"
	"net/http"
	"nw"
	"nw/common/httpserver"
	"nw/internel/conn"
	"nw/internel/hub"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	nw.Server
	Context    *nw.Context
	hub        *hub.Hub
	upgrader   *websocket.Upgrader
	httpServer *httpserver.HttpServer
	conns      map[uint32]nw.Conn
	mu         sync.RWMutex
}

// 为*Server实现nw.Server接口，使其成为nw.Server类型
// Start start websocket server, and start default http server if addr is not empty
func (server *Server) Start(addr string) error {
	if len(addr) > 0 {
		httpServer := httpserver.NewHttpServer()
		err := httpServer.Start(addr)
		if err != nil {
			return err
		}
		httpServer.Handle("/ws", server)
		server.httpServer = httpServer
	}
	return nil
}

// Stop stop websocket server, and the underline default http server
func (server *Server) Stop() {
	if server.httpServer != nil {
		server.httpServer.Stop()
	}

	server.hub.Stop()

	server.mu.Lock()
	defer server.mu.Unlock()
	server.conns = make(map[uint32]nw.Conn)
}

// Broadcast broadcast data to all active connections
func (server *Server) Broadcast(sessionIds []uint32, data []byte) {
	//server.hub.Broadcast(sessionIds, data)
	server.mu.RLock()
	defer server.mu.RUnlock()

	if len(sessionIds) == 0 {
		for _, conn := range server.conns {
			conn.Write(data)
		}
		return
	}

	for i := 0; i < len(sessionIds); i++ {
		conn := server.conns[sessionIds[i]]
		if conn != nil {
			conn.Write(data)
		}
	}
}

// GetActiveConnNum get count of active connections
func (server *Server) GetActiveConnNum() int {
	return server.hub.GetActiveConnNum()
}

// 为*Server实现http.Handler接口，使其成为http.Handler类型，另见Start()函数。
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("wsserver.ServeHTTP upgrade error: %s\n", err)
		return
	}

	conn := conn.NewWsConn(c, server.Context, server.hub)

	go func() {
		conn.SetRemoteAddr(server.getRemoteAddr(r)) //should called before conn.ServeIO()
		server.AddConn(conn)
		conn.ServeIO()
		conn.Wait()
		server.DeleteConn(conn)

	}()
}

// 为*Server实现其他一些方法
func (server *Server) getRemoteAddr(r *http.Request) string {
	addr := r.Header.Get("Remote_addr")
	if addr == "" {
		addr = r.RemoteAddr
	}
	fmt.Printf("get remote addr : %v\n", addr)
	return addr
}

func (server *Server) AddConn(conn nw.Conn) {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.conns[conn.GetSession().GetId()] = conn
}

func (server *Server) DeleteConn(conn nw.Conn) {
	server.mu.Lock()
	defer server.mu.Unlock()
	delete(server.conns, conn.GetSession().GetId())
}

// NewServer create a new websocket server
func NewServer(context *nw.Context) *Server {
	if context == nil || context.SessionCreator == nil {
		panic("wsserver.NewServer: context is nil or context.SessionCreator is nil")
	}

	server := &Server{
		Context: context,
		hub:     hub.NewHub(context.IdleTimeAfterOpen),
		conns:   make(map[uint32]nw.Conn),
	}

	server.upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024 * 2,
		WriteBufferSize: 1024 * 4,
		CheckOrigin:     func(r *http.Request) bool { return true }, // disable check
	}

	if context.ReadBufferSize > 0 {
		server.upgrader.ReadBufferSize = context.ReadBufferSize
	}

	if context.WriteBufferSize > 0 {
		server.upgrader.WriteBufferSize = context.WriteBufferSize
	}

	return server
}
