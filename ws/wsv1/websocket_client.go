package wsv1

import (
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
func (s *WebsocketClientV1) PublicV1() (*PublicV1Service, error) {
	url := s.Client.BaseURL + PublicV1Path
	c, _, err := s.Client.Dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &PublicV1Service{
		connection:    c,
		paramTradeMap: map[PublicV1TradeParamKey]func(PublicV1TradeResponse) error{},
	}, nil
}

// PublicV2 :
func (s *WebsocketClientV1) PublicV2() (*PublicV2Service, error) {
	url := s.Client.BaseURL + PublicV2Path
	c, _, err := s.Client.Dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &PublicV2Service{
		connection:    c,
		paramTradeMap: map[PublicV2TradeParamKey]func(PublicV2TradeResponse) error{},
	}, nil
}

// Private :
func (s *WebsocketClientV1) Private() (*PrivateService, error) {
	url := s.Client.BaseURL + PrivatePath
	c, _, err := s.Client.Dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &PrivateService{
		client:                      s.Client,
		connection:                  c,
		paramOutboundAccountInfoMap: map[PrivateParamKey]func(PrivateOutboundAccountInfoResponse) error{},
	}, nil
}
