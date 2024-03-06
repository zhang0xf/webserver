package conn

import (
	"fmt"
	"nw/internel/common"
	"time"

	"github.com/gorilla/websocket"
)

type WsPumper struct {
	wsConn             *websocket.Conn
	mergedWriteBufSize int
	disableMergedWrite bool
	waitPong           bool
	done               chan struct{}
}

// 为*WsPumper实现IOPumper接口，使其成为IOPumper类型
func (wsPumper *WsPumper) readPump(conn *Conn) {
	wsConn := wsPumper.wsConn
	context := conn.context
	stat := conn.stat

	if context.MaxMessageSize > 0 {
		wsConn.SetReadLimit(int64(context.MaxMessageSize))
	}

	// client must send ping, or conn will be closed after PongWait
	wsConn.SetReadDeadline(time.Now().Add(common.PongWait))

	wsConn.SetPongHandler(func(string) error {
		wsPumper.waitPong = false
		wsConn.SetReadDeadline(time.Now().Add(common.PongWait))
		return nil
	})

	for {
		_, data, err := wsConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure) {
				fmt.Printf("wsserver read error: %s\n", err.Error())
			}
			if wsPumper.waitPong {
				fmt.Println("没有等到客户端的pong包")
			}
			break
		}

		conn.Session.OnRecv(conn, data)

		if stat != nil {
			stat.AddTotalRecvStat(len(data), 1)
		}
	}

	conn.Close()
}

func (wsPumper *WsPumper) writePump(conn *Conn) {
	wsConn := wsPumper.wsConn
	pingTicker := time.NewTicker(common.PingPeriod)
	stat := conn.stat
	outChan := conn.msgChan.GetOutChan()
	buff := NewBuffer(wsPumper.mergedWriteBufSize, !wsPumper.disableMergedWrite)
loop: // label需要紧挨着for/switch/select
	for {
		select {
		case data := <-outChan:
			// outChan没有数据，以下逻辑不会执行。
			wsConn.SetWriteDeadline(time.Now().Add(common.WriteWait))
			rb, count := buff.MergeBytes(data, outChan)
			err := wsConn.WriteMessage(websocket.BinaryMessage, rb)
			if err != nil {
				fmt.Printf("wsserver write error: %s\n", err.Error())
				break loop
			}
			if stat != nil {
				stat.AddTotalSendStat(len(rb), count)
				stat.SetTotalSendChanItemCount(len(outChan))
			}
		case <-pingTicker.C:
			// 1.定时向客户端发送心跳包
			// 2.接受到客户端心跳包waitPong会被置为false
			wsConn.SetWriteDeadline(time.Now().Add(common.WriteWait))
			err := wsConn.WriteMessage(websocket.PingMessage, []byte{})
			wsPumper.waitPong = true
			if err != nil {
				fmt.Printf("wsserver send ping error: %s\n", err.Error())
				break loop
			}
		case <-wsPumper.done:
			fmt.Println("wsconn close done.")
			break loop
		}
	}
	pingTicker.Stop()
	conn.Close()
}
