package httpclient

import (
	"net"
	"time"
)

type TimeoutConn struct {
	net.Conn
	timeout time.Duration
}

// 为*TimeoutConn实现net.Conn部分接口。
func (timeoutConn *TimeoutConn) Read(b []byte) (n int, err error) {
	timeoutConn.SetReadDeadline(time.Now().Add(timeoutConn.timeout))
	return timeoutConn.Conn.Read(b)
}

func (timeoutConn *TimeoutConn) Write(b []byte) (n int, err error) {
	timeoutConn.SetWriteDeadline(time.Now().Add(timeoutConn.timeout))
	return timeoutConn.Conn.Write(b)
}

func NewTimeoutConn(conn net.Conn, timeout time.Duration) *TimeoutConn {
	return &TimeoutConn{conn, timeout}
}
