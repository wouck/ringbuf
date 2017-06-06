package ringbuf
import (
	"fmt"
	"io"
	"sync"
)

type RingBuf struct {
	buf []byte
	rpos int
	wpos int
	isclosed bool
	isreadwait bool
	iswritewait bool
	readwait sync.WaitGroup
	writewait sync.WaitGroup
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func MakeRingBuf(size int) *RingBuf {
	return &RingBuf{make([]byte, size+1), 0, 0, false, false, false, sync.WaitGroup{}, sync.WaitGroup{}}
}

func (rb *RingBuf) readWait() {
	rb.readwait.Add(1)
	rb.isreadwait = true
	rb.readwait.Wait()
}

func (rb *RingBuf) readTrigger() {
	if rb.isreadwait {
		rb.isreadwait = false
		rb.readwait.Done()
		fmt.Printf("** Wake Reader!\n")
	}
}

func (rb *RingBuf) writeWait() {
	rb.writewait.Add(1)
	rb.iswritewait = true
	rb.writewait.Wait()
}

func (rb *RingBuf) writerTrigger() {
	if rb.iswritewait {
		rb.iswritewait = false
		rb.writewait.Done()
		fmt.Printf("** Wake Writer!\n")
	}
}

func (rb *RingBuf) Dump() {
	fmt.Printf("bufalc:%d\n", len(rb.buf))
	fmt.Printf("rpos  :%d\n", rb.rpos)
	fmt.Printf("wpos  :%d\n", rb.wpos)
}

func (rb *RingBuf) isClosed() bool{
	return rb.isclosed
}

func (rb *RingBuf) Close() {
	rb.isclosed = true
	rb.readTrigger()
}

func (rb *RingBuf) isEmpty()bool {
	if rb.getDataSize() == 0{
		return true
	}

	return false
}

func (rb *RingBuf) isFull()bool {
	if rb.getFreeSpace() == 0 {
		return true
	}
	return false
}

func (rb *RingBuf) Read(p []byte)(int, error) {
	plen := len(p)

	if rb.isEmpty() && !rb.isClosed(){
		fmt.Printf("[Read]: Buf is empty, wait...\n")
		rb.readWait()
	}

	size := min(plen, rb.getDataSize())
	fmt.Printf("[Read]: read buffer: %d bytes\n", plen)
	fmt.Printf("[Read]: read chunk: %d bytes\n", size)
	read := rb.readChunk(p[:size])

	if read > 0 {
		rb.writerTrigger()
	}

	if read == 0 {
		return read, io.EOF
	}

	return read, nil
}

func (rb *RingBuf) getDataSize() int{
	return len(rb.buf) - rb.getFreeSpace() - 1
}

func (rb *RingBuf) getFreeSpace() int{
	buflen := len(rb.buf)
	if rb.wpos >= rb.rpos {
		return buflen - rb.wpos + rb.rpos - 1
	}

	return rb.rpos - rb.wpos - 1
}

func (rb *RingBuf) readChunk(p []byte) int{
	buflen := len(rb.buf)
	plen := len(p)
	read := 0

	for plen > 0 {
		end := min(buflen, rb.rpos + plen)
		n := copy(p, rb.buf[rb.rpos:end])
		p = p[n:]
		read += n
		plen -= n
		rb.rpos = (rb.rpos + n) % buflen
	}

	return read
}

func (rb *RingBuf) writeChunk(p []byte) int{
	write := 0
	buflen := len(rb.buf)
	plen := len(p)

	for plen > 0 {
		end := min(buflen, rb.wpos + plen)
		n := copy(rb.buf[rb.wpos:end], p)
		p = p[n:]
		write += n
		plen -= n
		rb.wpos = (rb.wpos + n) % buflen
	}

	return write
}

func (rb *RingBuf) Write(p []byte)(int, error) {
	write := 0
	plen := len(p)

	fmt.Printf("[Write]: Call Write(%d bytes)\n", plen)
	for plen > 0 {
		free := rb.getFreeSpace()
		chunk := min(free, plen)
		if chunk == 0 {
			fmt.Printf("[Write]: Buf is full wpos:%d, rpos:%d, left:%d\n", rb.wpos, rb.rpos, plen)
			rb.writeWait()
			continue //rpos changed, re-calc chunk size
		}

		n := rb.writeChunk(p[:chunk])
		p = p[n:]
		plen -= n
		write += n
		fmt.Printf("[Write]: chunk %d\n", n)

		rb.readTrigger()
	}

	fmt.Printf("[Write]: Finish %d\n", write)
	return write, nil
}

