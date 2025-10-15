package go_stomp_websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeAndUnsubscribe(t *testing.T) {
	client := &StompClient{
		writeCh: make(chan writeRequest, 2), // buffered!
	}

	sub, err := client.Subscribe("/topic/test")
	assert.NoError(t, err)
	assert.Equal(t, "/topic/test", sub.Topic)

	// read first message
	req1 := <-client.writeCh
	assert.Equal(t, SUBSCRIBE, req1.Frame.Command)

	// unsubscribe
	sub.Unsubscribe()

	// read second message
	req2 := <-client.writeCh
	assert.Equal(t, UNSUBSCRIBE, req2.Frame.Command)
	assert.Contains(t, req2.Frame.Headers[0], "id:"+sub.Id)
}
