package wsv1

import (
	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit/ws"
)

// WebsocketClientV1 :
type WebsocketClientV1 struct {
	Client *ws.WebSocketClient
}

// NewWSClient v1 client
func NewWSClient(c *ws.WebSocketClient) *WebsocketClientV1 {
	return &WebsocketClientV1{Client: c}
}

// PublicV1 :
func (s *WebsocketClientV1) PublicV1() (*SpotWebsocketV1PublicV1Service, error) {
	url := s.Client.BaseURL + SpotWebsocketV1PublicV1Path
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &SpotWebsocketV1PublicV1Service{
		connection:    c,
		paramTradeMap: map[SpotWebsocketV1PublicV1TradeParamKey]func(SpotWebsocketV1PublicV1TradeResponse) error{},
	}, nil
}

// PublicV2 :
func (s *WebsocketClientV1) PublicV2() (*SpotWebsocketV1PublicV2Service, error) {
	url := s.Client.BaseURL + SpotWebsocketV1PublicV2Path
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &SpotWebsocketV1PublicV2Service{
		connection:    c,
		paramTradeMap: map[SpotWebsocketV1PublicV2TradeParamKey]func(SpotWebsocketV1PublicV2TradeResponse) error{},
	}, nil
}

// Private :
func (s *WebsocketClientV1) Private() (*SpotWebsocketV1PrivateService, error) {
	url := s.Client.BaseURL + SpotWebsocketV1PrivatePath
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &SpotWebsocketV1PrivateService{
		client:                      s.Client,
		connection:                  c,
		paramOutboundAccountInfoMap: map[SpotWebsocketV1PrivateParamKey]func(SpotWebsocketV1PrivateOutboundAccountInfoResponse) error{},
	}, nil
}
