package conn

import "bytes"

const MinMergedWriteBufSize = 100 * 1024

type Buffer struct {
	buffer   *bytes.Buffer
	buffSize int
	enabled  bool
}

func NewBuffer(buffSize int, enabled bool) *Buffer {
	if !enabled {
		return &Buffer{enabled: enabled}
	}
	return &Buffer{
		buffer:   bytes.NewBuffer(make([]byte, buffSize+2*1024)), // buffer cap is a little bigger than buffSize
		buffSize: buffSize,
	}
}

// Merge bytes and bytes from bytesChan（bytes may be the first elem in bytesChan）
func (buffer *Buffer) MergeBytes(bytes []byte, bytesChan <-chan []byte) ([]byte, int) {
	if !buffer.enabled || len(bytesChan) == 0 {
		return bytes, 1
	}

	count := 0
	buffer.buffer.Reset()

	for {
		count++
		buffer.buffer.Write(bytes)
		if len(bytesChan) == 0 || buffer.buffer.Len() >= buffer.buffSize {
			break
		}
		bytes = <-bytesChan
	}

	return buffer.buffer.Bytes(), count
}
