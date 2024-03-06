package conn

type MessageChan interface {
	GetInChan() chan<- []byte
	GetOutChan() <-chan []byte
	Len() int
	Size() int
}
