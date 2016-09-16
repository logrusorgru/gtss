package gtss

import (
	"net"
	"sync"
	"testing"
	"time"
)

const address = "127.0.0.1:9001"

// TODO: refactor tests

func Test_all(t *testing.T) {
	s := &Server{
		Net:  "tcp",
		Addr: address,
	}
	done := make(chan struct{})
	go func() {
		err := s.ListernAndServe()
		if err != nil {
			t.Error(err)
		}
		close(done)
	}()
	time.Sleep(10 * time.Second)
	echoClient(t)
	if err := s.Close(); err != nil {
		t.Error(err)
	}
	<-done
}

func echoClient(t *testing.T) {
	wg := new(sync.WaitGroup)
	client, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("SOME DATA")
	wg.Add(2)
	go func() {
		defer wg.Done()
		if _, err := client.Write(data); err != nil {
			t.Error(err)
		}
	}()
	go func() {
		defer wg.Done()
		b := make([]byte, len(data))
		if _, err := io.ReadFull(client, b); err != nil {
			t.Error(err)
		}
		if string(b) != string(data) {
			t.Error("wrong response: '%s'", string(b))
		}
	}()
	wg.Wait()
}
