package wsv5

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit"
)

// SubscribeOrderBook :
func (s *PublicService) SubscribeOrderBook(
	key PublicOrderBookParamKey,
	f func(PublicOrderBookResponse) error,
) (func() error, error) {
	if err := s.addParamOrderBookFunc(key, f); err != nil {
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
		s.removeParamOrderBookFunc(key)
		return nil
	}, nil
}

// PublicOrderBookParamKey :
type PublicOrderBookParamKey struct {
	Depth  int
	Symbol bybit.SymbolV5
}

// Topic :
func (k *PublicOrderBookParamKey) Topic() string {
	return fmt.Sprintf("orderbook.%d.%s", k.Depth, k.Symbol)
}

// PublicOrderBookResponse :
type PublicOrderBookResponse struct {
	Topic     string              `json:"topic"`
	Type      string              `json:"type"`
	TimeStamp int64               `json:"ts"`
	Data      PublicOrderBookData `json:"data"`
}

// PublicOrderBookData :
type PublicOrderBookData struct {
	Symbol   bybit.SymbolV5      `json:"s"`
	Bids     PublicOrderBookBids `json:"b"`
	Asks     PublicOrderBookAsks `json:"a"`
	UpdateID int                 `json:"u"`
	Seq      int                 `json:"seq"`
}

// PublicOrderBookBids :
type PublicOrderBookBids []struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// UnmarshalJSON :
func (b *PublicOrderBookBids) UnmarshalJSON(data []byte) error {
	parsedData := [][]string{}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return err
	}
	items := make(PublicOrderBookBids, len(parsedData))
	for i, item := range parsedData {
		item := item
		if len(item) != 2 {
			return errors.New("so far len(item) must be 2, please check it on documents")
		}
		items[i].Price = item[0]
		items[i].Size = item[1]
	}
	*b = items
	return nil
}

// PublicOrderBookAsks :
type PublicOrderBookAsks []struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// UnmarshalJSON :
func (b *PublicOrderBookAsks) UnmarshalJSON(data []byte) error {
	parsedData := [][]string{}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return err
	}
	items := make(PublicOrderBookAsks, len(parsedData))
	for i, item := range parsedData {
		item := item
		if len(item) != 2 {
			return errors.New("so far len(item) must be 2, please check it on documents")
		}
		items[i].Price = item[0]
		items[i].Size = item[1]
	}
	*b = items
	return nil
}

// Key :
func (r *PublicOrderBookResponse) Key() PublicOrderBookParamKey {
	topic := r.Topic
	arr := strings.Split(topic, ".")
	if arr[0] != "orderbook" || len(arr) != 3 {
		return PublicOrderBookParamKey{}
	}
	depth, err := strconv.Atoi(arr[1])
	if err != nil {
		return PublicOrderBookParamKey{}
	}
	symbol := bybit.SymbolV5(arr[2])
	return PublicOrderBookParamKey{
		Depth:  depth,
		Symbol: symbol,
	}
}

// addParamOrderBookFunc :
func (s *PublicService) addParamOrderBookFunc(param PublicOrderBookParamKey, f func(PublicOrderBookResponse) error) error {
	if _, exist := s.paramOrderBookMap[param]; exist {
		return errors.New("already registered for this param")
	}
	s.paramOrderBookMap[param] = f
	return nil
}

// removeParamTradeFunc :
func (s *PublicService) removeParamOrderBookFunc(key PublicOrderBookParamKey) {
	delete(s.paramOrderBookMap, key)
}

// retrievePositionFunc :
func (s *PublicService) retrieveOrderBookFunc(key PublicOrderBookParamKey) (func(PublicOrderBookResponse) error, error) {
	f, exist := s.paramOrderBookMap[key]
	if !exist {
		return nil, errors.New("func not found")
	}
	return f, nil
}
