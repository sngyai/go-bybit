package v5

import (
	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit/ws"
)

// WebsocketClientV5 :
type WebsocketClientV5 struct {
	Client *ws.WebSocketClient
}

// NewWSClientV1 :
func NewWSClientV1(c *ws.WebSocketClient) *WebsocketClientV5 {
	return &WebsocketClientV5{Client: c}
}

// Spot :
func (s *WebsocketClientV5) Spot() V5WebsocketSpotServiceI {
	return &V5WebsocketSpotService{s.Client}
}

// Linear :
func (s *WebsocketClientV5) Linear() V5WebsocketLinearServiceI {
	return &V5WebsocketLinearService{s.Client}
}

// Inverse :
func (s *WebsocketClientV5) Inverse() V5WebsocketInverseServiceI {
	return &V5WebsocketInverseService{s.Client}
}

// Option :
func (s *WebsocketClientV5) Option() V5WebsocketOptionServiceI {
	return &V5WebsocketOptionService{s.Client}
}

// Private :
func (s *WebsocketClientV5) Private() (*V5WebsocketPrivateService, error) {
	url := s.Client.BaseURL + V5WebsocketPrivatePath
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &V5WebsocketPrivateService{
		client:          s.Client,
		connection:      c,
		paramPrivateMap: map[V5WebsocketPrivateParamKey]func(V5WebsocketPrivatePositionResponseContent) error{},
	}, nil
}
