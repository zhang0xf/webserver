package tcpserver

import (
	"fmt"
	"net"
	"nw"
	"nw/internel/conn"
	"nw/internel/hub"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Context  *nw.Context
	hub      *hub.Hub
	listener net.Listener
	conns    map[nw.Conn]bool //这里不能使用sessionid作为key, 因为Conn中关联的sessionid是在连接完成后才赋值(另见：tcpclient).
	mu       sync.RWMutex
}

// 为*Server实现nw.Server接口，使之成为nw.Server类型
func (server *Server) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Printf("server listen at addr( %s ) success.\n", l.Addr())
	server.listener = l
	go server.serve()
	return nil
}

func (server *Server) Stop() {
	if server.listener != nil {
		server.listener.Close()
	}

	server.hub.Stop()

	server.mu.Lock()
	defer server.mu.Unlock()

	for c := range server.conns {
		c.Close()
	}

	for c := range server.conns {
		c.Wait()
	}

	if len(server.conns) != 0 {
		server.conns = make(map[nw.Conn]bool)
	}
}

// Broadcast broadcast data to all active connections
func (server *Server) Broadcast(sessionIds []uint32, data []byte) {
	// server.hub.Broadcast(sessionIds, data)
	server.mu.RLock()
	defer server.mu.RUnlock()

	if len(sessionIds) == 0 {
		for conn := range server.conns {
			conn.Write(data)
		}
		return
	}

	ids := make(map[uint32]uint32)
	for i := 0; i < len(sessionIds); i++ {
		ids[sessionIds[i]] = sessionIds[i] // 构建map，使之高效.
	}

	for conn := range server.conns {
		if ids[conn.GetSession().GetId()] > 0 {
			conn.Write(data)
		}
	}
}

func (server *Server) GetActiveConnNum() int {
	return server.hub.GetActiveConnNum()
}

// *Server实现一些其他方法
func (server *Server) AddConn(conn nw.Conn) {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.conns[conn] = true
}

func (server *Server) DeleteConn(conn nw.Conn) {
	server.mu.Lock()
	defer server.mu.Unlock()
	delete(server.conns, conn)
}

func (server *Server) serve() {
	l := server.listener
	for {
		c, err := l.Accept() // 阻塞等待
		if err != nil {
			// if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			if _, ok := err.(net.Error); ok {
				fmt.Printf("temporary Accept() error : %s\n", err)
				runtime.Gosched()
				continue
			}
			// theres no direct way to detect server error because it is not exposed
			if !strings.Contains(err.Error(), "use of closed network connection") {
				fmt.Printf("listener.Accept() error : %s\n", err)
			}
			break
		}

		// 每次accept都建立一条新连接c
		conn := conn.NewTcpConn(c, server.Context, server.hub)

		// 开启一个线程处理新连接
		go func() {
			server.AddConn(conn)    // conn加入管理（非hub）
			go sendMessage(conn)    // debug：主动给客户端发送数据[可选]
			conn.ServeIO()          // 开启IO线程
			conn.Wait()             // 阻塞等待IO线程(需读，写线程全部关闭)
			server.DeleteConn(conn) // conn移除管理（非hub）
		}()
	}
}

func NewTcpServer(context *nw.Context) *Server {
	if context == nil || context.SessionCreator == nil || context.Splitter == nil {
		panic("NewTcpServer: context.SessionCreator is nil or context.Splitter is nil")
	}

	server := &Server{
		Context: context,
		hub:     hub.NewHub(context.IdleTimeAfterOpen),
		conns:   make(map[nw.Conn]bool),
	}

	return server
}

// debug
func sendMessage(conn *conn.Conn) {
	count := 0
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			in := conn.GetMsgChan().GetInChan()
			var testString string = "hello, I am server"
			in <- []byte(testString)
			count++
			if count > 5 {
				fmt.Println("exit debug goroutine : message finished.")
				return
			}
		}
	}
}
