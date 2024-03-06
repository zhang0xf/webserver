package conn

import (
	"errors"
	"nw"
	"nw/internel/hub"
	"time"

	"github.com/gorilla/websocket"
)

// wrap websocket.Conn to adopt net.Conn
type WsConn struct {
	*websocket.Conn // <- F12
}

// 为*WsConn实现net.Conn接口，使其成为net.Conn类型
func (wsConn *WsConn) Read(b []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (wsConn *WsConn) Write(data []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (wsConn *WsConn) SetDeadline(t time.Time) error {
	return errors.New("not implemented")
}

func NewWsConn(c *websocket.Conn, context *nw.Context, hub *hub.Hub) *Conn { // client hub is nil
	done := make(chan struct{})
	msgChan := NewMessageChan(context.UseNoneBlockingChan, context.ChanSize, done)
	conn := newConn(&WsConn{Conn: c}, context, msgChan, hub, done)

	mergedWriteBufSize := MinMergedWriteBufSize
	if context.MergedWriteBufferSize > mergedWriteBufSize {
		mergedWriteBufSize = context.MergedWriteBufferSize
	}

	if context.MaxMessageSize > 0 && mergedWriteBufSize > context.MaxMessageSize {
		mergedWriteBufSize = context.MaxMessageSize
	}

	conn.ioPumper = &WsPumper{
		wsConn:             c,
		mergedWriteBufSize: mergedWriteBufSize,
		disableMergedWrite: context.DisableMergedWrite,
		done:               done,
	}

	return conn
}
