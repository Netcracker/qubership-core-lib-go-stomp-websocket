package go_stomp_websocket

import "github.com/google/uuid"

type Subscription struct {
	FrameCh     chan *Frame
	Id          string
	Topic       string
	stompClient StompClient
}

func (stompClient StompClient) Subscribe(topic string) (*Subscription, error) {
	subscription := &Subscription{}
	subscriptionId := uuid.New()
	headers := []string{"id:" + subscriptionId.String(), "destination:" + topic}
	ch := make(chan *Frame)
	stompClient.writeCh <- writeRequest{
		Frame: CreateFrame(SUBSCRIBE, headers),
		C:     ch,
	}
	subscription = &Subscription{
		stompClient: stompClient,
		Id:          subscriptionId.String(),
		FrameCh:     ch,
		Topic:       topic,
	}
	return subscription, nil
}

func (s *Subscription) Unsubscribe() {
	headers := []string{"id:" + s.Id}
	ch := make(chan *Frame)
	s.stompClient.writeCh <- writeRequest{
		Frame: CreateFrame(UNSUBSCRIBE, headers),
		C:     ch,
	}
}
