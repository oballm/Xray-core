package inbound

import (
	"net"
	"testing"
	"time"
)

func TestConnTrackerClosesTrackedConnections(t *testing.T) {
	var tr connTracker
	a, aPeer := net.Pipe()
	b, bPeer := net.Pipe()
	defer aPeer.Close()
	defer bPeer.Close()

	if _, ok := tr.add(a); !ok {
		t.Fatal("add refused before close")
	}
	idB, _ := tr.add(b)
	tr.remove(idB) // finished on its own; must not be touched

	tr.closeAll()

	aPeer.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := aPeer.Read(make([]byte, 1)); err == nil || isTimeout(err) {
		t.Fatalf("tracked connection not closed: %v", err)
	}
	bPeer.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if _, err := bPeer.Read(make([]byte, 1)); !isTimeout(err) {
		t.Fatalf("removed connection was closed: %v", err)
	}

	c, _ := net.Pipe()
	if _, ok := tr.add(c); ok {
		t.Fatal("add accepted after closeAll")
	}
}

func isTimeout(err error) bool {
	ne, ok := err.(net.Error)
	return ok && ne.Timeout()
}
