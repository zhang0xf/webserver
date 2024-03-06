package conn

import "container/list"

// reference: go-nonblockingchan

const msgChanDefaultSize = 100

type DefaultMessageChan struct {
	In   chan<- []byte
	Out  <-chan []byte
	done chan struct{}
	size int
}

func NewDefaultMessageChan(size int, done chan struct{}) *DefaultMessageChan {
	if size < 10 {
		size = 10
	}
	msgChan := make(chan []byte, size)
	return &DefaultMessageChan{
		In:   msgChan,
		Out:  msgChan,
		done: done,
		size: size,
	}
}

// 为*DefaultMessageChan类型实现MessageChan接口（即成为conn.MessageChan类型）
func (defaultMessageChan *DefaultMessageChan) GetInChan() chan<- []byte { return defaultMessageChan.In }

func (defaultMessageChan *DefaultMessageChan) GetOutChan() <-chan []byte {
	return defaultMessageChan.Out
}

func (defaultMessageChan *DefaultMessageChan) Len() int { return len(defaultMessageChan.In) }

func (defaultMessageChan *DefaultMessageChan) Size() int { return defaultMessageChan.size }

// Special type that mimics the behavior of a channel but does not block when items are sent.
// Items are stored internally until received.
// Closing the Send channel will cause the Recv channel to be closed after all items have been received.
type NonBlockingMessageChan struct {
	*DefaultMessageChan
	items     *list.List
	itemCount int
}

// Create a new non-blocking channel.
func NewNonBlockingMessageChan(size int, done chan struct{}) *NonBlockingMessageChan {
	if size < 10 {
		size = 10
	}
	var in = make(chan []byte, size)
	var out = make(chan []byte, size)
	var n = &NonBlockingMessageChan{
		DefaultMessageChan: &DefaultMessageChan{
			In:   in,
			Out:  out,
			done: done,
			size: size,
		},
		items: list.New(),
	}
	go n.run(in, out)
	return n
}

// 注：虽然实现上用到了多个channel，但对外提供的是一个抽象的messageChan
// 注：in和out参数为单向channel（in：从channel中读取[]byte，out：向channel中写入[]byte）
// Loop for buffering items between the Send and Recv channels until the Send channel is closed.
func (nonBlockingMessageChan *NonBlockingMessageChan) run(in <-chan []byte, out chan<- []byte) {
	for {
		if in == nil && nonBlockingMessageChan.items.Len() == 0 {
			close(out)
			break
		}
		var (
			outChan chan<- []byte
			outVal  []byte
		)
		if nonBlockingMessageChan.items.Len() > 0 {
			outChan = out
			outVal = nonBlockingMessageChan.items.Front().Value.([]byte)
		}
		select {
		case i, ok := <-in:
			if ok {
				nonBlockingMessageChan.items.PushBack(i)
				nonBlockingMessageChan.itemCount++
			} else {
				in = nil
			}
		case outChan <- outVal:
			nonBlockingMessageChan.items.Remove(nonBlockingMessageChan.items.Front())
			nonBlockingMessageChan.itemCount--
		case <-nonBlockingMessageChan.done:
			return
		}
	}
}

// 为*NonBlockingMessageChan实现conn.MessageChan接口（即成为conn.MessageChan类型）
// 注：NonBlockingMessageChan包含*DefaultMessageChan，部分接口不必重复实现。
func (nonBlockingMessageChan *NonBlockingMessageChan) Len() int {
	return nonBlockingMessageChan.itemCount
}

func (nonBlockingMessageChan *NonBlockingMessageChan) Size() int {
	return nonBlockingMessageChan.size
}

func NewMessageChan(useNoneBlockingChan bool, size int, done chan struct{}) MessageChan {
	if size <= 0 {
		size = msgChanDefaultSize
	}
	if useNoneBlockingChan {
		return NewNonBlockingMessageChan(size, done)
	} else {
		return NewDefaultMessageChan(size, done)
	}
}
