package main

import (
	"bytes"
	"fmt"
	"nw"
	"nw/common/wsserver"
	"nw/internel/util"
)

// wsserver(websocket server)
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
	fmt.Println("recving data : " + string(data))
}

func Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	n := bytes.IndexByte(data, '-')
	if n > 0 {
		return n + 1, data[0:n], nil
	}

	return 0, nil, nil
}

func main() {
	context := &nw.Context{
		SessionCreator: NewMySession,
		Splitter:       Split,
	}

	server := wsserver.NewServer(context)

	if err := server.Start(":7008"); err != nil {
		fmt.Println("server start error : " + err.Error())
	}

	util.WaitForTerminate()

	server.Stop()
}
