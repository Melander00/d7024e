package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

const testTimeout = 5 * time.Second

// getFreeAddress reserves an ephemeral TCP port and releases it so the
// TCPDataNetwork listener can use the same address.
func getFreeAddress(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a TCP address: %v", err)
	}

	address := listener.Addr().String()

	if err := listener.Close(); err != nil {
		t.Fatalf("failed to release reserved TCP address: %v", err)
	}

	return address
}

// waitForServer repeatedly attempts to establish a TCP connection until
// the server is ready or the test timeout expires.
func waitForServer(t *testing.T, address string) {
	t.Helper()

	deadline := time.Now().Add(testTimeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("TCP server at %q did not become ready", address)
}

// TestNewTCPDataNetwork verifies that a new network can be created.
func TestNewTCPDataNetwork(t *testing.T) {
	network := NewTCPDataNetwork()

	if network == nil {
		t.Fatal("NewTCPDataNetwork() returned nil")
	}

	if network.handler != nil {
		t.Fatal("new TCPDataNetwork should have a nil handler")
	}
}

// TestWriteFrame verifies that writeFrame sends a four-byte big-endian
// payload length followed by the payload.
func TestWriteFrame(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "empty payload",
			payload: []byte{},
		},
		{
			name:    "single byte",
			payload: []byte{0x42},
		},
		{
			name:    "text payload",
			payload: []byte("hello, world"),
		},
		{
			name:    "binary payload",
			payload: []byte{0x00, 0xFF, 0x01, 0x80, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, writer := net.Pipe()
			defer reader.Close()
			defer writer.Close()

			type result struct {
				err error
			}

			writeResult := make(chan result, 1)

			go func() {
				writeResult <- result{err: writeFrame(writer, tt.payload)}
			}()

			expectedFrame := make([]byte, 4+len(tt.payload))
			binary.BigEndian.PutUint32(
				expectedFrame[:4],
				uint32(len(tt.payload)),
			)
			copy(expectedFrame[4:], tt.payload)

			actualFrame := make([]byte, len(expectedFrame))
			if _, err := io.ReadFull(reader, actualFrame); err != nil {
				t.Fatalf("failed to read frame: %v", err)
			}

			if !bytes.Equal(actualFrame, expectedFrame) {
				t.Fatalf(
					"unexpected frame:\n got: %v\nwant: %v",
					actualFrame,
					expectedFrame,
				)
			}

			select {
			case result := <-writeResult:
				if result.err != nil {
					t.Fatalf("writeFrame() error = %v", result.err)
				}
			case <-time.After(testTimeout):
				t.Fatal("writeFrame() timed out")
			}
		})
	}
}

// TestReadFrame verifies that readFrame correctly decodes framed data.
func TestReadFrame(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "empty payload",
			payload: []byte{},
		},
		{
			name:    "single byte",
			payload: []byte{0x42},
		},
		{
			name:    "text payload",
			payload: []byte("hello, world"),
		},
		{
			name:    "binary payload",
			payload: []byte{0x00, 0xFF, 0x01, 0x80, 0x00},
		},
		{
			name:    "large payload",
			payload: bytes.Repeat([]byte("network"), 16*1024),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, writer := net.Pipe()
			defer reader.Close()
			defer writer.Close()

			writeResult := make(chan error, 1)

			go func() {
				writeResult <- writeFrame(writer, tt.payload)
			}()

			actual, err := readFrame(reader)
			if err != nil {
				t.Fatalf("readFrame() error = %v", err)
			}

			if !bytes.Equal(actual, tt.payload) {
				t.Fatalf(
					"readFrame() = %v, want %v",
					actual,
					tt.payload,
				)
			}

			select {
			case err := <-writeResult:
				if err != nil {
					t.Fatalf("writeFrame() error = %v", err)
				}
			case <-time.After(testTimeout):
				t.Fatal("writeFrame() timed out")
			}
		})
	}
}

