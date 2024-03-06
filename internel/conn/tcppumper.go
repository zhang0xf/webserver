package conn

import (
	"bufio"
	"fmt"
	"nw/internel/common"
	"time"
)

const DefaultReadBufSize = 8 * 1024
const DefaultWriteBufSize = 16 * 1024

type TcpPumper struct {
	readBufSize        int
	mergedWriteBufSize int
	disableMergedWrite bool
	done               chan struct{}
}

func (tcpPumper *TcpPumper) readPump(conn *Conn) {
	context := conn.context
	stat := conn.stat
	scanner := bufio.NewScanner(conn.Conn)
	scanner.Buffer(make([]byte, tcpPumper.readBufSize), tcpPumper.readBufSize)
	scanner.Split(context.Splitter)
	for {
		if ok := scanner.Scan(); ok {
			data := scanner.Bytes()
			conn.Session.OnRecv(conn, data)
			if stat != nil {
				stat.AddTotalRecvStat(len(data), 1)
			}
		} else {
			fmt.Printf("read pump error: %v\n", scanner.Err())
			break
		}
	}
	close(tcpPumper.done) // 关闭writePump
	conn.Close()
}

func (tcpPumper *TcpPumper) writePump(conn *Conn) {
	tickerPing := time.NewTicker(common.PingPeriod)
	stat := conn.stat
	outChan := conn.msgChan.GetOutChan()
	buf := NewBuffer(tcpPumper.mergedWriteBufSize, !tcpPumper.disableMergedWrite)
loop: // label需要紧挨着for/switch/select
	for {
		select {
		case bytes := <-outChan:
			// outChan没有数据，writePump会死循环。 对于client，使用ctrl+c终止。对于server，尚未实现。
			conn.SetWriteDeadline(time.Now().Add(common.WriteWait))
			rb, count := buf.MergeBytes(bytes, outChan)
			_, err := conn.Conn.Write(rb)
			if err != nil {
				fmt.Printf("write pump error: %v\n", err.Error())
				break loop // 跳出for，而非select
			}
			if stat != nil {
				stat.AddTotalSendStat(len(rb), count)
				stat.SetTotalSendChanItemCount(len(outChan))
			}
		case <-tickerPing.C:
			// fmt.Println("message pump ticking ...")
		case <-tcpPumper.done:
			break loop
		}
	}
	tickerPing.Stop()
	conn.Close()
}
