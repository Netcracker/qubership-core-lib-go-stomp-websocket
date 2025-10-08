package go_stomp_websocket

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestRandomIntn(t *testing.T) {
	tests := []struct {
		name        string
		max         int
		expectedLen int
	}{
		{
			name:        "single digit max",
			max:         9,
			expectedLen: 1,
		},
		{
			name:        "double digit max",
			max:         99,
			expectedLen: 2,
		},
		{
			name:        "triple digit max",
			max:         999,
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := randomIntn(tt.max)
			assert.Len(t, result, tt.expectedLen)
			assert.Regexp(t, "^[0-9]+$", result)
		})
	}
}

func TestRandomString(t *testing.T) {
	tests := []struct {
		name        string
		expectedLen int
	}{
		{
			name:        "default length",
			expectedLen: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := randomString()
			assert.Len(t, result, tt.expectedLen)
			assert.Regexp(t, "^[A-Za-z0-9]+$", result)
		})
	}
}

func TestExtractSchema(t *testing.T) {
	tests := []struct {
		name          string
		webSocketURL  url.URL
		expected      string
		expectedError bool
	}{
		{
			name: "ws schema",
			webSocketURL: url.URL{
				Scheme: "ws",
			},
			expected:      "http",
			expectedError: false,
		},
		{
			name: "wss schema",
			webSocketURL: url.URL{
				Scheme: "wss",
			},
			expected:      "https",
			expectedError: false,
		},
		{
			name: "invalid schema",
			webSocketURL: url.URL{
				Scheme: "http",
			},
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractSchema(tt.webSocketURL)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSendError(t *testing.T) {
	// Create test channels
	ch1 := make(chan *Frame)
	ch2 := make(chan *Frame)
	ch3 := make(chan *Frame)

	// Create channel map
	channels := map[string]chan *Frame{
		"sub1": ch1,
		"sub2": ch2,
		"sub3": ch3,
	}

	// Test error message
	errorMsg := "test error message"

	// Start goroutine to send error
	go sendError(channels, errorMsg)

	// Create a timeout channel
	timeout := time.After(1 * time.Second)

	// Check all channels receive the error frame
	for i := 0; i < 3; i++ {
		select {
		case frame := <-ch1:
			checkErrorFrame(t, frame, errorMsg)
		case frame := <-ch2:
			checkErrorFrame(t, frame, errorMsg)
		case frame := <-ch3:
			checkErrorFrame(t, frame, errorMsg)
		case <-timeout:
			t.Fatal("Timeout waiting for error frames")
		}
	}
}

func TestSendErrorEmptyMap(t *testing.T) {
	// Test with empty channel map
	channels := map[string]chan *Frame{}
	errorMsg := "test error message"

	// This should not panic
	sendError(channels, errorMsg)
}

// checkErrorFrame verifies that a frame contains the expected error message
func checkErrorFrame(t *testing.T, frame *Frame, expectedMsg string) {
	t.Helper()
	if frame.Command != ERROR {
		t.Errorf("Expected ERROR command, got %s", frame.Command)
	}
	if msg, ok := frame.Contains(Message); !ok || msg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, msg)
	}
}

func TestRandomStringAndIntn(t *testing.T) {
	r1 := randomIntn(999)
	r2 := randomIntn(999)
	assert.Len(t, r1, 3)
	assert.NotEqual(t, r1, r2)

	rs := randomString()
	assert.Len(t, rs, 16)
}

func TestConnectWithToken_InvalidSchema(t *testing.T) {
	u, _ := url.Parse("ftp://localhost/test")
	dialer := websocket.Dialer{}
	client, err := ConnectWithToken(*u, dialer, "token123")
	assert.Nil(t, client)
	assert.Error(t, err)
}

// startTestWSServer starts a websocket test server that:
// - Reads the first client message (CONNECT frame)
// - Sends a dummy message to satisfy the initial ReadMessage in establishConnection
// - Echoes a RECEIPT with the receipt-id extracted from a DISCONNECT frame
func startTestWSServer(t *testing.T) (*httptest.Server, chan struct{}) {
	t.Helper()
	done := make(chan struct{})
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade error: %v", err)
			return
		}
		defer c.Close()

		// Read initial CONNECT frame from client
		_, _, err = c.ReadMessage()
		if err != nil {
			t.Errorf("failed reading initial client message: %v", err)
			return
		}
		// Send a dummy message to let client proceed
		_ = c.WriteMessage(websocket.TextMessage, []byte("o"))

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				close(done)
				return
			}
			// Expect a DISCONNECT frame array payload like: ["DISCONNECT\nreceipt:<id>\n\n\u0000"]
			// We'll do a very light parse to find "receipt:" and extract value until \n
			m := string(msg)
			idx := -1
			if i := findSubstring(m, "receipt:"); i >= 0 {
				idx = i + len("receipt:")
			}
			if idx >= 0 {
				end := idx
				for end < len(m) && m[end] != '\\' && m[end] != '\n' && m[end] != '"' {
					end++
				}
				receiptID := m[idx:end]
				resp := "a[\"RECEIPT\\nreceipt-id:" + receiptID + "\\n\\n\\u0000\"]"
				_ = c.WriteMessage(websocket.TextMessage, []byte(resp))
			}
		}
	})

	ts := httptest.NewServer(h)
	return ts, done
}

// findSubstring is a small helper to avoid strings.Index to keep control on escaping.
func findSubstring(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestConnectWithToken_SuccessAndDisconnect(t *testing.T) {
	// Start a websocket server
	ts, done := startTestWSServer(t)
	defer ts.Close()

	// Convert test server URL to ws scheme
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	u.Scheme = "ws"
	// Append base path expected by ConnectWithToken code path augmentation
	u.Path = u.Path + "/test"

	client, err := ConnectWithToken(*u, websocket.Dialer{}, "token-abc")
	if err != nil {
		t.Fatalf("ConnectWithToken failed: %v", err)
	}
	if client == nil {
		t.Fatal("client is nil")
	}

	// Trigger graceful disconnect which expects a RECEIPT
	derr := client.Disconnect()
	assert.NoError(t, derr)

	// Wait a short time for server to observe the close
	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("server did not finish in time")
	}
}
