package websocket_test

import (
	"bufio"
	"context"
	gonet "net"
	"net/http"
	"testing"
	"time"

	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/testing/servers/tcp"
	"github.com/xtls/xray-core/transport/internet"
	"github.com/xtls/xray-core/transport/internet/stat"
	. "github.com/xtls/xray-core/transport/internet/websocket"
)

// A CDN keeps its connections to the origin open and sends new requests over
// them; an upgrade can arrive on an idle one. Closing the listener must close
// those too, or they keep reaching the closed instance.
func TestCloseDropsKeepAliveConnections(t *testing.T) {
	port := tcp.PickPort()
	ln, err := ListenWS(context.Background(), net.LocalHostIP, port, &internet.MemoryStreamConfig{
		ProtocolName:     "websocket",
		ProtocolSettings: &Config{Path: "ws"},
	}, func(conn stat.Connection) { conn.Close() })
	if err != nil {
		t.Fatal(err)
	}

	c, err := gonet.Dial("tcp", gonet.JoinHostPort("127.0.0.1", port.String()))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	r := bufio.NewReader(c)
	if _, err := c.Write([]byte("GET /other HTTP/1.1\r\nHost: example.com\r\n\r\n")); err != nil {
		t.Fatal(err)
	}
	resp, err := http.ReadResponse(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.Close {
		t.Fatal("server did not keep the connection alive; the test needs it to")
	}

	ln.Close()

	c.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, err := r.ReadByte(); err == nil {
		t.Fatal("got data on the idle connection after Close")
	} else if ne, ok := err.(gonet.Error); ok && ne.Timeout() {
		t.Fatal("idle keep-alive connection still open after Close")
	}
}
