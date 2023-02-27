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
	Public(bybit.CategoryV5) (PublicService, error)
	Private() (PrivateService, error)
}

// Public :
func (s *WebsocketClientV5) Public(category bybit.CategoryV5) (PublicServiceI, error) {
	url := s.Client.BaseURL + PublicPathFor(category)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &PublicService{
		client:            s.Client,
		connection:        c,
		paramOrderBookMap: map[PublicOrderBookParamKey]func(PublicOrderBookResponse) error{},
	}, nil
}

// Private :
func (s *WebsocketClientV5) Private() (PrivateServiceI, error) {
	url := s.Client.BaseURL + PrivatePath
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &PrivateService{
		client:           s.Client,
		connection:       c,
		paramOrderMap:    map[PrivateParamKey]func(PrivateOrderResponse) error{},
		paramPositionMap: map[PrivateParamKey]func(PrivatePositionResponse) error{},
	}, nil
}
