package rest

import (
	"encoding/json"
	"fmt"

	"github.com/sngyai/go-bybit"
)

// V5OrderServiceI :
type V5OrderServiceI interface {
	CreateOrder(V5CreateOrderParam) (*V5CreateOrderResponse, error)
}

// V5OrderService :
type V5OrderService struct {
	client *Client
}

// V5CreateOrderParam :
type V5CreateOrderParam struct {
	Category  bybit.CategoryV5 `json:"category"`
	Symbol    bybit.SymbolV5   `json:"symbol"`
	Side      bybit.Side       `json:"side"`
	OrderType bybit.OrderType  `json:"orderType"`
	Qty       string           `json:"qty"`

	IsLeverage            *bybit.IsLeverage       `json:"isLeverage,omitempty"`
	Price                 *string                 `json:"price,omitempty"`
	TriggerDirection      *bybit.TriggerDirection `json:"triggerDirection,omitempty"`
	OrderFilter           *bybit.OrderFilter      `json:"orderFilter,omitempty"` // If not passed, Order by default
	TriggerPrice          *string                 `json:"triggerPrice,omitempty"`
	TriggerBy             *bybit.TriggerBy        `json:"triggerBy,omitempty"`
	OrderIv               *string                 `json:"orderIv,omitempty"`     // option only.
	TimeInForce           *bybit.TimeInForce      `json:"timeInForce,omitempty"` // If not passed, GTC is used by default
	PositionIdx           *bybit.PositionIdx      `json:"positionIdx,omitempty"` // Under hedge-mode, this param is required
	OrderLinkID           *string                 `json:"orderLinkId,omitempty"`
	TakeProfit            *string                 `json:"takeProfit,omitempty"`
	StopLoss              *string                 `json:"stopLoss,omitempty"`
	TpTriggerBy           *bybit.TriggerBy        `json:"tpTriggerBy,omitempty"`
	SlTriggerBy           *bybit.TriggerBy        `json:"slTriggerBy,omitempty"`
	ReduceOnly            *bool                   `json:"reduce_only,omitempty"`
	CloseOnTrigger        *bool                   `json:"closeOnTrigger,omitempty"`
	MarketMakerProtection *bool                   `json:"mmp,omitempty"` // option only
}

// V5CreateOrderResponse :
type V5CreateOrderResponse struct {
	CommonV5Response `json:",inline"`
	Result           V5CreateOrderResult `json:"result"`
}

// V5CreateOrderResult :
type V5CreateOrderResult struct {
	OrderID     string `json:"orderId"`
	OrderLinkID string `json:"orderLinkId"`
}

// CreateOrder :
func (s *V5OrderService) CreateOrder(param V5CreateOrderParam) (*V5CreateOrderResponse, error) {
	var res V5CreateOrderResponse

	body, err := json.Marshal(param)
	if err != nil {
		return &res, fmt.Errorf("json marshal: %w", err)
	}

	if err := s.client.postV5JSON("/v5/order/create", body, &res); err != nil {
		return &res, err
	}

	return &res, nil
}
