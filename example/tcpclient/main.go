package main

import (
	"fmt"
	"nw"
	"nw/common/tcpclient"
	"nw/internel/util"
	"time"
)

type MySession struct {
	Conn nw.Conn
}

func NewMySession(conn nw.Conn) nw.Session {
	return &MySession{Conn: conn}
}

// 为*MySession类型，实现Session接口，使之成为nw.Session类型
func (mySession *MySession) GetId() uint32 { return 0 }

func (mySession *MySession) GetConn() nw.Conn { return mySession.Conn }

func (mySession *MySession) OnOpen(conn nw.Conn) {}

func (mySession *MySession) OnClose(conn nw.Conn) {}

func (mySession *MySession) OnRecv(conn nw.Conn, data []byte) {
	fmt.Println("recving server data : " + string(data))
}

func Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	return len(data), data, nil
}

func main() {
	context := &nw.Context{
		SessionCreator: NewMySession,
		Splitter:       Split,
	}

	done := tcpclient.DialEx("127.0.0.1:7008", context, 3*time.Second)

	util.WaitForTerminate()

	close(done) // 通知拨号器退出
}
