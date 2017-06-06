package main
import (
	"io"
	"os"
	"time"
	"strings"
	. "github.com/wouck/ringbuf"
)
func main() {
	rb := MakeRingBuf(4096)
	go io.Copy(os.Stdout, rb)
	for {
		time.Sleep(1*time.Second)
		rb.Write([]byte(strings.Repeat("a", 4097)))
		time.Sleep(5*time.Second)
	}

	wait := make(chan bool)
	<-wait
}
