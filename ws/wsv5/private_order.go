package wsv5

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
)

// SubscribeOrder :
func (s *PrivateService) SubscribeOrder(
	f func(PrivateOrderResponse) error,
) (func() error, error) {
	key := PrivateParamKey{
		Topic: PrivateTopicOrder,
	}
	if err := s.addParamOrderFunc(key, f); err != nil {
		return nil, err
	}
	param := struct {
		Op   string        `json:"op"`
		Args []interface{} `json:"args"`
	}{
		Op:   "subscribe",
		Args: []interface{}{PrivateTopicOrder},
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
			Args: []interface{}{PrivateTopicOrder},
		}
		buf, err := json.Marshal(param)
		if err != nil {
			return err
		}
		if err := s.connection.WriteMessage(websocket.TextMessage, []byte(buf)); err != nil {
			return err
		}
		s.removeParamOrderFunc(key)
		return nil
	}, nil
}

// PrivateOrderResponse :
type PrivateOrderResponse struct {
	ID           string             `json:"id"`
	Topic        PrivateTopic       `json:"topic"`
	CreationTime int64              `json:"creationTime"`
	Data         []PrivateOrderData `json:"data"`
}

// PrivateOrderData :
type PrivateOrderData struct {
	AvgPrice           string `json:"avgPrice"`
	BlockTradeID       string `json:"blockTradeId"`
	CancelType         string `json:"cancelType"`
	Category           string `json:"category"`
	CloseOnTrigger     bool   `json:"closeOnTrigger"`
	CreatedTime        string `json:"createdTime"`
	CumExecFee         string `json:"cumExecFee"`
	CumExecQty         string `json:"cumExecQty"`
	CumExecValue       string `json:"cumExecValue"`
	LeavesQty          string `json:"leavesQty"`
	LeavesValue        string `json:"leavesValue"`
	OrderID            string `json:"orderId"`
	OrderIv            string `json:"orderIv"`
	IsLeverage         string `json:"isLeverage"`
	LastPriceOnCreated string `json:"lastPriceOnCreated"`
	OrderStatus        string `json:"orderStatus"`
	OrderLinkID        string `json:"orderLinkId"`
	OrderType          string `json:"orderType"`
	PositionIdx        int    `json:"positionIdx"`
	Price              string `json:"price"`
	Qty                string `json:"qty"`
	ReduceOnly         bool   `json:"reduceOnly"`
	RejectReason       string `json:"rejectReason"`
	Side               string `json:"side"`
	SlTriggerBy        string `json:"slTriggerBy"`
	StopLoss           string `json:"stopLoss"`
	StopOrderType      string `json:"stopOrderType"`
	Symbol             string `json:"symbol"`
	TakeProfit         string `json:"takeProfit"`
	TimeInForce        string `json:"timeInForce"`
	TpTriggerBy        string `json:"tpTriggerBy"`
	TriggerBy          string `json:"triggerBy"`
	TriggerDirection   int    `json:"triggerDirection"`
	TriggerPrice       string `json:"triggerPrice"`
	UpdatedTime        string `json:"updatedTime"`
}

// Key :
func (r *PrivateOrderResponse) Key() PrivateParamKey {
	return PrivateParamKey{
		Topic: r.Topic,
	}
}

// addParamOrderFunc :
func (s *PrivateService) addParamOrderFunc(param PrivateParamKey, f func(PrivateOrderResponse) error) error {
	if _, exist := s.paramOrderMap[param]; exist {
		return errors.New("already registered for this param")
	}
	s.paramOrderMap[param] = f
	return nil
}

// removeParamOrderFunc :
func (s *PrivateService) removeParamOrderFunc(key PrivateParamKey) {
	delete(s.paramOrderMap, key)
}

// retrieveOrderFunc :
func (s *PrivateService) retrieveOrderFunc(key PrivateParamKey) (func(PrivateOrderResponse) error, error) {
	f, exist := s.paramOrderMap[key]
	if !exist {
		return nil, errors.New("func not found")
	}
	return f, nil
}
