package netstat

import "sync/atomic"

type NetStat struct {
	current   *StatData
	max       *StatData
	total     *StatData
	lastTotal *StatData
}

func NewNetStat() *NetStat {
	s := &NetStat{
		current:   new(StatData),
		max:       new(StatData),
		total:     new(StatData),
		lastTotal: new(StatData),
	}

	if atomic.CompareAndSwapInt32(&isRunning, 0, 1) {
		manager = newStatManager()
		manager.run()
	}

	<-manager.start
	return s
}

func (netStat *NetStat) Start() { manager.startChan <- netStat }

func (netStat *NetStat) Stop() { manager.stopChan <- netStat }

func (netStat *NetStat) GetCurrent() *StatData { return netStat.current }

func (netStat *NetStat) GetMax() *StatData { return netStat.max }

func (netStat *NetStat) GetTotal() *StatData { return netStat.total }

func (netStat *NetStat) doCalc() {
	total, current, max, lastTotal := netStat.total, netStat.current, netStat.max, netStat.lastTotal
	messagesSend := atomic.LoadInt64(&total.messagesSend)
	messagesRecv := atomic.LoadInt64(&total.messagesRecv)
	bytesSend := atomic.LoadInt64(&total.bytesSend)
	bytesRecv := atomic.LoadInt64(&total.bytesRecv)

	current.messagesSend = messagesSend - lastTotal.messagesSend
	current.messagesRecv = messagesRecv - lastTotal.messagesRecv
	current.bytesSend = bytesSend - lastTotal.bytesSend
	current.bytesRecv = bytesRecv - lastTotal.bytesRecv

	if current.messagesSend > max.messagesSend {
		max.messagesSend = current.messagesSend
	}
	if current.messagesRecv > max.messagesRecv {
		max.messagesRecv = current.messagesRecv
	}
	if current.bytesSend > max.bytesSend {
		max.bytesSend = current.bytesSend
	}
	if current.bytesRecv > max.bytesRecv {
		max.bytesSend = current.bytesRecv
	}

	lastTotal.messagesSend = messagesSend
	lastTotal.messagesRecv = messagesRecv
	lastTotal.bytesSend = bytesSend
	lastTotal.bytesRecv = bytesRecv
}

func (netStat *NetStat) AddTotalSendStat(msgLen int, msgCount int) {
	data := netStat.total
	atomic.AddInt64(&data.bytesSend, int64(msgLen))
	atomic.AddInt64(&data.messagesSend, int64(msgCount))
}

func (netStat *NetStat) AddTotalRecvStat(msgLen int, msgCount int) {
	data := netStat.total
	atomic.AddInt64(&data.bytesRecv, int64(msgLen))
	atomic.AddInt64(&data.messagesRecv, int64(msgCount))
}

func (netStat *NetStat) SetTotalSendChanItemCount(chanItemCount int) {
	data := netStat.total
	count := atomic.LoadInt64(&data.sendChanItems)
	if int64(chanItemCount) > count {
		atomic.StoreInt64(&data.sendChanItems, int64(chanItemCount))
	}
}

func (netStat *NetStat) SetTotalRecvChanItemCount(chanItemCount int) {
	data := netStat.total
	count := atomic.LoadInt64(&data.recvChanItems)
	if int64(chanItemCount) > count {
		atomic.StoreInt64(&data.recvChanItems, int64(chanItemCount))
	}
}
