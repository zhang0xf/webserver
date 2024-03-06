package netstat

type StatData struct {
	messagesRecv  int64 // message count received
	messagesSend  int64 // message count sent
	bytesRecv     int64 // byte count received
	bytesSend     int64 // byte count sent
	sendChanItems int64 // buffered item count of send chan
	recvChanItems int64 // buffered item count of recv chan
}

func (statData *StatData) GetMessagesRecv() int { return int(statData.messagesRecv) }

func (statData *StatData) GetMessagesSend() int { return int(statData.messagesSend) }

func (statData *StatData) GetBytesRecv() int { return int(statData.bytesRecv) }

func (statData *StatData) GetBytesSend() int { return int(statData.bytesSend) }

func (statData *StatData) GetSendChanItemCount() int { return int(statData.sendChanItems) }

func (statData *StatData) GetRecvChanItemCount() int { return int(statData.recvChanItems) }