// TestFrameRoundTrip verifies that data survives encoding and decoding.
func TestFrameRoundTrip(t *testing.T) {
	payload := bytes.Repeat([]byte{0x00, 0x7F, 0x80, 0xFF}, 4096)

	reader, writer := net.Pipe()
	defer reader.Close()
	defer writer.Close()

	writeResult := make(chan error, 1)

	go func() {
		writeResult <- writeFrame(writer, payload)
	}()

	actual, err := readFrame(reader)
	if err != nil {
		t.Fatalf("readFrame() error = %v", err)
	}

	if !bytes.Equal(actual, payload) {
		t.Fatal("frame round trip produced different data")
	}

	if err := <-writeResult; err != nil {
		t.Fatalf("writeFrame() error = %v", err)
	}
}

// TestReadFrameTruncatedLength verifies that an incomplete length header
// returns an error.
func TestReadFrameTruncatedLength(t *testing.T) {
	reader, writer := net.Pipe()
	defer reader.Close()

	// Write and close in a separate goroutine so the reader can
	// consume the partial header and observe EOF.
	go func() {
		_, _ = writer.Write([]byte{0x00, 0x00})
		_ = writer.Close()
	}()

	payload, err := readFrame(reader)

	if err == nil {
		t.Fatal("readFrame() expected an error for a truncated header")
	}

	if payload != nil {
		t.Fatalf("readFrame() payload = %v, want nil", payload)
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("readFrame() error = %v, want io.ErrUnexpectedEOF", err)
	}
}

// TestReadFrameTruncatedPayload verifies that an incomplete payload
// returns an error.
func TestReadFrameTruncatedPayload(t *testing.T) {
	reader, writer := net.Pipe()

	go func() {
		header := make([]byte, 4)
		binary.BigEndian.PutUint32(header, 5)

		_, _ = writer.Write(header)
		_, _ = writer.Write([]byte("ab"))
		_ = writer.Close()
	}()

	defer reader.Close()

	payload, err := readFrame(reader)

	if err == nil {
		t.Fatal("readFrame() expected an error for a truncated payload")
	}

	if payload != nil {
		t.Fatalf("readFrame() payload = %v, want nil", payload)
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("readFrame() error = %v, want io.ErrUnexpectedEOF", err)
	}
}

// TestReadFrameClosedConnection verifies that readFrame returns an error
// when the connection is closed before a header is received.
func TestReadFrameClosedConnection(t *testing.T) {
	reader, writer := net.Pipe()
	defer reader.Close()
	defer writer.Close()

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	payload, err := readFrame(reader)

	if err == nil {
		t.Fatal("readFrame() expected an error on a closed connection")
	}

	if payload != nil {
		t.Fatalf("readFrame() payload = %v, want nil", payload)
	}

	if !errors.Is(err, io.EOF) {
		t.Fatalf("readFrame() error = %v, want io.EOF", err)
	}
}

// TestWriteFrameClosedConnection verifies that writeFrame returns an error
// when the connection is already closed.
func TestWriteFrameClosedConnection(t *testing.T) {
	reader, writer := net.Pipe()
	defer reader.Close()
	defer writer.Close()

	if err := reader.Close(); err != nil {
		t.Fatalf("failed to close reader: %v", err)
	}

	err := writeFrame(writer, []byte("test"))

	if err == nil {
		t.Fatal("writeFrame() expected an error on a closed connection")
	}
}

// TestRequestConnectionError verifies that Request returns an error
// when it cannot connect to the specified address.
func TestRequestConnectionError(t *testing.T) {
	network := NewTCPDataNetwork()

	_, err := network.Request("invalid-address", []byte("test"))

	if err == nil {
		t.Fatal("Request() expected an error for an invalid address")
	}
}

// TestTCPDataNetworkListenAndRequest verifies a complete TCP request-response
// cycle using the public API.
func TestTCPDataNetworkListenAndRequest(t *testing.T) {
	address := getFreeAddress(t)
	network := NewTCPDataNetwork()

	handler := func(request []byte) []byte {
		return append([]byte("response: "), request...)
	}

	if err := network.Listen(address, handler); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	waitForServer(t, address)

	tests := []struct {
		name    string
		request []byte
		want    []byte
	}{
		{
			name:    "text request",
			request: []byte("hello"),
			want:    []byte("response: hello"),
		},
		{
			name:    "empty request",
			request: []byte{},
			want:    []byte("response: "),
		},
		{
			name:    "binary request",
			request: []byte{0x00, 0xFF, 0x80, 0x01},
			want:    []byte("response: \x00\xff\x80\x01"),
		},
		{
			name:    "large request",
			request: bytes.Repeat([]byte("payload"), 16*1024),
			want:    append([]byte("response: "), bytes.Repeat([]byte("payload"), 16*1024)...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := network.Request(address, tt.request)
			if err != nil {
				t.Fatalf("Request() error = %v", err)
			}

			if !bytes.Equal(response, tt.want) {
				t.Fatalf(
					"Request() = %v, want %v",
					response,
					tt.want,
				)
			}
		})
	}
}

