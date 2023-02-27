package wsv5

import (
	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit"
	"github.com/sngyai/go-bybit/ws"
)

// WebsocketClientV5 :
type WebsocketClientV5 struct {
	Client *ws.WebSocketClient
}

// NewWSClient v5 client
func NewWSClient(c *ws.WebSocketClient) *WebsocketClientV5 {
	return &WebsocketClientV5{Client: c}
}

// V5WebsocketServiceI :
type V5WebsocketServiceI interface {
	Public(bybit.CategoryV5) (V5WebsocketPublicService, error)
	Private() (V5WebsocketPrivateService, error)
}

// Public :
func (s *WebsocketClientV5) Public(category bybit.CategoryV5) (V5WebsocketPublicServiceI, error) {
	url := s.Client.BaseURL + V5WebsocketPublicPathFor(category)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &V5WebsocketPublicService{
		client:            s.Client,
		connection:        c,
		paramOrderBookMap: map[V5WebsocketPublicOrderBookParamKey]func(V5WebsocketPublicOrderBookResponse) error{},
	}, nil
}

// Private :
func (s *WebsocketClientV5) Private() (V5WebsocketPrivateServiceI, error) {
	url := s.Client.BaseURL + V5WebsocketPrivatePath
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &V5WebsocketPrivateService{
		client:           s.Client,
		connection:       c,
		paramOrderMap:    map[V5WebsocketPrivateParamKey]func(V5WebsocketPrivateOrderResponse) error{},
		paramPositionMap: map[V5WebsocketPrivateParamKey]func(V5WebsocketPrivatePositionResponse) error{},
	}, nil
}
