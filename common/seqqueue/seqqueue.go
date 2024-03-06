package seqqueue

import (
	"nw/internel/common"
	"sync"
	"time"
)

var serverDone chan struct{} // 代表整个模块关闭
var DefaultOption *Option
var waitGroup common.WaitGroupWrapper

type SeqQueue struct {
	LastActiveTime   int64
	LastLeaveTime    int64 //用户下线离开时间。等于0时代表在线，大于0时代表超过10秒无api调用，做离线处理
	LastCallbackTime int64
	userId           int
	messages         chan MessageInterface
	done             chan struct{}
	quitted          chan struct{}
	stopOnce         sync.Once
	UserData         interface{}
}

// 由外部主动调用，关闭事件处理循环
func (seqQueue *SeqQueue) Stop(wait bool) {
	seqQueue.stopOnce.Do(func() {
		close(seqQueue.done)
	})
	if wait {
		<-seqQueue.quitted
	}
}

func (seqQueue *SeqQueue) SetActiveTime(activeTime time.Time) {
	seqQueue.LastActiveTime = activeTime.Unix()
	if seqQueue.LastLeaveTime > 0 { //离线后又重新上线了（从后台切换回来）
		DefaultOption.OnCallback(seqQueue, REACTIVE)
	}
}

func New(userData interface{}, userId int) *SeqQueue {
	now := time.Now().Unix()
	seqQueue := &SeqQueue{
		userId:           userId,
		messages:         make(chan MessageInterface, 5),
		done:             make(chan struct{}),
		quitted:          make(chan struct{}),
		UserData:         userData,
		LastActiveTime:   now,
		LastCallbackTime: now,
	}
	waitGroup.Wrap(func() { seqQueue.messageLoop() })
	return seqQueue
}

func (seqQueue *SeqQueue) messageLoop() {
	var NONACTIVE_DURATION = int64(DefaultOption.NonactiveDuration / time.Second) //以秒为单位
	var CALLBACK_INTERVAL = int64(DefaultOption.CallbackInterval / time.Second)   //以秒为单位
	// CALLBACK_INTERVAL = 30                                                       //TODO
	var ticker = time.NewTicker(DefaultOption.CheckInterval)
	var now int64
	for {
		select {
		case <-serverDone:
			DefaultOption.OnCallback(seqQueue, SERVERDONE)
			goto exit
		case <-seqQueue.done:
			DefaultOption.OnCallback(seqQueue, DONE)
			goto exit
		case <-ticker.C:
			now = time.Now().Unix()
			if now-seqQueue.LastActiveTime >= NONACTIVE_DURATION+2 {
				DefaultOption.OnCallback(seqQueue, EXPIRE)
				goto exit
			} else if now-seqQueue.LastCallbackTime >= CALLBACK_INTERVAL {
				DefaultOption.OnCallback(seqQueue, REGULAR)
				seqQueue.LastCallbackTime = now
			}
		case message := <-seqQueue.messages:
			message.Process() // 任何发送给SeqQueue执行的消息需实现MessageInterface接口
		}
	}
exit:
	ticker.Stop()
	close(seqQueue.quitted)
}
