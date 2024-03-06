package hub

import (
	"fmt"
	"nw"
	"sync"
	"sync/atomic"
	"time"
)

const (
	chanTypeInit = iota + 1
	chanTypeActive
	chanTypeClose
)

type broadcastMessage struct {
	sessionIds []uint32
	data       []byte
}

type connChanData struct {
	conn nw.Conn
	typ  int
}

//hub:中心，中转站，枢纽
type Hub struct {
	broadcastChan chan *broadcastMessage
	connChan      chan *connChanData
	initConns     map[nw.Conn]bool   // connected but not authed Conn
	activeConns   map[uint32]nw.Conn // authed Conn
	activeConnNum int                // the number of activeConns
	idleDuration  time.Duration
	done          chan struct{}
	closed        int32
	wg            sync.WaitGroup
}

func NewHub(idleDuration time.Duration) *Hub {
	hub := &Hub{
		connChan:      make(chan *connChanData, 100),
		broadcastChan: make(chan *broadcastMessage, 1000),
		idleDuration:  idleDuration,
		initConns:     make(map[nw.Conn]bool),
		activeConns:   make(map[uint32]nw.Conn),
		done:          make(chan struct{}),
	}
	hub.wg.Add(1)
	go func() {
		hub.run()
		hub.wg.Done()
	}()
	return hub
}

func (hub *Hub) pushChanData(conn nw.Conn, typ int) {
	select {
	case hub.connChan <- &connChanData{conn, typ}:
	case <-hub.done:
	}
}

func (hub *Hub) ActivateConn(conn nw.Conn) {
	closed := atomic.LoadInt32(&hub.closed) == 1
	if !closed && hub.idleDuration > 0 {
		hub.pushChanData(conn, chanTypeActive)
	}
}

func (hub *Hub) GetActiveConnNum() int {
	return hub.activeConnNum
}

func (hub *Hub) run() {
	var ticker = time.NewTicker(3 * time.Second)
	defer func() { // defer：延后执行
		ticker.Stop()
	}()

	for {
		select {
		case data := <-hub.connChan:
			conn := data.conn
			switch data.typ {
			case chanTypeInit:
				hub.initConns[conn] = true
				fmt.Println("hub : init a conn")
			case chanTypeActive:
				delete(hub.initConns, conn)
				hub.activeConns[conn.GetSession().GetId()] = conn
				fmt.Println("hub : active a conn")
			case chanTypeClose:
				delete(hub.initConns, conn)
				delete(hub.activeConns, conn.GetSession().GetId())
				fmt.Println("hub : close a conn")
			}
		case message := <-hub.broadcastChan:
			fmt.Println("hub : broadcast message")
			if len(message.sessionIds) == 0 {
				for _, conn := range hub.activeConns {
					conn.Write(message.data)
				}
			} else {
				for _, id := range message.sessionIds {
					conn := hub.activeConns[id]
					if conn != nil {
						conn.Write(message.data)
					}
				}
			}
		case <-ticker.C:
			// fmt.Println("hub : check active state")
			hub.activeConnNum = len(hub.activeConns)
			if hub.idleDuration > 0 && len(hub.initConns) > 0 {
				now := time.Now()
				for conn := range hub.initConns {
					if now.Sub(conn.GetConnTime()) > hub.idleDuration { // 闲置（过期）时间
						fmt.Println("hub ticker : close expired conn")
						delete(hub.initConns, conn)
						conn.Close()
					}
				}
			}
		case <-hub.done:
			hub.clear()
			fmt.Println("hub : close hub.")
			return
		}
	}
}

func (hub *Hub) Broadcast(sessionIds []uint32, data []byte) {
	if len(hub.broadcastChan) > 990 {
		fmt.Println("hub Broadcast: not enouth broadcastChan, sessionIds:", sessionIds)
	}
	hub.broadcastChan <- &broadcastMessage{sessionIds: sessionIds, data: data}
}

func (hub *Hub) AddConn(conn nw.Conn) {
	closed := atomic.LoadInt32(&hub.closed) == 1
	if !closed {
		if hub.idleDuration > 0 {
			hub.pushChanData(conn, chanTypeInit)
		} else {
			hub.pushChanData(conn, chanTypeActive)
		}
	}
}

func (hub *Hub) RemoveConn(conn nw.Conn) {
	closed := atomic.LoadInt32(&hub.closed) == 1
	if !closed {
		hub.pushChanData(conn, chanTypeClose)
	}
}

func (hub *Hub) clear() {
	conns := make(map[nw.Conn]bool)
	n := len(hub.connChan)
	for i := 0; i < n; i++ {
		data := <-hub.connChan // 队列中的conn
		if data.typ != chanTypeClose {
			conns[data.conn] = true
		}
	}
	for conn := range hub.initConns { // 已初始化的conn
		conns[conn] = true
	}
	for _, conn := range hub.activeConns { // 已经激活的conn
		conns[conn] = true
	}
	for conn := range conns {
		conn.Close()
	}
	for conn := range conns {
		conn.Wait()
	}
}

func (hub *Hub) Stop() {
	if atomic.CompareAndSwapInt32(&hub.closed, 0, 1) {
		close(hub.done) // 通知hub退出
		hub.wg.Wait()   // 阻塞等待（另见：NewHub）
	}
}
