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

// PublicV2Service :
type PublicV2Service struct {
	connection *websocket.Conn

	paramTradeMap map[PublicV2TradeParamKey]func(PublicV2TradeResponse) error
}

const (
	// PublicV2Path :
	PublicV2Path = "/spot/quote/ws/v2"
)

// PublicV2Event :
type PublicV2Event string

const (
	// PublicV2EventSubscribe :
	PublicV2EventSubscribe = "sub"
	// PublicV2EventUnsubscribe :
	PublicV2EventUnsubscribe = "cancel"
)

// PublicV2Topic :
type PublicV2Topic string

const (
	// PublicV2TopicTrade :
	PublicV2TopicTrade = PublicV2Topic("trade")
)

// PublicV2TradeParamKey :
type PublicV2TradeParamKey struct {
	Symbol bybit.SymbolSpot
	Topic  PublicV2Topic
}

// PublicV2TradeResponse :
type PublicV2TradeResponse struct {
	Topic  PublicV2Topic               `json:"topic"`
	Params PublicV2TradeResponseParams `json:"params"`
	Data   PublicV2TradeContent        `json:"data"`
}

// PublicV2TradeResponseParams :
type PublicV2TradeResponseParams struct {
	Symbol     bybit.SymbolSpot `json:"symbol"`
	SymbolName string           `json:"symbolName"`
	Binary     string           `json:"binary"`
}

// PublicV2TradeContent :
type PublicV2TradeContent struct {
	TradeID        string `json:"v"`
	Timestamp      int    `json:"t"`
	Price          string `json:"p"`
	Quantity       string `json:"q"`
	IsBuySideTaker bool   `json:"m"`
}

// Key :
func (p *PublicV2TradeResponse) Key() PublicV2TradeParamKey {
	return PublicV2TradeParamKey{
		Symbol: p.Params.Symbol,
		Topic:  p.Topic,
	}
}

// PublicV2TradeParamChild :
type PublicV2TradeParamChild struct {
	Symbol bybit.SymbolSpot `json:"symbol"`
	Binary bool             `json:"binary"`
}

// PublicV2TradeParam :
type PublicV2TradeParam struct {
	Topic  PublicV2Topic           `json:"topic"`
	Event  PublicV2Event           `json:"event"`
	Params PublicV2TradeParamChild `json:"params"`
}

// Key :
func (p *PublicV2TradeParam) Key() PublicV2TradeParamKey {
	return PublicV2TradeParamKey{
		Symbol: p.Params.Symbol,
		Topic:  p.Topic,
	}
}

// addParamTradeFunc :
func (s *PublicV2Service) addParamTradeFunc(param PublicV2TradeParamKey, f func(PublicV2TradeResponse) error) error {
	if _, exist := s.paramTradeMap[param]; exist {
		return errors.New("already registered for this param")
	}
	s.paramTradeMap[param] = f
	return nil
}

// removeParamTradeFunc :
func (s *PublicV2Service) removeParamTradeFunc(key PublicV2TradeParamKey) {
	delete(s.paramTradeMap, key)
}

// retrieveTradeFunc :
func (s *PublicV2Service) retrieveTradeFunc(key PublicV2TradeParamKey) (func(PublicV2TradeResponse) error, error) {
	f, exist := s.paramTradeMap[key]
	if !exist {
		return nil, errors.New("func not found")
	}
	return f, nil
}

// judgeTopic :
func (s *PublicV2Service) judgeTopic(respBody []byte) (PublicV2Topic, error) {
	result := struct {
		Topic PublicV2Topic `json:"topic"`
		Event PublicV2Event `json:"event"`
	}{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	if result.Event == PublicV2EventSubscribe {
		return "", nil
	}
	return result.Topic, nil
}

// parseResponse :
func (s *PublicV2Service) parseResponse(respBody []byte, response interface{}) error {
	if err := json.Unmarshal(respBody, &response); err != nil {
		return err
	}
	return nil
}

// SubscribeTrade :
func (s *PublicV2Service) SubscribeTrade(symbol bybit.SymbolSpot, f func(response PublicV2TradeResponse) error) (func() error, error) {
	param := PublicV2TradeParam{
		Topic: PublicV2TopicTrade,
		Event: PublicV2EventSubscribe,
		Params: PublicV2TradeParamChild{
			Binary: false,
			Symbol: symbol,
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
		param.Event = PublicV2EventUnsubscribe
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
func (s *PublicV2Service) Start(ctx context.Context) {
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
func (s *PublicV2Service) Run() error {
	_, message, err := s.connection.ReadMessage()
	if err != nil {
		return err
	}

	topic, err := s.judgeTopic(message)
	if err != nil {
		return err
	}
	switch topic {
	case PublicV2TopicTrade:
		var resp PublicV2TradeResponse
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
func (s *PublicV2Service) Ping() error {
	if err := s.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
		return err
	}
	return nil
}

// Close :
func (s *PublicV2Service) Close() error {
	if err := s.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		return err
	}
	return nil
}
