package conn

import (
	"errors"
	"fmt"
	"net"
	"nw"
	"nw/internel/hub"
	"nw/internel/netstat"
	"sync"
	"time"
)

// Conn拥有嵌套了net.Conn，包含了net.Conn所有接口（即成为net.Conn类型）
type Conn struct {
	net.Conn
	msgChan  MessageChan
	inChan   chan<- []byte
	context  *nw.Context
	hub      *hub.Hub // 对于client而言，可能不需要hub
	ioPumper IOPumper
	wg       sync.WaitGroup
	Session  nw.Session
	UserData interface{}
	ConnTime time.Time
	stat     *netstat.NetStat
	done     chan struct{}
	// closeOnce  sync.Once
	remoteAddr *netstat.NetAddr
}

// 为*Conn实现nw.Conn接口（即成为nw.conn类型）
func (conn *Conn) SetUserData(userData interface{}) { conn.UserData = userData }

func (conn *Conn) GetUserData() interface{} { return conn.UserData }

func (conn *Conn) GetSession() nw.Session { return conn.Session }

func (conn *Conn) GetConnTime() time.Time { return conn.ConnTime }

func (conn *Conn) Activate() {
	if conn.hub != nil {
		conn.hub.ActivateConn(conn)
	}
}

func (conn *Conn) GraceClose() error {
	select {
	case conn.inChan <- nil:
		return nil
	case <-conn.done:
		return errors.New("connection already closed")
	}
}

func (conn *Conn) Wait() {
	conn.wg.Wait()
}

func (conn *Conn) GetContext() interface{} { return conn.context }

// 为*Conn实现一些方法（包括net.Conn中接口和自身方法）
func (conn *Conn) ServeIO() {
	conn.wg.Add(1)
	go func() {
		conn.ioPumper.writePump(conn) // 只有IO失败(写超时会失败)，或通过done主动关闭，才会退出 writePump，并且关闭连接
		fmt.Println("exit writePump goroutine")
		conn.wg.Done()
	}()

	conn.wg.Add(1)
	go func() {
		if conn.hub != nil {
			conn.hub.AddConn(conn)
		}
		conn.Session.OnOpen(conn)
		conn.ioPumper.readPump(conn) // IO失败，就会退出 readPump，并且关闭连接
		fmt.Println("exit readPump goroutine")
		conn.Session.OnClose(conn)
		if conn.hub != nil {
			conn.hub.RemoveConn(conn)
		}
		if conn.stat != nil {
			conn.stat.Stop()
		}
		conn.wg.Done()
	}()
}

func (conn *Conn) GetMsgChan() MessageChan { return conn.msgChan }

func (conn *Conn) GetRemoteAddr() net.Addr { return conn.remoteAddr }

func (conn *Conn) SetRemoteAddr(addr string) {
	netAddr := &netstat.NetAddr{}
	netAddr.SetAddr(addr)
	conn.remoteAddr = netAddr
}

func newConn(c net.Conn, context *nw.Context, msgChan MessageChan, hub *hub.Hub, done chan struct{}) *Conn {
	conn := &Conn{
		Conn:     c,
		msgChan:  msgChan,
		inChan:   msgChan.GetInChan(),
		context:  context,
		hub:      hub,
		ConnTime: time.Now(),
		done:     done,
	}

	if context.SessionCreator != nil {
		fmt.Println("create a session for conn")
		conn.Session = context.SessionCreator(conn)
	}

	if context.EnableStatistics {
		conn.stat = netstat.NewNetStat()
		conn.stat.Start()
	}

	return conn
}
