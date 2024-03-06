package seqqueue

type MessageInterface interface {
	Process()
	Wait()
}
