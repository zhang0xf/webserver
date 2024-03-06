package conn

import (
	"net"
	"nw"
	"nw/internel/hub"
)

func NewTcpConn(c net.Conn, context *nw.Context, hub *hub.Hub) *Conn {
	done := make(chan struct{})
	msgChan := NewMessageChan(context.UseNoneBlockingChan, context.ChanSize, done)
	conn := newConn(c, context, msgChan, hub, done)

	readBufSize := DefaultReadBufSize
	if context.ReadBufferSize > 0 {
		readBufSize = context.ReadBufferSize
	}

	writeBufSize := DefaultWriteBufSize
	if context.WriteBufferSize > 0 {
		writeBufSize = context.WriteBufferSize
	}

	mergedWriteBufSize := MinMergedWriteBufSize
	if context.MergedWriteBufferSize > mergedWriteBufSize {
		mergedWriteBufSize = context.MergedWriteBufferSize
	}

	c.(*net.TCPConn).SetReadBuffer(readBufSize)
	c.(*net.TCPConn).SetWriteBuffer(writeBufSize)

	conn.ioPumper = &TcpPumper{
		done:               done,
		readBufSize:        readBufSize,
		mergedWriteBufSize: mergedWriteBufSize,
		disableMergedWrite: context.DisableMergedWrite,
	}

	return conn
}
