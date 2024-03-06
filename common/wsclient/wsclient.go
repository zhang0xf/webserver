package wsclient

import (
	"net"
	"nw"
	"nw/internel/conn"

	"github.com/gorilla/websocket"
)

// 使用websocket提供的接口，创建websocket链接。（或使用浏览器，另见：index.html）
func Dial(addr string, context *nw.Context) (net.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil, err
	}
	conn := conn.NewWsConn(c, context, nil)
	conn.ServeIO()
	return conn, nil
}
