package tcpclient

import (
	"fmt"
	"net"
	"nw"
	"nw/internel/conn"
	"runtime"
	"time"
)

func DialEx(addr string, context *nw.Context, retryWait time.Duration) chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		retryChan := make(chan bool, 1)
		retryChan <- true
		needWait := false
		for {
			select {
			case <-doneChan:
				return
			case <-retryChan:
				if needWait && retryWait > 0 {
					select {
					case <-time.NewTimer(retryWait).C:
						fmt.Printf("after %f seconds, retry dialing addr : %s\n", retryWait.Seconds(), addr)
					case <-doneChan:
						return
					}
				} else {
					needWait = true
				}

				// 发起连接（拨号）
				c, err := net.DialTimeout("tcp", addr, 5*time.Second)
				if err != nil {
					fmt.Printf("connect error: %s\n", err.Error())
					retryChan <- true
					continue
				}

				// 创建连接成功
				nwConn := conn.NewTcpConn(c, context, nil)
				go sendMessage(nwConn) // debug：发送数据
				nwConn.ServeIO()       // 开启IO线程
				nwConn.Wait()          // this will block the for loop, the application layer need to close the conn to let it go along.
				runtime.Gosched()      // 让当前goroutine让出cpu，好让其他goroutine获得执行机会，同时当前goroutine也会在未来某个时间点再次执行。
				retryChan <- true      // 重新创建连接（IO线程失败会关闭连接）
			}
		}
	}()
	return doneChan
}

// debug
func sendMessage(conn *conn.Conn) {
	count := 0
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			in := conn.GetMsgChan().GetInChan()
			var testString string = "hello, I am client"
			in <- []byte(testString)
			count++
			if count > 5 {
				fmt.Println("exit debug goroutine : message finished.")
				return
			}
		}
	}
}
