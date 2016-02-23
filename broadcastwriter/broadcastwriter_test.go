package broadcastwriter

import (
	"bytes"
	"sync"
	"testing"

	"fmt"
)

func TestBroadcastWriter(t *testing.T) {
	var wg sync.WaitGroup
	consumeListenerC := func(c <-chan []byte) <-chan string {
		outc := make(chan string)
		wg.Add(1)
		go func() {
			var buf bytes.Buffer
			for b := range c {
				buf.Write(b)
			}
			outc <- buf.String()
			wg.Done()
		}()
		return outc
	}

	bw := NewBroadcastWriter()

	l1 := bw.NewListener()
	c1 := consumeListenerC(l1)

	fmt.Fprintln(bw, "foo")

	l2 := bw.NewListener()
	c2 := consumeListenerC(l2)

	fmt.Fprintln(bw, "bar")

	l3 := bw.NewListener()
	c3 := consumeListenerC(l3)

	bw.Close()

	t.Logf("%q", <-c1)
	t.Logf("%q", <-c2)
	t.Logf("%q", <-c3)

	wg.Wait()
}
