package go_stomp_websocket

import (
	"net/url"
	"testing"
	"time"

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
