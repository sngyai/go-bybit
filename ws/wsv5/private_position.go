package wsv5

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit"
)

// SubscribePosition :
func (s *PrivateService) SubscribePosition(
	f func(PrivatePositionResponse) error,
) (func() error, error) {
	key := PrivateParamKey{
		Topic: PrivateTopicPosition,
	}
	if err := s.addParamPositionFunc(key, f); err != nil {
		return nil, err
	}
	param := struct {
		Op   string        `json:"op"`
		Args []interface{} `json:"args"`
	}{
		Op:   "subscribe",
		Args: []interface{}{PrivateTopicPosition},
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
			Args: []interface{}{PrivateTopicPosition},
		}
		buf, err := json.Marshal(param)
		if err != nil {
			return err
		}
		if err := s.connection.WriteMessage(websocket.TextMessage, []byte(buf)); err != nil {
			return err
		}
		s.removeParamPositionFunc(key)
		return nil
	}, nil
}

// PrivatePositionResponse :
type PrivatePositionResponse struct {
	ID           string                `json:"id"`
	Topic        PrivateTopic          `json:"topic"`
	CreationTime int64                 `json:"creationTime"`
	Data         []PrivatePositionData `json:"data"`
}

// PrivatePositionData :
type PrivatePositionData struct {
	AutoAddMargin   int              `json:"autoAddMargin"`
	PositionIdx     int              `json:"positionIdx"`
	TpSlMode        bybit.TpSlMode   `json:"tpSlMode"`
	TradeMode       int              `json:"tradeMode"`
	RiskID          int              `json:"riskId"`
	RiskLimitValue  string           `json:"riskLimitValue"`
	Symbol          bybit.SymbolV5   `json:"symbol"`
	Side            bybit.Side       `json:"side"`
	Size            string           `json:"size"`
	EntryPrice      string           `json:"entryPrice"`
	Leverage        string           `json:"leverage"`
	PositionValue   string           `json:"positionValue"`
	MarkPrice       string           `json:"markPrice"`
	PositionBalance string           `json:"positionBalance"`
	PositionIM      string           `json:"positionIM"`
	PositionMM      string           `json:"positionMM"`
	TakeProfit      string           `json:"takeProfit"`
	StopLoss        string           `json:"stopLoss"`
	TrailingStop    string           `json:"trailingStop"`
	UnrealisedPnl   string           `json:"unrealisedPnl"`
	CumRealisedPnl  string           `json:"cumRealisedPnl"`
	CreatedTime     string           `json:"CreatedTime"`
	UpdatedTime     string           `json:"updatedTime"`
	TpslMode        bybit.TpSlMode   `json:"tpslMode"`
	LiqPrice        string           `json:"liqPrice"`
	BustPrice       string           `json:"bustPrice"`
	Category        bybit.CategoryV5 `json:"category"`
	PositionStatus  string           `json:"positionStatus"`
}

// Key :
func (r *PrivatePositionResponse) Key() PrivateParamKey {
	return PrivateParamKey{
		Topic: r.Topic,
	}
}

// addParamPositionFunc :
func (s *PrivateService) addParamPositionFunc(param PrivateParamKey, f func(PrivatePositionResponse) error) error {
	if _, exist := s.paramPositionMap[param]; exist {
		return errors.New("already registered for this param")
	}
	s.paramPositionMap[param] = f
	return nil
}

// removeParamPositionFunc :
func (s *PrivateService) removeParamPositionFunc(key PrivateParamKey) {
	delete(s.paramPositionMap, key)
}

// retrievePositionFunc :
func (s *PrivateService) retrievePositionFunc(key PrivateParamKey) (func(PrivatePositionResponse) error, error) {
	f, exist := s.paramPositionMap[key]
	if !exist {
		return nil, errors.New("func not found")
	}
	return f, nil
}