// TestTCPDataNetworkRequestEcho verifies that the server can echo arbitrary
// payloads without changing their contents.
func TestTCPDataNetworkRequestEcho(t *testing.T) {
	address := getFreeAddress(t)
	network := NewTCPDataNetwork()

	handler := func(request []byte) []byte {
		return append([]byte(nil), request...)
	}

	if err := network.Listen(address, handler); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	waitForServer(t, address)

	payload := bytes.Repeat([]byte{0x00, 0xFF, 0x01, 0x80}, 8192)

	response, err := network.Request(address, payload)
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}

	if !bytes.Equal(response, payload) {
		t.Fatal("echo response differs from original payload")
	}
}

// TestTCPDataNetworkConcurrentRequests verifies that multiple clients can
// send requests concurrently.
func TestTCPDataNetworkConcurrentRequests(t *testing.T) {
	address := getFreeAddress(t)
	network := NewTCPDataNetwork()

	handler := func(request []byte) []byte {
		return append([]byte("echo:"), request...)
	}

	if err := network.Listen(address, handler); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	waitForServer(t, address)

	const requestCount = 50

	var wg sync.WaitGroup
	errorsCh := make(chan error, requestCount)

	for i := 0; i < requestCount; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			request := []byte{
				byte(i),
				byte(i >> 8),
				0xFF,
			}

			response, err := network.Request(address, request)
			if err != nil {
				errorsCh <- err
				return
			}

			want := append([]byte("echo:"), request...)

			if !bytes.Equal(response, want) {
				errorsCh <- errors.New("unexpected response for concurrent request")
			}
		}(i)
	}

	wg.Wait()
	close(errorsCh)

	for err := range errorsCh {
		t.Errorf("concurrent request failed: %v", err)
	}
}

// TestTCPDataNetworkHandlerReceivesExactRequest verifies that the handler
// receives the exact request payload.
func TestTCPDataNetworkHandlerReceivesExactRequest(t *testing.T) {
	address := getFreeAddress(t)
	network := NewTCPDataNetwork()

	expectedRequest := []byte{0x00, 0x01, 0xFE, 0xFF}

	requestReceived := make(chan []byte, 1)

	handler := func(request []byte) []byte {
		requestReceived <- append([]byte(nil), request...)
		return []byte("ok")
	}

	if err := network.Listen(address, handler); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	waitForServer(t, address)

	response, err := network.Request(address, expectedRequest)
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}

	if !bytes.Equal(response, []byte("ok")) {
		t.Fatalf("response = %v, want %v", response, []byte("ok"))
	}

	select {
	case actual := <-requestReceived:
		if !reflect.DeepEqual(actual, expectedRequest) {
			t.Fatalf("handler received %v, want %v", actual, expectedRequest)
		}
	case <-time.After(testTimeout):
		t.Fatal("handler did not receive the request")
	}
}

// TestTCPDataNetworkMultipleSequentialRequests verifies that each connection
// can independently complete a request-response exchange.
func TestTCPDataNetworkMultipleSequentialRequests(t *testing.T) {
	address := getFreeAddress(t)
	network := NewTCPDataNetwork()

	handler := func(request []byte) []byte {
		return bytes.ToUpper(request)
	}

	if err := network.Listen(address, handler); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	waitForServer(t, address)

	requests := []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
	}

	for _, request := range requests {
		t.Run(request, func(t *testing.T) {
			response, err := network.Request(address, []byte(request))
			if err != nil {
				t.Fatalf("Request() error = %v", err)
			}

			want := []byte(bytes.ToUpper([]byte(request)))

			if !bytes.Equal(response, want) {
				t.Fatalf("response = %q, want %q", response, want)
			}
		})
	}
}
