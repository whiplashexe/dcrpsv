package ps5

import (
	"net"
	"strconv"
	"testing"
)

func TestPoll(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		c, _ := l.Accept()
		c.Write([]byte(`{"v":1,"running":true,"title_id":"PPSA01234","title_name":"Test"}` + "\n"))
		c.Close()
	}()
	port := l.Addr().(*net.TCPAddr).Port
	st, err := New("127.0.0.1", port).Poll()
	if err != nil {
		t.Fatal(err, strconv.Itoa(port))
	}
	if !st.Running || st.TitleID != "PPSA01234" || st.TitleName != "Test" {
		t.Fatalf("unexpected status: %+v", st)
	}
}

func TestPollUnreachable(t *testing.T) {
	if _, err := New("127.0.0.1", 1).Poll(); err == nil {
		t.Fatal("expected error")
	}
}
