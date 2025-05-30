package go_stomp_websocket

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	defaultTimeout = time.Second
	shortTimeout   = 100 * time.Millisecond
)

func TestSubscribe(t *testing.T) {
	tests := []struct {
		name          string
		topic         string
		expectedError bool
	}{
		{
			name:          "successful subscription",
			topic:         "/topic/test",
			expectedError: false,
		},
		{
			name:          "empty topic",
			topic:         "",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			client := setupTestClient(t)
			done := make(chan struct{})
			defer close(done)

			// Start process loop
			go runProcessLoop(client, tt.topic, done)

			// Test subscription
			subscription, err := client.Subscribe(tt.topic)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, subscription)
				return
			}

			// Verify subscription
			assert.NoError(t, err)
			assert.NotNil(t, subscription)
			assert.NotEmpty(t, subscription.Id)
			assert.Equal(t, tt.topic, subscription.Topic)
			assert.NotNil(t, subscription.FrameCh)
			assert.Equal(t, client, subscription.stompClient)

			// Verify UUID format
			_, err = uuid.Parse(subscription.Id)
			assert.NoError(t, err)

			// Verify frames in sequence
			verifyFrames(t, subscription)

			// Test unsubscribe
			subscription.Unsubscribe()

			// Verify no more messages
			select {
			case <-subscription.FrameCh:
				t.Error("Unexpected message after unsubscribe")
			case <-time.After(shortTimeout):
				// Expected - no messages should arrive
			}
		})
	}
}

// setupTestClient creates a StompClient with a mock WebSocketConn
func setupTestClient(t *testing.T) StompClient {
	mockConn := new(MockWebSocketConn)
	mockConn.On("WriteMessage", 1, mock.Anything).Return(nil)
	mockConn.On("ReadMessage").Return(1, []byte("CONNECTED\nversion:1.2\n\n\x00"), nil)

	return StompClient{
		webSocketURL: url.URL{},
		connection:   mockConn,
		readCh:       make(chan *Frame),
		writeCh:      make(chan writeRequest),
	}
}

// runProcessLoop handles the client's write requests and sends appropriate responses
func runProcessLoop(client StompClient, topic string, done chan struct{}) {
	for {
		select {
		case req := <-client.writeCh:
			if req.C != nil {
				switch req.Frame.Command {
				case SUBSCRIBE:
					// Send RECEIPT and then test message
					req.C <- &Frame{Command: RECEIPT}
					go func() {
						time.Sleep(shortTimeout)
						req.C <- &Frame{
							Command: MESSAGE,
							Headers: []string{"destination:" + topic},
							Body:    "test message",
						}
					}()
				case UNSUBSCRIBE:
					req.C <- &Frame{Command: RECEIPT}
				}
			}
		case <-done:
			return
		}
	}
}

// verifyFrames checks that the subscription receives frames in the expected order
func verifyFrames(t *testing.T, subscription *Subscription) {
	// First frame should be RECEIPT
	select {
	case frame := <-subscription.FrameCh:
		assert.Equal(t, RECEIPT, frame.Command)
	case <-time.After(defaultTimeout):
		t.Error("Timeout waiting for RECEIPT")
	}

	// Second frame should be MESSAGE
	select {
	case frame := <-subscription.FrameCh:
		assert.Equal(t, MESSAGE, frame.Command)
		assert.Equal(t, "test message", frame.Body)
	case <-time.After(defaultTimeout):
		t.Error("Timeout waiting for MESSAGE")
	}
}
