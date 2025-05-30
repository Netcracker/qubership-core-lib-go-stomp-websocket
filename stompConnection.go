package go_stomp_websocket

import (
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
)

var logger = logging.GetLogger("stomp")

type WebSocketConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type StompClient struct {
	webSocketURL url.URL
	connection   WebSocketConn
	readCh       chan *Frame
	writeCh      chan writeRequest
}

type writeRequest struct {
	Frame *Frame      // frame to send
	C     chan *Frame // response channel
}

type ConnectionDialer interface {
	Dial(webSocketURL url.URL, dialer websocket.Dialer, requestHeaders http.Header) (WebSocketConn, *http.Response, error)
}

func Connect(webSocketURL url.URL, dialer websocket.Dialer, requestHeaders http.Header, connDialer ConnectionDialer) (*StompClient, error) {
	webSocketURL.Path = webSocketURL.Path + "/" + randomIntn(999) + "/" + randomString() + "/websocket"
	logger.Infof("connecting to %s", webSocketURL.String())
	conn, _, err := connDialer.Dial(webSocketURL, dialer, requestHeaders)
	if err != nil {
		return nil, err
	}
	return establishConnection(webSocketURL, conn)
}

func ConnectWithToken(webSocketURL url.URL, dialer websocket.Dialer, token string) (*StompClient, error) {
	webSocketURL.Path = webSocketURL.Path + "/" + randomIntn(999) + "/" + randomString() + "/websocket"
	logger.Infof("connecting to %s", webSocketURL.String())
	schema, err := extractSchema(webSocketURL)
	if err != nil {
		logger.Errorf("Schema have to start with ws or wss \n %v", err)
		return nil, err
	}
	requestHeaders := http.Header{}
	requestHeaders.Add("Host", webSocketURL.Host)
	requestHeaders.Add("Origin", schema+"://"+webSocketURL.Host)
	requestHeaders.Add("Authorization", "Bearer "+token)
	conn, _, err := dialer.Dial(webSocketURL.String(), requestHeaders)
	if err != nil {
		return nil, err
	}
	return establishConnection(webSocketURL, conn)
}

func establishConnection(webSocketURL url.URL, conn WebSocketConn) (*StompClient, error) {
	readCh := make(chan *Frame)
	writeCh := make(chan writeRequest)
	stompClient := &StompClient{
		webSocketURL: webSocketURL,
		connection:   conn,
		readCh:       readCh,
		writeCh:      writeCh,
	}

	headers := []string{"accept-version:1.2,1.1,1.0", "heart-beat:10000,10000"}
	connectFrame := CreateFrame(CONNECT, headers)
	if connectErr := stompClient.connection.WriteMessage(1, connectFrame.Bytes()); connectErr != nil {
		return nil, connectErr
	} else {
		_, _, err := stompClient.connection.ReadMessage()
		if err != nil {
			return nil, err
		}
	}
	go readLoop(stompClient)
	go processLoop(stompClient)
	return stompClient, nil
}

func (stompClient StompClient) Disconnect() error {
	receiptId := uuid.New()
	headers := []string{"receipt:" + receiptId.String()}

	ch := make(chan *Frame)
	stompClient.writeCh <- writeRequest{
		Frame: CreateFrame(DISCONNECT, headers),
		C:     ch,
	}
	response := <-ch
	if response.Command == RECEIPT {
		logger.Infof("Connection closed")
		stompClient.connection.Close()
	}
	return nil
}

func readLoop(stompClient *StompClient) {
	for {
		_, data, err := stompClient.connection.ReadMessage()
		if err != nil {
			logger.Errorf("An error occurred while reading message: %s\n", err)
			stompClient.readCh <- &Frame{Command: ERROR}
			break
		}
		if len(data) < 1 {
			continue
		}
		switch data[0] {
		case 'h':
			// Heartbeat
			continue
		case 'a':
			// Normal message
			stompClient.readCh <- ReadFrame(data)
		case 'c':
			// Session closed
			break
		}
	}
}

func processLoop(stompClient *StompClient) {
	channels := make(map[string]chan *Frame)
	for {
		select {

		case f, _ := <-stompClient.readCh:
			switch f.Command {
			case RECEIPT:
				if id, ok := f.Contains(ReceiptId); ok {
					if ch, ok := channels[id]; ok {
						ch <- f
						delete(channels, id)
						close(ch)
					}
				} else {
					err := "missing receipt-id"
					sendError(channels, err)
					return
				}

			case ERROR:
				logger.Errorf("received ERROR; Closing underlying connection")
				for _, ch := range channels {
					ch <- f
					close(ch)
				}
				stompClient.connection.Close()

				return

			case MESSAGE:
				if id, ok := f.Contains(Subscription_h); ok {
					if ch, ok := channels[id]; ok {
						ch <- f
					} else {
						logger.Infof("ignored MESSAGE for subscription", id)
					}
				}
			}

		case req, _ := <-stompClient.writeCh:
			if req.C != nil {
				if receipt, ok := req.Frame.Contains(Receipt); ok {
					// remember the channel for this receipt
					channels[receipt] = req.C
				}
			}
			switch req.Frame.Command {
			case SUBSCRIBE:
				id, _ := req.Frame.Contains(Id)
				channels[id] = req.C
			}
			err := stompClient.connection.WriteMessage(1, req.Frame.Bytes())
			if err != nil {
				logger.Infof("Can't send message", err)
			}
		}
	}
}

func sendError(m map[string]chan *Frame, err string) {
	headers := []string{Message + ":" + err}
	frame := CreateFrame(ERROR, headers)
	for _, ch := range m {
		ch <- frame
	}
}

func randomIntn(max int) string {
	var (
		ml = len(strconv.Itoa(max))
		ri = rand.Intn(max)
		is = strconv.Itoa(ri)
	)
	if len(is) < ml {
		is = strings.Repeat("0", ml-len(is)) + is
	}
	return is
}

func randomString() string {
	length := 16
	chars := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	clen := len(chars)
	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			panic("uniuri: error reading random bytes: " + err.Error())
		}
		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				// Skip this number to avoid modulo bias.
				continue
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

func extractSchema(webSocketURL url.URL) (string, error) {
	switch webSocketURL.Scheme {
	case "ws":
		return "http", nil
	case "wss":
		return "https", nil
	}
	return "", errors.New("malformed ws or wss URL")
}
