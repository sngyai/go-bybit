package wsv1

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit"
	"github.com/sngyai/go-bybit/ws"
)

// PublicV1Service :
type PublicV1Service struct {
	connection *websocket.Conn

	paramTradeMap map[PublicV1TradeParamKey]func(PublicV1TradeResponse) error
}

const (
	// PublicV1Path :
	PublicV1Path = "/spot/quote/ws/v1"
)

// PublicV1Event :
type PublicV1Event string

const (
	// PublicV1EventSubscribe :
	PublicV1EventSubscribe = "sub"
	// PublicV1EventUnsubscribe :
	PublicV1EventUnsubscribe = "cancel"
)

// PublicV1Topic :
type PublicV1Topic string

const (
	// PublicV1TopicTrade :
	PublicV1TopicTrade = PublicV1Topic("trade")
)

// PublicV1TradeParamKey :
type PublicV1TradeParamKey struct {
	Symbol bybit.SymbolSpot
	Topic  PublicV1Topic
}

// PublicV1TradeResponse :
type PublicV1TradeResponse struct {
	Symbol         bybit.SymbolSpot `json:"symbol"`
	SymbolName     string           `json:"symbolName"`
	Topic          PublicV1Topic    `json:"topic"`
	SendTime       int              `json:"sendTime"`
	IsFirstMessage bool             `json:"f"`

	Params PublicV1TradeResponseParams `json:"params"`
	Data   []PublicV1TradeContent      `json:"data"`
}

// PublicV1TradeResponseParams :
type PublicV1TradeResponseParams struct {
	RealtimeInterval string `json:"realtimeInterval"`
	Binary           string `json:"binary"`
}

// PublicV1TradeContent :
type PublicV1TradeContent struct {
	TradeID        string `json:"v"`
	Timestamp      int    `json:"t"`
	Price          string `json:"p"`
	Quantity       string `json:"q"`
	IsBuySideTaker bool   `json:"m"`
}

// Key :
func (p *PublicV1TradeResponse) Key() PublicV1TradeParamKey {
	return PublicV1TradeParamKey{
		Symbol: p.Symbol,
		Topic:  p.Topic,
	}
}

// PublicV1TradeParamChild :
type PublicV1TradeParamChild struct {
	Binary bool `json:"binary"`
}

// PublicV1TradeParam :
type PublicV1TradeParam struct {
	Symbol bybit.SymbolSpot        `json:"symbol"`
	Topic  PublicV1Topic           `json:"topic"`
	Event  PublicV1Event           `json:"event"`
	Params PublicV1TradeParamChild `json:"params"`
}

// Key :
func (p *PublicV1TradeParam) Key() PublicV1TradeParamKey {
	return PublicV1TradeParamKey{
		Symbol: p.Symbol,
		Topic:  p.Topic,
	}
}

// addParamTradeFunc :
func (s *PublicV1Service) addParamTradeFunc(param PublicV1TradeParamKey, f func(PublicV1TradeResponse) error) error {
	if _, exist := s.paramTradeMap[param]; exist {
		return errors.New("already registered for this param")
	}
	s.paramTradeMap[param] = f
	return nil
}

// removeParamTradeFunc :
func (s *PublicV1Service) removeParamTradeFunc(key PublicV1TradeParamKey) {
	delete(s.paramTradeMap, key)
}

// retrieveTradeFunc :
func (s *PublicV1Service) retrieveTradeFunc(key PublicV1TradeParamKey) (func(PublicV1TradeResponse) error, error) {
	f, exist := s.paramTradeMap[key]
	if !exist {
		return nil, errors.New("func not found")
	}
	return f, nil
}

// judgeTopic :
func (s *PublicV1Service) judgeTopic(respBody []byte) (PublicV1Topic, error) {
	result := struct {
		Topic PublicV1Topic `json:"topic"`
	}{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	return result.Topic, nil
}

// parseResponse :
func (s *PublicV1Service) parseResponse(respBody []byte, response interface{}) error {
	if err := json.Unmarshal(respBody, &response); err != nil {
		return err
	}
	return nil
}

// SubscribeTrade :
func (s *PublicV1Service) SubscribeTrade(symbol bybit.SymbolSpot, f func(response PublicV1TradeResponse) error) (func() error, error) {
	param := PublicV1TradeParam{
		Symbol: symbol,
		Topic:  PublicV1TopicTrade,
		Event:  PublicV1EventSubscribe,
		Params: PublicV1TradeParamChild{
			Binary: false,
		},
	}
	if err := s.addParamTradeFunc(param.Key(), f); err != nil {
		return nil, err
	}
	buf, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	if err := s.connection.WriteMessage(websocket.TextMessage, []byte(buf)); err != nil {
		return nil, err
	}

	return func() error {
		param.Event = PublicV1EventUnsubscribe
		buf, err := json.Marshal(param)
		if err != nil {
			return err
		}
		if err := s.connection.WriteMessage(websocket.TextMessage, []byte(buf)); err != nil {
			return err
		}
		s.removeParamTradeFunc(param.Key())
		return nil
	}, nil
}

// Start :
func (s *PublicV1Service) Start(ctx context.Context) {
	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			if err := s.Run(); err != nil {
				if ws.IsErrWebsocketClosed(err) {
					return
				}
				log.Println(err)
				return
			}
		}
	}()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := s.Ping(); err != nil {
				return
			}
		case <-ctx.Done():
			log.Println("interrupt")

			if err := s.Close(); err != nil {
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

// Run :
func (s *PublicV1Service) Run() error {
	_, message, err := s.connection.ReadMessage()
	if err != nil {
		return err
	}

	topic, err := s.judgeTopic(message)
	if err != nil {
		return err
	}
	switch topic {
	case PublicV1TopicTrade:
		var resp PublicV1TradeResponse
		if err := s.parseResponse(message, &resp); err != nil {
			return err
		}
		f, err := s.retrieveTradeFunc(resp.Key())
		if err != nil {
			return err
		}
		if err := f(resp); err != nil {
			return err
		}
	}
	return nil
}

// Ping :
func (s *PublicV1Service) Ping() error {
	if err := s.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
		return err
	}
	return nil
}

// Close :
func (s *PublicV1Service) Close() error {
	if err := s.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		return err
	}
	return nil
}
