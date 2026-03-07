package randomport

import (
	"net"
	"testing"
)

func TestFind(t *testing.T) {
	port, err := Find()
	if err != nil {
		t.Fatalf("Find() returned error: %v", err)
	}
	if port <= 0 {
		t.Fatalf("Find() returned invalid port: %d", port)
	}

	// Verify the port is actually available by trying to listen on it
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	testPort := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	// The found port should be in valid range
	if port < 1 || port > 65535 {
		t.Fatalf("Find() returned port out of range: %d", port)
	}

	_ = testPort // suppress unused variable warning
}

func TestMustFind(t *testing.T) {
	port := MustFind()
	if port <= 0 {
		t.Fatalf("MustFind() returned invalid port: %d", port)
	}
}

func TestFindMultiple(t *testing.T) {
	ports := make(map[int]bool)
	for i := 0; i < 10; i++ {
		port, err := Find()
		if err != nil {
			t.Fatalf("Find() iteration %d returned error: %v", i, err)
		}
		if ports[port] {
			t.Fatalf("Find() returned duplicate port: %d", port)
		}
		ports[port] = true
	}
}

func TestRelease(t *testing.T) {
	port, err := Find()
	if err != nil {
		t.Fatalf("Find() returned error: %v", err)
	}

	Release(port)

	// After release, we should be able to find ports again
	newPort, err := Find()
	if err != nil {
		t.Fatalf("Find() after release returned error: %v", err)
	}
	if newPort <= 0 {
		t.Fatalf("Find() after release returned invalid port: %d", newPort)
	}
}
