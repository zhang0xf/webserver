package seqqueue

import "time"

type CallbackType int

const (
	REGULAR    CallbackType = iota + 1 // 常规的定时回调
	EXPIRE                             // 过期回调
	DONE                               // Queue结束回调
	SERVERDONE                         // 模块退出回调
	REACTIVE                           // Session销毁前，从后台激活
)

type Option struct {
	NonactiveDuration time.Duration // 多长时间不活跃停止队列
	CallbackInterval  time.Duration // 定时回调
	CheckInterval     time.Duration // 定时器
	OnCallback        func(q *SeqQueue, cType CallbackType)
}
