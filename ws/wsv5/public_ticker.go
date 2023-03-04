package wsv5

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit"
)

// SubscribeTickers :
func (s *PublicService) SubscribeTickers(
	key PublicTickersParamKey,
	f func(PublicTickersResponse) error,
) (func() error, error) {
	if err := s.addParamTickersFunc(key, f); err != nil {
		return nil, err
	}
	param := struct {
		Op   string        `json:"op"`
		Args []interface{} `json:"args"`
	}{
		Op:   "subscribe",
		Args: []interface{}{key.Topic()},
	}
	buf, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	if err := s.connection.WriteMessage(websocket.TextMessage, buf); err != nil {
		return nil, err
	}
	return func() error {
		param := struct {
			Op   string        `json:"op"`
			Args []interface{} `json:"args"`
		}{
			Op:   "unsubscribe",
			Args: []interface{}{key.Topic()},
		}
		buf, err := json.Marshal(param)
		if err != nil {
			return err
		}
		if err := s.connection.WriteMessage(websocket.TextMessage, buf); err != nil {
			return err
		}
		s.removeParamTickersFunc(key)
		return nil
	}, nil
}

// PublicTickersParamKey :
type PublicTickersParamKey struct {
	Symbol bybit.SymbolV5
}

// Topic :
func (k *PublicTickersParamKey) Topic() string {
	return fmt.Sprintf("tickers.%s", k.Symbol)
}

// PublicTickersResponse :
type PublicTickersResponse struct {
	Topic     string            `json:"topic"`
	Type      string            `json:"type"`
	TimeStamp int64             `json:"ts"`
	Data      PublicTickersData `json:"data"`
}

// PublicTickersData :
type PublicTickersData struct {
	Symbol        string `json:"symbol"`
	LastPrice     string `json:"lastPrice"`
	HighPrice24H  string `json:"highPrice24h"`
	LowPrice24H   string `json:"lowPrice24h"`
	PrevPrice24H  string `json:"prevPrice24h"`
	Volume24H     string `json:"volume24h"`
	Turnover24H   string `json:"turnover24h"`
	Price24HPcnt  string `json:"price24hPcnt"`
	UsdIndexPrice string `json:"usdIndexPrice"`
}

// Key :
func (r *PublicTickersResponse) Key() PublicTickersParamKey {
	topic := r.Topic
	arr := strings.Split(topic, ".")
	if arr[0] != "tickers" || len(arr) != 2 {
		return PublicTickersParamKey{}
	}
	symbol := bybit.SymbolV5(arr[1])
	return PublicTickersParamKey{
		Symbol: symbol,
	}
}

// addParamTickersFunc :
func (s *PublicService) addParamTickersFunc(param PublicTickersParamKey, f func(PublicTickersResponse) error) error {
	if _, exist := s.paramTickersMap[param]; exist {
		return errors.New("already registered for this param")
	}
	s.paramTickersMap[param] = f
	return nil
}

// removeParamTradeFunc :
func (s *PublicService) removeParamTickersFunc(key PublicTickersParamKey) {
	delete(s.paramTickersMap, key)
}

// retrievePositionFunc :
func (s *PublicService) retrieveTickersFunc(key PublicTickersParamKey) (func(PublicTickersResponse) error, error) {
	f, exist := s.paramTickersMap[key]
	if !exist {
		return nil, errors.New("tickers func not found")
	}
	return f, nil
}
